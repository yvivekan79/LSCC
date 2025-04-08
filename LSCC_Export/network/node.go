package network

import (
        "context"
        "encoding/json"
        "fmt"
        "net"
        "strconv"
        "sync"
        "time"

        "lscc/config"
        "lscc/consensus"
        "lscc/core"
        "lscc/sharding"
        "lscc/utils"
)

// Node represents a network node in the LSCC blockchain
type Node struct {
        ID            string
        IP            string
        Port          int
        Peers         map[string]*Peer
        Blockchain    *core.Blockchain
        ShardManager  *sharding.Manager
        Consensus     consensus.ConsensusEngine
        Config        *config.Config
        listener      net.Listener
        ctx           context.Context
        cancel        context.CancelFunc
        mu            sync.RWMutex
        logger        *utils.Logger
        isRunning     bool
}

// NewNode creates a new network node
func NewNode(cfg *config.Config, shardManager *sharding.Manager) (*Node, error) {
        logger := utils.GetLogger()
        
        // Get the shard this node belongs to
        shardID, err := shardManager.GetNodeShard(cfg.NodeID)
        if err != nil {
                return nil, err
        }
        
        // Get the shard
        shard, err := shardManager.GetShard(shardID)
        if err != nil {
                return nil, err
        }
        
        // Create blockchain from shard
        blockchain := shard.Blockchain
        
        // Create node
        ctx, cancel := context.WithCancel(context.Background())
        node := &Node{
                ID:           cfg.NodeID,
                Port:         cfg.Port,
                Peers:        make(map[string]*Peer),
                Blockchain:   blockchain,
                ShardManager: shardManager,
                Config:       cfg,
                ctx:          ctx,
                cancel:       cancel,
                logger:       logger,
        }
        
        // Create consensus engine
        consensusEngine, err := consensus.NewConsensusEngine(cfg, blockchain)
        if err != nil {
                logger.Error("Failed to create consensus engine", "error", err)
                return nil, err
        }
        node.Consensus = consensusEngine
        
        logger.Info("Node created", 
                "nodeID", node.ID, 
                "shardID", shardID, 
                "port", node.Port)
        
        return node, nil
}

// Start starts the node
func (n *Node) Start() error {
        n.mu.Lock()
        if n.isRunning {
                n.mu.Unlock()
                return fmt.Errorf("node already running")
        }
        n.isRunning = true
        n.mu.Unlock()
        
        // Start listening for incoming connections
        addr := fmt.Sprintf("0.0.0.0:%d", n.Port)
        listener, err := net.Listen("tcp", addr)
        if err != nil {
                n.logger.Error("Failed to start node listener", "error", err)
                return err
        }
        n.listener = listener
        
        // Start consensus engine
        err = n.Consensus.Start()
        if err != nil {
                n.logger.Error("Failed to start consensus engine", "error", err)
                return err
        }
        
        // Connect to bootstrap nodes
        for _, bootstrapAddr := range n.Config.BootstrapNodes {
                go n.connectToPeer(bootstrapAddr)
        }
        
        // Start accepting connections
        go n.acceptConnections()
        
        // Start peer discovery and maintenance
        go n.maintainPeers()
        
        n.logger.Info("Node started", "address", addr, "nodeID", n.ID)
        return nil
}

// Stop stops the node
func (n *Node) Stop() error {
        n.mu.Lock()
        defer n.mu.Unlock()
        
        if !n.isRunning {
                return fmt.Errorf("node not running")
        }
        
        // Cancel context to stop all goroutines
        n.cancel()
        
        // Close listener
        if n.listener != nil {
                n.listener.Close()
        }
        
        // Stop consensus engine
        if n.Consensus != nil {
                n.Consensus.Stop()
        }
        
        // Close all peer connections
        for _, peer := range n.Peers {
                peer.Disconnect()
        }
        
        n.isRunning = false
        n.logger.Info("Node stopped", "nodeID", n.ID)
        
        return nil
}

// acceptConnections accepts incoming connections
func (n *Node) acceptConnections() {
        for {
                conn, err := n.listener.Accept()
                if err != nil {
                        select {
                        case <-n.ctx.Done():
                                return // Context cancelled, exit gracefully
                        default:
                                n.logger.Error("Error accepting connection", "error", err)
                                continue
                        }
                }
                
                // Handle connection in a new goroutine
                go n.handleConnection(conn)
        }
}

// handleConnection handles a new incoming connection
func (n *Node) handleConnection(conn net.Conn) {
        // Set connection deadline
        conn.SetDeadline(time.Now().Add(time.Duration(n.Config.ConnectionTimeout) * time.Second))
        
        // Read handshake message
        var handshake HandshakeMessage
        decoder := json.NewDecoder(conn)
        if err := decoder.Decode(&handshake); err != nil {
                n.logger.Error("Error decoding handshake", "error", err)
                conn.Close()
                return
        }
        
        // Create peer
        peer := NewPeer(
                handshake.NodeID,
                conn.RemoteAddr().String(),
                conn,
                n,
        )
        
        // Add peer
        n.mu.Lock()
        n.Peers[peer.ID] = peer
        n.mu.Unlock()
        
        n.logger.Info("New peer connected", "peerID", peer.ID, "address", peer.Address)
        
        // Start peer message handling
        peer.Start()
        
        // Send our own handshake
        n.sendHandshake(peer)
        
        // Request peer list
        n.requestPeerList(peer)
}

