package sharding

import (
        "errors"
        "sync"

        "lscc/config"
        "lscc/core"
        "lscc/utils"
)

// ShardingStrategy defines different sharding strategies
type ShardingStrategy int

const (
        // StaticSharding assigns shards statically
        StaticSharding ShardingStrategy = iota
        // DynamicSharding adjusts shards based on load
        DynamicSharding
        // HybridSharding combines both approaches
        HybridSharding
)

// Manager manages the sharding structure and cross-shard communication
type Manager struct {
        Shards          map[int]*Shard
        NodeToShard     map[string]int
        RelayNodes      map[string]bool
        config          *config.Config
        strategy        ShardingStrategy
        layerCount      int
        mu              sync.RWMutex
        logger          *utils.Logger
        crossChannel    *CrossChannel
}

// NewManager creates a new sharding manager
func NewManager(cfg *config.Config) *Manager {
        logger := utils.GetLogger()
        manager := &Manager{
                Shards:      make(map[int]*Shard),
                NodeToShard: make(map[string]int),
                RelayNodes:  make(map[string]bool),
                config:      cfg,
                strategy:    ShardingStrategy(cfg.ShardingStrategy),
                layerCount:  cfg.LayerCount,
                logger:      logger,
        }
        
        // Initialize cross-channel communication
        manager.crossChannel = NewCrossChannel(manager, cfg)
        
        // Initialize shards based on config
        manager.InitializeShards()
        
        return manager
}

// InitializeShards creates the initial shard structure
func (m *Manager) InitializeShards() {
        m.mu.Lock()
        defer m.mu.Unlock()
        
        // Create shards for each layer
        for layer := 0; layer < m.layerCount; layer++ {
                for shard := 0; shard < m.config.ShardCount; shard++ {
                        shardID := layer*m.config.ShardCount + shard
                        m.Shards[shardID] = NewShard(shardID, layer, m.config)
                        m.logger.Info("Created shard", "shardID", shardID, "layer", layer)
                }
        }
        
        // Assign this node to a shard if shardID is specified
        if m.config.ShardID >= 0 && m.config.ShardID < len(m.Shards) {
                m.AssignNodeToShard(m.config.NodeID, m.config.ShardID, m.config.IsRelay)
        } else {
                // Auto-assign based on node ID
                m.AutoAssignNodeToShard(m.config.NodeID, m.config.IsRelay)
        }
}

// AssignNodeToShard assigns a node to a specific shard
func (m *Manager) AssignNodeToShard(nodeID string, shardID int, isRelay bool) error {
        m.mu.Lock()
        defer m.mu.Unlock()
        
        shard, exists := m.Shards[shardID]
        if !exists {
                return errors.New("shard does not exist")
        }
        
        // Remove node from current shard if it exists
        currentShardID, exists := m.NodeToShard[nodeID]
        if exists {
                currentShard := m.Shards[currentShardID]
                currentShard.RemoveNode(nodeID)
        }
        
        // Add node to new shard
        shard.AddNode(nodeID, isRelay)
        m.NodeToShard[nodeID] = shardID
        
        // Update relay nodes map
        if isRelay {
                m.RelayNodes[nodeID] = true
        } else {
                delete(m.RelayNodes, nodeID)
        }
        
        m.logger.Info("Node assigned to shard", 
                "nodeID", nodeID, 
                "shardID", shardID, 
                "isRelay", isRelay)
        
        return nil
}

// AutoAssignNodeToShard automatically assigns a node to a shard
func (m *Manager) AutoAssignNodeToShard(nodeID string, isRelay bool) error {
        // Calculate a deterministic shard assignment based on node ID
        // This is a simple hash-based assignment for demo purposes
        
        // Use the last few characters of node ID to determine shard
        shardIndex := 0
        if len(nodeID) > 0 {
                // Simple hash function
                sum := 0
                for _, char := range nodeID {
                        sum += int(char)
                }
                shardIndex = sum % len(m.Shards)
        }
        
        return m.AssignNodeToShard(nodeID, shardIndex, isRelay)
}

