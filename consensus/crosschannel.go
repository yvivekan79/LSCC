
package consensus

import (
	"fmt"
	"lscc/core"
	"lscc/utils"
	"sync"
)

type CrossChannel struct {
	sourceShardID int
	targetShardID int
	relayNodes    []string
	pendingTxs    map[string]*core.Transaction
	confirmedTxs  map[string]*core.Transaction
	mu            sync.RWMutex
	logger        *utils.Logger
}

func NewCrossChannel(sourceShardID, targetShardID int, relayNodes []string) *CrossChannel {
	return &CrossChannel{
		sourceShardID: sourceShardID,
		targetShardID: targetShardID,
		relayNodes:    relayNodes,
		pendingTxs:    make(map[string]*core.Transaction),
		confirmedTxs:  make(map[string]*core.Transaction),
		logger:        utils.GetLogger(),
	}
}

func (cc *CrossChannel) ProcessCrossShardTx(tx *core.Transaction) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Validate cross-shard transaction
	if tx.SourceShard != cc.sourceShardID || tx.TargetShard != cc.targetShardID {
		return fmt.Errorf("transaction shard mismatch")
	}

	// Add to pending transactions
	cc.pendingTxs[tx.Hash] = tx
	cc.logger.Info("Cross-shard transaction added to pending", "txHash", tx.Hash)

	return nil
}

func (cc *CrossChannel) ConfirmTransaction(txHash string, block *core.Block) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Validate block contains the transaction
	found := false
	for _, tx := range block.Transactions {
		if tx.Hash == txHash {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("transaction not found in block")
	}

	// Move from pending to confirmed
	if tx, exists := cc.pendingTxs[txHash]; exists {
		cc.confirmedTxs[txHash] = tx
		delete(cc.pendingTxs, txHash)
		cc.logger.Info("Cross-shard transaction confirmed", "txHash", txHash, "blockHeight", block.Height)
	}

	return nil
}

func (cc *CrossChannel) GetPendingTransactions() map[string]*core.Transaction {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	result := make(map[string]*core.Transaction)
	for hash, tx := range cc.pendingTxs {
		result[hash] = tx
	}
	return result
}

func (cc *CrossChannel) GetConfirmedTransactions() map[string]*core.Transaction {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	result := make(map[string]*core.Transaction)
	for hash, tx := range cc.confirmedTxs {
		result[hash] = tx
	}
	return result
}