// connectToPeer attempts to connect to a peer at the given address
func (n *Node) connectToPeer(address string) {
        // Don't connect to self
        if address == fmt.Sprintf("%s:%d", n.IP, n.Port) {
                return
        }
        
        // Check if already connected
        n.mu.RLock()
        for _, peer := range n.Peers {
                if peer.Address == address {
                        n.mu.RUnlock()
                        return
                }
        }
        n.mu.RUnlock()
        
        // Connect to peer
        conn, err := net.DialTimeout("tcp", address, time.Duration(n.Config.ConnectionTimeout)*time.Second)
        if err != nil {
                n.logger.Error("Failed to connect to peer", "address", address, "error", err)
                return
        }
        
        // Create peer
        peer := NewPeer(
                "", // ID will be set after handshake
                address,
                conn,
                n,
        )
        
        // Start peer message handling
        peer.Start()
        
        // Send handshake
        n.sendHandshake(peer)
        
        n.logger.Info("Connected to peer", "address", address)
}

// sendHandshake sends a handshake message to a peer
func (n *Node) sendHandshake(peer *Peer) {
        handshake := HandshakeMessage{
                NodeID:       n.ID,
                Version:      "1.0.0",
                ShardID:      n.Config.ShardID,
                IsRelay:      n.Config.IsRelay,
                Port:         n.Port,
                Timestamp:    time.Now().Unix(),
        }
        
        peer.SendMessage(MessageTypeHandshake, handshake)
}

// requestPeerList requests the peer list from a peer
func (n *Node) requestPeerList(peer *Peer) {
        peer.SendMessage(MessageTypePeerListRequest, nil)
}

// broadcastMessage broadcasts a message to all peers
func (n *Node) broadcastMessage(messageType MessageType, data interface{}) {
        n.mu.RLock()
        defer n.mu.RUnlock()
        
        for _, peer := range n.Peers {
                peer.SendMessage(messageType, data)
        }
}

// broadcastTransaction broadcasts a transaction to all peers
func (n *Node) BroadcastTransaction(tx *core.Transaction) {
        n.broadcastMessage(MessageTypeTransaction, tx)
        n.logger.Info("Transaction broadcasted", "txHash", tx.Hash)
}

// broadcastBlock broadcasts a block to all peers
func (n *Node) BroadcastBlock(block *core.Block) {
        n.broadcastMessage(MessageTypeBlock, block)
        blockHash, _ := block.Hash()
        n.logger.Info("Block broadcasted", "blockHash", blockHash, "height", block.Header.Height)
}

// getPeerCount returns the number of connected peers
func (n *Node) GetPeerCount() int {
        n.mu.RLock()
        defer n.mu.RUnlock()
        return len(n.Peers)
}

// getPeers returns a list of all peers
func (n *Node) GetPeers() []*Peer {
        n.mu.RLock()
        defer n.mu.RUnlock()
        
        peers := make([]*Peer, 0, len(n.Peers))
        for _, peer := range n.Peers {
                peers = append(peers, peer)
        }
        
        return peers
}

// maintainPeers periodically checks peer connections and discovers new peers
func (n *Node) maintainPeers() {
        ticker := time.NewTicker(time.Duration(n.Config.SyncInterval) * time.Second)
        defer ticker.Stop()
        
        for {
                select {
                case <-n.ctx.Done():
                        return // Context cancelled, exit gracefully
                case <-ticker.C:
                        // Discover new peers if needed
                        if n.GetPeerCount() < n.Config.PeerLimit {
                                n.discoverPeers()
                        }
                        
                        // Remove disconnected peers
                        n.cleanupPeers()
                }
        }
}

// discoverPeers discovers new peers from existing peers
func (n *Node) discoverPeers() {
        peers := n.GetPeers()
        
        // Request peer list from a random subset of existing peers
        for _, peer := range peers {
                n.requestPeerList(peer)
        }
}

// cleanupPeers removes disconnected peers
func (n *Node) cleanupPeers() {
        n.mu.Lock()
        defer n.mu.Unlock()
        
        for id, peer := range n.Peers {
                if !peer.IsConnected() {
                        delete(n.Peers, id)
                        n.logger.Info("Removed disconnected peer", "peerID", id)
                }
        }
}

// handlePeerList processes a peer list received from a peer
func (n *Node) HandlePeerList(peerAddresses []string) {
        for _, addr := range peerAddresses {
                // Check if we need more peers
                if n.GetPeerCount() >= n.Config.PeerLimit {
                        break
                }
                
                // Connect to new peer
                go n.connectToPeer(addr)
        }
}

// GetStatus returns the current status of the node
func (n *Node) GetStatus() map[string]interface{} {
        n.mu.RLock()
        defer n.mu.RUnlock()
        
        status := map[string]interface{}{
                "node_id":        n.ID,
                "is_running":     n.isRunning,
                "peer_count":     len(n.Peers),
                "shard_id":       n.Config.ShardID,
                "is_relay":       n.Config.IsRelay,
                "blockchain_height": n.Blockchain.GetHeight(),
                "consensus_type": n.Consensus.GetType(),
        }
        
        return status
}

// String returns a string representation of the node
func (n *Node) String() string {
        return fmt.Sprintf("Node{ID: %s, Port: %d, PeerCount: %d, ShardID: %d, IsRelay: %v}",
                n.ID, n.Port, n.GetPeerCount(), n.Config.ShardID, n.Config.IsRelay)
}

// GetPeerList returns a list of peer addresses
func (n *Node) GetPeerList() []string {
        n.mu.RLock()
        defer n.mu.RUnlock()
        
        peerList := make([]string, 0, len(n.Peers))
        for _, peer := range n.Peers {
                host, portStr, err := net.SplitHostPort(peer.Address)
                if err != nil {
                        continue
                }
                
                port, err := strconv.Atoi(portStr)
                if err != nil {
                        continue
                }
                
                // Use peer's actual listening port, not the connection port
                peerAddress := fmt.Sprintf("%s:%d", host, port)
                peerList = append(peerList, peerAddress)
        }
        
        return peerList
}
