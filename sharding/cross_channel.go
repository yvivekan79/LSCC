
package sharding

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "lscc/core"
    "lscc/utils"
    "sync"
    "time"
)

type CrossChannelConsensus struct {
    channels             map[string]*Channel
    relayNodes           map[string]bool
    crossShardQueues     map[int][]*core.Transaction
    pendingRelayBlocks   map[string]*RelayBlock
    validatedRelayBlocks map[string]*RelayBlock
    validationThreshold  int
    mu                   sync.RWMutex
    logger               *utils.Logger
}

type Channel struct {
    SourceShard      int
    TargetShard      int
    Transactions     []*core.Transaction
    LastProcessed    int64
    ValidationCount  int
    mu               sync.RWMutex
}

type RelayBlock struct {
    ID               string
    Timestamp        int64
    CrossShardTxs    []*core.Transaction
    SourceShards     []int
    TargetShards     []int
    Hash             string
    Validations      map[string]bool
    IsFinalized      bool
    CreatedBy        string
}

func NewCrossChannelConsensus() *CrossChannelConsensus {
    return &CrossChannelConsensus{
        channels:             make(map[string]*Channel),
        relayNodes:           make(map[string]bool),
        crossShardQueues:     make(map[int][]*core.Transaction),
        pendingRelayBlocks:   make(map[string]*RelayBlock),
        validatedRelayBlocks: make(map[string]*RelayBlock),
        validationThreshold:  2, // Minimum validations needed
        logger:               utils.InitLoggerLevel("debug"),
    }
}

func (cc *CrossChannelConsensus) RegisterRelayNode(nodeID string) {
    cc.mu.Lock()
    defer cc.mu.Unlock()
    
    cc.relayNodes[nodeID] = true
    cc.logger.Info("Relay node registered", "nodeID", nodeID)
}

func (cc *CrossChannelConsensus) SubmitCrossShardTransaction(tx *core.Transaction) error {
    cc.mu.Lock()
    defer cc.mu.Unlock()
    
    if tx.SourceShard == tx.TargetShard {
        return fmt.Errorf("not a cross-shard transaction")
    }
    
    // Add to cross-shard queue for target shard
    cc.crossShardQueues[tx.TargetShard] = append(cc.crossShardQueues[tx.TargetShard], tx)
    
    // Create or update channel
    channelID := fmt.Sprintf("%d-%d", tx.SourceShard, tx.TargetShard)
    if cc.channels[channelID] == nil {
        cc.channels[channelID] = &Channel{
            SourceShard:  tx.SourceShard,
            TargetShard:  tx.TargetShard,
            Transactions: []*core.Transaction{},
        }
    }
    
    cc.channels[channelID].mu.Lock()
    cc.channels[channelID].Transactions = append(cc.channels[channelID].Transactions, tx)
    cc.channels[channelID].mu.Unlock()
    
    cc.logger.Info("Cross-shard transaction submitted", 
        "txHash", tx.Hash, 
        "from", tx.SourceShard, 
        "to", tx.TargetShard)
    
    // Check if we should create a relay block
    if len(cc.crossShardQueues[tx.TargetShard]) >= 5 {
        go cc.createRelayBlock(tx.TargetShard)
    }
    
    return nil
}

func (cc *CrossChannelConsensus) createRelayBlock(targetShard int) {
    cc.mu.Lock()
    defer cc.mu.Unlock()
    
    if len(cc.crossShardQueues[targetShard]) == 0 {
        return
    }
    
    // Create relay block with pending cross-shard transactions
    relayBlock := &RelayBlock{
        ID:            fmt.Sprintf("relay_%d_%d", targetShard, time.Now().Unix()),
        Timestamp:     time.Now().Unix(),
        CrossShardTxs: make([]*core.Transaction, len(cc.crossShardQueues[targetShard])),
        TargetShards:  []int{targetShard},
        Validations:   make(map[string]bool),
        IsFinalized:   false,
        CreatedBy:     "relay_system",
    }
    
    copy(relayBlock.CrossShardTxs, cc.crossShardQueues[targetShard])
    
    // Calculate hash
    hashData := fmt.Sprintf("%s:%d", relayBlock.ID, relayBlock.Timestamp)
    for _, tx := range relayBlock.CrossShardTxs {
        hashData += ":" + tx.Hash
    }
    hash := sha256.Sum256([]byte(hashData))
    relayBlock.Hash = hex.EncodeToString(hash[:])
    
    // Collect source shards
    sourceShards := make(map[int]bool)
    for _, tx := range relayBlock.CrossShardTxs {
        sourceShards[tx.SourceShard] = true
    }
    for shard := range sourceShards {
        relayBlock.SourceShards = append(relayBlock.SourceShards, shard)
    }
    
    cc.pendingRelayBlocks[relayBlock.ID] = relayBlock
    
    // Clear the queue
    cc.crossShardQueues[targetShard] = []*core.Transaction{}
    
    cc.logger.Info("Relay block created", 
        "id", relayBlock.ID, 
        "hash", relayBlock.Hash,
        "txCount", len(relayBlock.CrossShardTxs),
        "targetShard", targetShard)
    
    // Start validation process
    go cc.validateRelayBlock(relayBlock.ID)
}