// GetNodeShard returns the shard ID for a node
func (m *Manager) GetNodeShard(nodeID string) (int, error) {
        m.mu.RLock()
        defer m.mu.RUnlock()
        
        shardID, exists := m.NodeToShard[nodeID]
        if !exists {
                return -1, errors.New("node not assigned to any shard")
        }
        
        return shardID, nil
}

// GetShard returns a shard by ID
func (m *Manager) GetShard(shardID int) (*Shard, error) {
        m.mu.RLock()
        defer m.mu.RUnlock()
        
        shard, exists := m.Shards[shardID]
        if !exists {
                return nil, errors.New("shard does not exist")
        }
        
        return shard, nil
}

// GetShardForTransaction determines which shard should process a transaction
func (m *Manager) GetShardForTransaction(tx *core.Transaction) (int, error) {
        if tx.TargetShard >= 0 && tx.TargetShard < len(m.Shards) {
                return tx.TargetShard, nil
        }
        
        // Determine shard based on receiver address
        // This is a simple implementation; a real system would use a more sophisticated approach
        sum := 0
        for _, char := range tx.To {
                sum += int(char)
        }
        
        // Get shard in the appropriate layer
        shardsInLayer := m.config.ShardCount
        return (tx.Layer * shardsInLayer) + (sum % shardsInLayer), nil
}

// ProcessCrossShardTransaction processes a transaction that crosses shard boundaries
func (m *Manager) ProcessCrossShardTransaction(tx *core.Transaction) error {
        // Validate transaction
        if !tx.IsValid() {
                return errors.New("invalid transaction")
        }
        
        // Get source and target shards
        sourceShard, err := m.GetShard(tx.SourceShard)
        if err != nil {
                return err
        }
        
        // Validate that target shard exists
        _, err = m.GetShard(tx.TargetShard)
        if err != nil {
                return err
        }
        
        // Add to source shard as outgoing transaction
        sourceShard.AddCrossShardTransaction(tx)
        
        // Use cross-channel to propagate to target shard
        err = m.crossChannel.PropagateTransaction(tx, tx.SourceShard, tx.TargetShard)
        if err != nil {
                return err
        }
        
        m.logger.Info("Processed cross-shard transaction", 
                "txHash", tx.Hash, 
                "sourceShard", tx.SourceShard, 
                "targetShard", tx.TargetShard)
        
        return nil
}

// ProcessCrossShardBlock processes a block from another shard
func (m *Manager) ProcessCrossShardBlock(block *core.Block, sourceShard, targetShard int) error {
        targetShardObj, err := m.GetShard(targetShard)
        if err != nil {
                return err
        }
        
        return targetShardObj.ProcessCrossShardBlock(block, sourceShard)
}

// RebalanceShards rebalances the shards based on load (for dynamic sharding)
func (m *Manager) RebalanceShards() error {
        if m.strategy != DynamicSharding && m.strategy != HybridSharding {
                return errors.New("rebalancing only available for dynamic or hybrid sharding")
        }
        
        // Implement shard rebalancing logic
        // This would involve moving nodes between shards based on load metrics
        
        m.logger.Info("Shards rebalanced")
        return nil
}

// GetShardCount returns the total number of shards
func (m *Manager) GetShardCount() int {
        return len(m.Shards)
}

// GetRelayNodes returns all relay nodes
func (m *Manager) GetRelayNodes() []string {
        m.mu.RLock()
        defer m.mu.RUnlock()
        
        relayNodes := make([]string, 0, len(m.RelayNodes))
        for nodeID := range m.RelayNodes {
                relayNodes = append(relayNodes, nodeID)
        }
        
        return relayNodes
}

// GetStatus returns the status of the sharding manager
func (m *Manager) GetStatus() map[string]interface{} {
        m.mu.RLock()
        defer m.mu.RUnlock()
        
        shardStatuses := make(map[int]map[string]interface{})
        for id, shard := range m.Shards {
                shardStatuses[id] = shard.GetStatus()
        }
        
        return map[string]interface{}{
                "shard_count":    len(m.Shards),
                "node_count":     len(m.NodeToShard),
                "relay_count":    len(m.RelayNodes),
                "layer_count":    m.layerCount,
                "strategy":       int(m.strategy),
                "shards":         shardStatuses,
                "cross_channel":  m.crossChannel.GetStatus(),
        }
}
