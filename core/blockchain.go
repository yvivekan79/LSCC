
package core

import (
	"fmt"
	"sync"
)

type Blockchain struct {
	Blocks      []*Block
	Mempool     []*Transaction
	ShardID     int
	NodeID      string
	Height      uint64
	mu          sync.RWMutex
	txIndex     map[string]*Transaction
}

func NewBlockchain(shardID int, nodeID string) *Blockchain {
	bc := &Blockchain{
		Blocks:  []*Block{},
		Mempool: []*Transaction{},
		ShardID: shardID,
		NodeID:  nodeID,
		txIndex: make(map[string]*Transaction),
	}

	// Add genesis block
	genesis := GenesisBlock(shardID)
	bc.Blocks = append(bc.Blocks, genesis)

	return bc
}

func (bc *Blockchain) GetBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Create a copy to avoid race conditions
	blocks := make([]*Block, len(bc.Blocks))
	copy(blocks, bc.Blocks)
	return blocks
}

func (bc *Blockchain) ValidateBlock(block *Block) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Basic validation
	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}

	// Check if block height is correct
	expectedHeight := bc.GetHeight() + 1
	if block.Height != expectedHeight {
		return fmt.Errorf("invalid block height: expected %d, got %d", expectedHeight, block.Height)
	}

	// Check previous block hash
	if len(bc.Blocks) > 0 {
		lastBlock := bc.Blocks[len(bc.Blocks)-1]
		if block.PrevBlockHash != lastBlock.Hash {
			return fmt.Errorf("invalid previous block hash")
		}
	}

	return nil
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate the block before adding
	if err := bc.ValidateBlock(block); err != nil {
		return err
	}

	bc.Blocks = append(bc.Blocks, block)
	bc.Height = uint64(len(bc.Blocks))
	return nil
}

func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if !tx.Validate() {
		return fmt.Errorf("invalid transaction")
	}

	// Check if transaction already exists
	if bc.txIndex[tx.Hash] != nil {
		return fmt.Errorf("transaction already exists")
	}

	// Check if transaction is already in mempool
	for _, mempoolTx := range bc.Mempool {
		if mempoolTx.Hash == tx.Hash {
			return fmt.Errorf("transaction already in mempool")
		}
	}

	bc.Mempool = append(bc.Mempool, tx)
	return nil
}

func (bc *Blockchain) GetLastBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) GetBlock(height uint64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Blocks)) {
		return nil
	}
	return bc.Blocks[height]
}

func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.Blocks {
		if block.Hash == hash {
			return block
		}
	}
	return nil
}

func (bc *Blockchain) GetTransaction(hash string) *Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.txIndex[hash]
}

func (bc *Blockchain) GetPendingTransactions() []*Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Create a copy to avoid race conditions
	pending := make([]*Transaction, len(bc.Mempool))
	copy(pending, bc.Mempool)
	return pending
}

func (bc *Blockchain) RemoveFromMempool(txHash string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for i, tx := range bc.Mempool {
		if tx.Hash == txHash {
			bc.Mempool = append(bc.Mempool[:i], bc.Mempool[i+1:]...)
			break
		}
	}
}

func (bc *Blockchain) GetHeight() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.Blocks) == 0 {
		return 0
	}
	return bc.Blocks[len(bc.Blocks)-1].Height
}

func (bc *Blockchain) GetBalance(address string) float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	balance := 0.0

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.To == address {
				balance += tx.Amount
			}
			if tx.From == address {
				balance -= (tx.Amount + tx.Fee)
			}
		}
	}

	return balance
}

func (bc *Blockchain) GetBlockchainInfo() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return map[string]interface{}{
		"shard_id":      bc.ShardID,
		"node_id":       bc.NodeID,
		"height":        len(bc.Blocks) - 1, // Subtract genesis block
		"total_blocks":  len(bc.Blocks),
		"pending_txs":   len(bc.Mempool),
		"total_txs":     len(bc.txIndex),
	}
}

func (bc *Blockchain) GetCrossShardTransactions() []*Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var crossShardTxs []*Transaction

	for _, tx := range bc.Mempool {
		if tx.SourceShard != tx.TargetShard {
			crossShardTxs = append(crossShardTxs, tx)
		}
	}

	return crossShardTxs
}

func (bc *Blockchain) GetNetworkStats() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	totalTxs := 0
	for _, block := range bc.Blocks {
		totalTxs += len(block.Transactions)
	}

	return map[string]interface{}{
		"total_blocks":       len(bc.Blocks),
		"total_transactions": totalTxs,
		"pending_txs":        len(bc.Mempool),
		"blockchain_height":  bc.GetHeight(),
		"shard_id":          bc.ShardID,
		"node_id":           bc.NodeID,
	}
}
