package network

import (
        "encoding/json"
        "fmt"
        "net"
        "sync"
        "time"

        "lscc/core"
        "lscc/utils"
)

// Peer represents a connection to another node
type Peer struct {
        ID             string
        Address        string
        conn           net.Conn
        node           *Node
        sendMu         sync.Mutex
        isConnected    bool
        disconnectOnce sync.Once
        lastSeen       time.Time
        logger         *utils.Logger
}

// NewPeer creates a new peer connection
func NewPeer(id string, address string, conn net.Conn, node *Node) *Peer {
        return &Peer{
                ID:          id,
                Address:     address,
                conn:        conn,
                node:        node,
                isConnected: true,
                lastSeen:    time.Now(),
                logger:      utils.GetLogger(),
        }
}

// Start begins listening for messages from the peer
func (p *Peer) Start() {
        go p.receiveLoop()
}

// receiveLoop continuously reads messages from the peer
func (p *Peer) receiveLoop() {
        defer p.Disconnect()

        decoder := json.NewDecoder(p.conn)
        for {
                var msg Message
                err := decoder.Decode(&msg)
                if err != nil {
                        p.logger.Debug("Error reading from peer", "peerID", p.ID, "error", err)
                        return
                }

                // Update last seen time
                p.lastSeen = time.Now()

                // Handle message
                err = p.handleMessage(msg)
                if err != nil {
                        p.logger.Error("Error handling message", "peerID", p.ID, "error", err)
                        continue
                }
        }
}

// handleMessage processes a received message
func (p *Peer) handleMessage(msg Message) error {
        switch msg.Type {
        case MessageTypeHandshake:
                return p.handleHandshake(msg.Data)
        case MessageTypePeerListRequest:
                return p.handlePeerListRequest()
        case MessageTypePeerList:
                return p.handlePeerList(msg.Data)
        case MessageTypeTransaction:
                return p.handleTransaction(msg.Data)
        case MessageTypeBlock:
                return p.handleBlock(msg.Data)
        case MessageTypeBlockRequest:
                return p.handleBlockRequest(msg.Data)
        case MessageTypeBlockResponse:
                return p.handleBlockResponse(msg.Data)
        default:
                return fmt.Errorf("unknown message type: %d", msg.Type)
        }
}

// handleHandshake processes a handshake message
func (p *Peer) handleHandshake(data json.RawMessage) error {
        var handshake HandshakeMessage
        err := json.Unmarshal(data, &handshake)
        if err != nil {
                return err
        }

        // Update peer ID if not set
        if p.ID == "" {
                p.ID = handshake.NodeID
                p.node.mu.Lock()
                p.node.Peers[p.ID] = p
                p.node.mu.Unlock()
        }

        p.logger.Info("Received handshake from peer", 
                "peerID", p.ID, 
                "version", handshake.Version,
                "shardID", handshake.ShardID)

        return nil
}

// handlePeerListRequest responds to a peer list request
func (p *Peer) handlePeerListRequest() error {
        peerList := p.node.GetPeerList()
        return p.SendMessage(MessageTypePeerList, peerList)
}

// handlePeerList processes a received peer list
func (p *Peer) handlePeerList(data json.RawMessage) error {
        var peerAddresses []string
        err := json.Unmarshal(data, &peerAddresses)
        if err != nil {
                return err
        }

        p.node.HandlePeerList(peerAddresses)
        return nil
}

// handleTransaction processes a received transaction
func (p *Peer) handleTransaction(data json.RawMessage) error {
        var tx core.Transaction
        err := json.Unmarshal(data, &tx)
        if err != nil {
                return err
        }

        // Check if transaction is cross-shard
        if tx.IsCrossShard() {
                // Process using shard manager
                err = p.node.ShardManager.ProcessCrossShardTransaction(&tx)
        } else {
                // Add to local blockchain
                err = p.node.Blockchain.AddTransaction(&tx)
        }

        if err != nil {
                return err
        }

        // Relay to other peers
        p.node.BroadcastTransaction(&tx)
        return nil
}

// handleBlock processes a received block
func (p *Peer) handleBlock(data json.RawMessage) error {
        var block core.Block
        err := json.Unmarshal(data, &block)
        if err != nil {
                return err
        }

        // Validate and process block
        if !p.node.Consensus.ValidateBlock(&block) {
                return fmt.Errorf("invalid block")
        }

        err = p.node.Consensus.ProcessBlock(&block)
        if err != nil {
                return err
        }

        // If this is a relay node and the block is from a different shard,
        // propagate to appropriate shards
        if p.node.Config.IsRelay && block.ShardID != p.node.Config.ShardID {
                // Determine target shards and propagate
                // This is simplified; a real implementation would be more complex
                blockHash, _ := block.Hash()
                p.logger.Info("Relay node propagating block", 
                        "blockHash", blockHash, 
                        "fromShard", block.ShardID)
        }

        return nil
}

// handleBlockRequest processes a block request
func (p *Peer) handleBlockRequest(data json.RawMessage) error {
        var request BlockRequestMessage
        err := json.Unmarshal(data, &request)
        if err != nil {
                return err
        }

        var block *core.Block
        if request.Hash != "" {
                block = p.node.Blockchain.GetBlockByHash(request.Hash)
        } else if request.Height > 0 {
                block = p.node.Blockchain.GetBlockByHeight(request.Height)
        }

        if block == nil {
                return fmt.Errorf("block not found")
        }

        response := BlockResponseMessage{
                RequestID: request.RequestID,
                Block:     *block,
        }

        return p.SendMessage(MessageTypeBlockResponse, response)
}

// handleBlockResponse processes a block response
func (p *Peer) handleBlockResponse(data json.RawMessage) error {
        var response BlockResponseMessage
        err := json.Unmarshal(data, &response)
        if err != nil {
                return err
        }

        // Process the block
        if !p.node.Consensus.ValidateBlock(&response.Block) {
                return fmt.Errorf("invalid block in response")
        }

        err = p.node.Consensus.ProcessBlock(&response.Block)
        if err != nil {
                return err
        }

        return nil
}

// SendMessage sends a message to the peer
func (p *Peer) SendMessage(msgType MessageType, data interface{}) error {
        p.sendMu.Lock()
        defer p.sendMu.Unlock()

        if !p.isConnected {
                return fmt.Errorf("peer disconnected")
        }

        msg := Message{
                Type:      msgType,
                Timestamp: time.Now().Unix(),
        }

        if data != nil {
                dataBytes, err := json.Marshal(data)
                if err != nil {
                        return err
                }
                msg.Data = dataBytes
        }

        msgBytes, err := json.Marshal(msg)
        if err != nil {
                return err
        }

        _, err = p.conn.Write(msgBytes)
        if err != nil {
                p.Disconnect()
                return err
        }

        return nil
}

// Disconnect closes the connection to the peer
func (p *Peer) Disconnect() {
        p.disconnectOnce.Do(func() {
                p.isConnected = false
                if p.conn != nil {
                        p.conn.Close()
                }
                p.logger.Info("Peer disconnected", "peerID", p.ID, "address", p.Address)
        })
}

// IsConnected returns whether the peer is still connected
func (p *Peer) IsConnected() bool {
        return p.isConnected
}

// String returns a string representation of the peer
func (p *Peer) String() string {
        return fmt.Sprintf("Peer{ID: %s, Address: %s, Connected: %v, LastSeen: %v}",
                p.ID, p.Address, p.isConnected, p.lastSeen)
}
