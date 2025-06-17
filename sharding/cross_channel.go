package sharding

import (
        "errors"
        "sync"
        "time"

        "lscc/config"
        "lscc/core"
        "lscc/utils"
)

// CrossChannel manages cross-shard communication
type CrossChannel struct {
        manager          *Manager
        config           *config.Config
        pendingTxs       map[string]*core.Transaction
        pendingBlocks    map[string]*core.Block
        txConfirmations  map[string]map[int]bool // Maps tx hash to a map of shard IDs that confirmed it
        blockConfirmations map[string]map[int]bool // Maps block hash to a map of shard IDs that confirmed it
        mu               sync.RWMutex
        logger           *utils.Logger
}

// NewCrossChannel creates a new cross-channel mechanism
func NewCrossChannel(manager *Manager, cfg *config.Config) *CrossChannel {
        return &CrossChannel{
                manager:          manager,
                config:           cfg,
                pendingTxs:       make(map[string]*core.Transaction),
                pendingBlocks:    make(map[string]*core.Block),
                txConfirmations:  make(map[string]map[int]bool),
                blockConfirmations: make(map[string]map[int]bool),
                logger:           utils.GetLogger(),
        }
}

// PropagateTransaction propagates a transaction to the target shard
func (cc *CrossChannel) PropagateTransaction(tx *core.Transaction, sourceShard, targetShard int) error {
        cc.mu.Lock()
        defer cc.mu.Unlock()
        
        // Store pending transaction
        cc.pendingTxs[tx.Hash] = tx
        
        // Initialize confirmation map for this transaction
        if _, exists := cc.txConfirmations[tx.Hash]; !exists {
                cc.txConfirmations[tx.Hash] = make(map[int]bool)
        }
        
        cc.logger.Info("Transaction propagated through cross-channel", 
                "txHash", tx.Hash, 
                "sourceShard", sourceShard, 
                "targetShard", targetShard)
        
        return nil
}

// PropagateBlock propagates a block to the target shard
func (cc *CrossChannel) PropagateBlock(block *core.Block, targetShards []int) error {
        cc.mu.Lock()
        defer cc.mu.Unlock()
        
        blockHash, err := block.Hash()
        if err != nil {
                return err
        }
        
        // Store pending block
        cc.pendingBlocks[blockHash] = block
        
        // Initialize confirmation map for this block
        if _, exists := cc.blockConfirmations[blockHash]; !exists {
                cc.blockConfirmations[blockHash] = make(map[int]bool)
        }
        
        cc.logger.Info("Block propagated through cross-channel", 
                "blockHash", blockHash, 
                "sourceShard", block.ShardID, 
                "targetShards", targetShards)
        
        return nil
}

// ConfirmTransaction marks a transaction as confirmed by a shard
func (cc *CrossChannel) ConfirmTransaction(txHash string, shardID int) error {
        cc.mu.Lock()
        defer cc.mu.Unlock()
        
        confirmations, exists := cc.txConfirmations[txHash]
        if !exists {
                return errors.New("transaction not found in cross-channel")
        }
        
        confirmations[shardID] = true
        cc.logger.Info("Transaction confirmed by shard", "txHash", txHash, "shardID", shardID)
        
        // Check if transaction is confirmed by enough shards
        if cc.isTransactionConfirmed(txHash) {
                cc.logger.Info("Transaction fully confirmed across shards", "txHash", txHash)
                
                // Clean up
                delete(cc.pendingTxs, txHash)
                delete(cc.txConfirmations, txHash)
        }
        
        return nil
}

// ConfirmBlock marks a block as confirmed by a shard
func (cc *CrossChannel) ConfirmBlock(blockHash string, shardID int) error {
        cc.mu.Lock()
        defer cc.mu.Unlock()
        
        confirmations, exists := cc.blockConfirmations[blockHash]
        if !exists {
                return errors.New("block not found in cross-channel")
        }
        
        confirmations[shardID] = true
        cc.logger.Info("Block confirmed by shard", "blockHash", blockHash, "shardID", shardID)
        
        // Check if block is confirmed by enough shards
        if cc.isBlockConfirmed(blockHash) {
                cc.logger.Info("Block fully confirmed across shards", "blockHash", blockHash)
                
                // Clean up
                delete(cc.pendingBlocks, blockHash)
                delete(cc.blockConfirmations, blockHash)
        }
        
        return nil
}

// isTransactionConfirmed checks if a transaction is confirmed by enough shards
func (cc *CrossChannel) isTransactionConfirmed(txHash string) bool {
        confirmations, exists := cc.txConfirmations[txHash]
        if !exists {
                return false
        }
        
        // For cross-shard transactions, we need confirmations from both source and target shards
        return len(confirmations) >= 2
}

// isBlockConfirmed checks if a block is confirmed by enough shards
func (cc *CrossChannel) isBlockConfirmed(blockHash string) bool {
        confirmations, exists := cc.blockConfirmations[blockHash]
        if !exists {
                return false
        }
        
        // A block is confirmed if it's confirmed by the target shards
        // For simplicity, we'll say 2 confirmations
        return len(confirmations) >= 2
}

// CleanupOldTransactions removes old pending transactions
func (cc *CrossChannel) CleanupOldTransactions() {
        cc.mu.Lock()
        defer cc.mu.Unlock()
        
        // Remove transactions older than some threshold
        // In a real implementation, this would use timestamps
        
        cc.logger.Info("Cleaned up old cross-channel transactions")
}

// ProcessCrossShardTransaction processes a transaction received from another shard
func (cc *CrossChannel) ProcessCrossShardTransaction(tx *core.Transaction) error {
        if !tx.IsCrossShard() {
                return errors.New("not a cross-shard transaction")
        }
        
        // Get the target shard
        targetShard, err := cc.manager.GetShard(tx.TargetShard)
        if err != nil {
                return err
        }
        
        // Process in target shard
        targetShard.AddCrossShardTransaction(tx)
        
        // Mark as confirmed by target shard
        cc.ConfirmTransaction(tx.Hash, tx.TargetShard)
        
        cc.logger.Info("Processed cross-shard transaction", 
                "txHash", tx.Hash, 
                "sourceShard", tx.SourceShard, 
                "targetShard", tx.TargetShard)
        
        return nil
}

// Start starts the cross-channel service
func (cc *CrossChannel) Start() error {
        // Start cleanup routine
        go func() {
                ticker := time.NewTicker(5 * time.Minute)
                defer ticker.Stop()
                
                for {
                        select {
                        case <-ticker.C:
                                cc.CleanupOldTransactions()
                        }
                }
        }()
        
        cc.logger.Info("Cross-channel service started")
        return nil
}

// Stop stops the cross-channel service
func (cc *CrossChannel) Stop() error {
        cc.logger.Info("Cross-channel service stopped")
        return nil
}

// GetStatus returns the status of the cross-channel
func (cc *CrossChannel) GetStatus() map[string]interface{} {
        cc.mu.RLock()
        defer cc.mu.RUnlock()
        
        return map[string]interface{}{
                "pending_tx_count":    len(cc.pendingTxs),
                "pending_block_count": len(cc.pendingBlocks),
                "tx_confirmations":    len(cc.txConfirmations),
                "block_confirmations": len(cc.blockConfirmations),
        }
}
