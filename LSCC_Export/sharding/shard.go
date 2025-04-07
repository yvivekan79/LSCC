package sharding

import (
        "sync"

        "lscc/config"
        "lscc/core"
        "lscc/utils"
)

// Shard represents a blockchain shard
type Shard struct {
        ID             int
        Layer          int
        Blockchain     *core.Blockchain
        Nodes          []string // Node IDs in this shard
        RelayNodes     []string // Relay nodes connecting to other shards
        mu             sync.RWMutex
        config         *config.Config
        logger         *utils.Logger
        crossShardTxs  map[string]*core.Transaction // Transactions that cross shards
        pendingBlocks  map[string]*core.Block       // Blocks pending cross-shard validation
}

// NewShard creates a new shard
func NewShard(id int, layer int, cfg *config.Config) *Shard {
        logger := utils.GetLogger()
        
        // Create a config copy with the shard ID
        shardConfig := *cfg
        shardConfig.ShardID = id
        
        return &Shard{
                ID:             id,
                Layer:          layer,
                Blockchain:     core.NewBlockchain(&shardConfig),
                Nodes:          make([]string, 0),
                RelayNodes:     make([]string, 0),
                config:         &shardConfig,
                logger:         logger,
                crossShardTxs:  make(map[string]*core.Transaction),
                pendingBlocks:  make(map[string]*core.Block),
        }
}

// AddNode adds a node to the shard
func (s *Shard) AddNode(nodeID string, isRelay bool) {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        // Check if node already exists
        for _, id := range s.Nodes {
                if id == nodeID {
                        return
                }
        }
        
        s.Nodes = append(s.Nodes, nodeID)
        if isRelay {
                s.RelayNodes = append(s.RelayNodes, nodeID)
        }
        
        s.logger.Info("Node added to shard", 
                "shardID", s.ID, 
                "layer", s.Layer, 
                "nodeID", nodeID, 
                "isRelay", isRelay)
}

// RemoveNode removes a node from the shard
func (s *Shard) RemoveNode(nodeID string) {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        // Remove from nodes list
        for i, id := range s.Nodes {
                if id == nodeID {
                        s.Nodes = append(s.Nodes[:i], s.Nodes[i+1:]...)
                        break
                }
        }
        
        // Remove from relay nodes list if present
        for i, id := range s.RelayNodes {
                if id == nodeID {
                        s.RelayNodes = append(s.RelayNodes[:i], s.RelayNodes[i+1:]...)
                        break
                }
        }
        
        s.logger.Info("Node removed from shard", "shardID", s.ID, "nodeID", nodeID)
}

// GetNodeCount returns the number of nodes in the shard
func (s *Shard) GetNodeCount() int {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return len(s.Nodes)
}

// IsNodeInShard checks if a node is in this shard
func (s *Shard) IsNodeInShard(nodeID string) bool {
        s.mu.RLock()
        defer s.mu.RUnlock()
        
        for _, id := range s.Nodes {
                if id == nodeID {
                        return true
                }
        }
        return false
}

// IsRelayNode checks if a node is a relay node in this shard
func (s *Shard) IsRelayNode(nodeID string) bool {
        s.mu.RLock()
        defer s.mu.RUnlock()
        
        for _, id := range s.RelayNodes {
                if id == nodeID {
                        return true
                }
        }
        return false
}

// AddCrossShardTransaction adds a transaction that crosses shard boundaries
func (s *Shard) AddCrossShardTransaction(tx *core.Transaction) {
        s.mu.Lock()
        defer s.mu.Unlock()
        
        s.crossShardTxs[tx.Hash] = tx
        s.logger.Info("Cross-shard transaction added", 
                "txHash", tx.Hash, 
                "sourceShard", tx.SourceShard, 
                "targetShard", tx.TargetShard)
}

// GetCrossShardTransactions returns all cross-shard transactions
func (s *Shard) GetCrossShardTransactions() []*core.Transaction {
        s.mu.RLock()
        defer s.mu.RUnlock()
        
        txs := make([]*core.Transaction, 0, len(s.crossShardTxs))
        for _, tx := range s.crossShardTxs {
                txs = append(txs, tx)
        }
        return txs
}

// ProcessCrossShardBlock processes a block from another shard
func (s *Shard) ProcessCrossShardBlock(block *core.Block, sourceShard int) error {
        // Handle cross-shard transactions that target this shard
        for _, tx := range block.Transactions {
                if tx.TargetShard == s.ID && tx.SourceShard == sourceShard {
                        // Process the incoming transaction
                        localTx, err := core.NewTransaction(
                                tx.From,
                                tx.To,
                                tx.Amount,
                                tx.Fee,
                                s.ID, // Now it's in this shard
                                s.ID, // Target is also this shard
                                s.Layer,
                                tx.Type,
                        )
                        if err != nil {
                                s.logger.Error("Failed to create local transaction from cross-shard tx", 
                                        "error", err, 
                                        "txHash", tx.Hash)
                                continue
                        }
                        
                        // Add to local blockchain
                        err = s.Blockchain.AddTransaction(localTx)
                        if err != nil {
                                s.logger.Error("Failed to add cross-shard transaction to local chain", 
                                        "error", err, 
                                        "txHash", localTx.Hash)
                                continue
                        }
                        
                        s.logger.Info("Processed cross-shard transaction", 
                                "txHash", tx.Hash, 
                                "sourceShard", sourceShard, 
                                "targetShard", s.ID)
                }
        }
        
        // Add cross-shard reference to a local block if needed
        latestBlock := s.Blockchain.GetLatestBlock()
        if latestBlock != nil {
                blockHash, _ := block.Hash()
                latestBlock.AddCrossReference(sourceShard, blockHash, block.Header.Height)
                s.logger.Info("Added cross-shard reference", 
                        "localBlock", latestBlock.Header.Height, 
                        "remoteBlock", block.Header.Height, 
                        "remoteShard", sourceShard)
        }
        
        return nil
}

// GetStatus returns the current status of the shard
func (s *Shard) GetStatus() map[string]interface{} {
        s.mu.RLock()
        defer s.mu.RUnlock()
        
        return map[string]interface{}{
                "id":                  s.ID,
                "layer":               s.Layer,
                "node_count":          len(s.Nodes),
                "relay_node_count":    len(s.RelayNodes),
                "blockchain_height":   s.Blockchain.GetHeight(),
                "cross_shard_tx_count": len(s.crossShardTxs),
                "pending_blocks":      len(s.pendingBlocks),
        }
}