func (cc *CrossChannelConsensus) validateRelayBlock(relayBlockID string) {
    cc.mu.RLock()
    relayBlock, exists := cc.pendingRelayBlocks[relayBlockID]
    if !exists {
        cc.mu.RUnlock()
        return
    }
    cc.mu.RUnlock()
    
    // Simulate validation by relay nodes
    for nodeID := range cc.relayNodes {
        time.Sleep(100 * time.Millisecond) // Simulate validation time
        
        // Validate relay block (simplified validation)
        isValid := cc.performRelayBlockValidation(relayBlock)
        
        if isValid {
            cc.mu.Lock()
            relayBlock.Validations[nodeID] = true
            cc.mu.Unlock()
            
            cc.logger.Info("Relay block validated", 
                "relayBlockID", relayBlockID, 
                "validator", nodeID,
                "validationCount", len(relayBlock.Validations))
        }
        
        // Check if we have enough validations
        if len(relayBlock.Validations) >= cc.validationThreshold {
            cc.finalizeRelayBlock(relayBlockID)
            break
        }
    }
}

func (cc *CrossChannelConsensus) performRelayBlockValidation(relayBlock *RelayBlock) bool {
    // Validate relay block structure and transactions
    if relayBlock.ID == "" || relayBlock.Hash == "" {
        return false
    }
    
    if len(relayBlock.CrossShardTxs) == 0 {
        return false
    }
    
    // Validate each transaction in the relay block
    for _, tx := range relayBlock.CrossShardTxs {
        if tx.SourceShard == tx.TargetShard {
            cc.logger.Error("Invalid cross-shard transaction in relay block", "txHash", tx.Hash)
            return false
        }
        
        if tx.Hash == "" {
            cc.logger.Error("Transaction missing hash in relay block", "from", tx.From, "to", tx.To)
            return false
        }
    }
    
    // Validate hash
    hashData := fmt.Sprintf("%s:%d", relayBlock.ID, relayBlock.Timestamp)
    for _, tx := range relayBlock.CrossShardTxs {
        hashData += ":" + tx.Hash
    }
    hash := sha256.Sum256([]byte(hashData))
    calculatedHash := hex.EncodeToString(hash[:])
    
    return calculatedHash == relayBlock.Hash
}

func (cc *CrossChannelConsensus) finalizeRelayBlock(relayBlockID string) {
    cc.mu.Lock()
    defer cc.mu.Unlock()
    
    relayBlock, exists := cc.pendingRelayBlocks[relayBlockID]
    if !exists {
        return
    }
    
    relayBlock.IsFinalized = true
    cc.validatedRelayBlocks[relayBlockID] = relayBlock
    delete(cc.pendingRelayBlocks, relayBlockID)
    
    cc.logger.Info("Relay block finalized", 
        "relayBlockID", relayBlockID,
        "validationCount", len(relayBlock.Validations),
        "txCount", len(relayBlock.CrossShardTxs))
    
    // Update channels
    for _, targetShard := range relayBlock.TargetShards {
        for _, sourceShard := range relayBlock.SourceShards {
            channelID := fmt.Sprintf("%d-%d", sourceShard, targetShard)
            if channel, exists := cc.channels[channelID]; exists {
                channel.mu.Lock()
                channel.LastProcessed = relayBlock.Timestamp
                channel.mu.Unlock()
            }
        }
    }
}

func (cc *CrossChannelConsensus) GetCrossShardTransactions(shardID int) []*core.Transaction {
    cc.mu.RLock()
    defer cc.mu.RUnlock()
    
    var transactions []*core.Transaction
    
    // Look for finalized relay blocks targeting this shard
    for _, relayBlock := range cc.validatedRelayBlocks {
        for _, targetShard := range relayBlock.TargetShards {
            if targetShard == shardID {
                transactions = append(transactions, relayBlock.CrossShardTxs...)
            }
        }
    }
    
    return transactions
}

func (cc *CrossChannelConsensus) GetChannelStatus(sourceShardID, targetShardID int) map[string]interface{} {
    cc.mu.RLock()
    defer cc.mu.RUnlock()
    
    channelID := fmt.Sprintf("%d-%d", sourceShardID, targetShardID)
    channel, exists := cc.channels[channelID]
    
    if !exists {
        return map[string]interface{}{
            "exists":         false,
            "source_shard":   sourceShardID,
            "target_shard":   targetShardID,
        }
    }
    
    channel.mu.RLock()
    defer channel.mu.RUnlock()
    
    return map[string]interface{}{
        "exists":           true,
        "source_shard":     channel.SourceShard,
        "target_shard":     channel.TargetShard,
        "pending_txs":      len(channel.Transactions),
        "last_processed":   channel.LastProcessed,
        "validation_count": channel.ValidationCount,
    }
}

func (cc *CrossChannelConsensus) GetSystemStatus() map[string]interface{} {
    cc.mu.RLock()
    defer cc.mu.RUnlock()
    
    totalPendingTxs := 0
    for _, queue := range cc.crossShardQueues {
        totalPendingTxs += len(queue)
    }
    
    return map[string]interface{}{
        "relay_nodes":          len(cc.relayNodes),
        "active_channels":      len(cc.channels),
        "pending_relay_blocks": len(cc.pendingRelayBlocks),
        "finalized_relay_blocks": len(cc.validatedRelayBlocks),
        "total_pending_txs":    totalPendingTxs,
        "validation_threshold": cc.validationThreshold,
    }
}
