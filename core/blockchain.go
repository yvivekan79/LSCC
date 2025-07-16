
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Blockchain struct {
	Blocks      []*Block
	Mempool     []*Transaction
	NodeID      string
	ShardID     int
	mu          sync.RWMutex
	GenesisTime time.Time
}

func NewBlockchain(nodeID string, shardID int) *Blockchain {
	bc := &Blockchain{
		Blocks:      make([]*Block, 0),
		Mempool:     make([]*Transaction, 0),
		NodeID:      nodeID,
		ShardID:     shardID,
		GenesisTime: time.Now(),
	}
	
	// Create genesis block
	genesisBlock := &Block{
		Index:         0,
		PrevBlockHash: "0",
		Timestamp:     time.Now(),
		Transactions:  []*Transaction{},
		Validator:     nodeID,
		ShardID:       shardID,
		Hash:          "",
	}
	genesisBlock.Hash = genesisBlock.CalculateHash()
	bc.Blocks = append(bc.Blocks, genesisBlock)
	
	return bc
}

func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if err := bc.ValidateBlock(block); err != nil {
		return err
	}
	
	bc.Blocks = append(bc.Blocks, block)
	bc.removeTransactionsFromMempool(block.Transactions)
	
	return nil
}

func (bc *Blockchain) ValidateBlock(block *Block) error {
	if len(bc.Blocks) == 0 {
		return fmt.Errorf("no blocks in chain")
	}
	
	latestBlock := bc.Blocks[len(bc.Blocks)-1]
	if block.Index != latestBlock.Index+1 {
		return fmt.Errorf("invalid block index")
	}
	
	if block.PrevBlockHash != latestBlock.Hash {
		return fmt.Errorf("invalid previous block hash")
	}
	
	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}
	
	return nil
}

func (bc *Blockchain) GetHeight() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return uint64(len(bc.Blocks))
}

func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if err := tx.Validate(); err != nil {
		return err
	}
	
	bc.Mempool = append(bc.Mempool, tx)
	return nil
}

func (bc *Blockchain) GetPendingTransactions() []*Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	result := make([]*Transaction, len(bc.Mempool))
	copy(result, bc.Mempool)
	return result
}

func (bc *Blockchain) removeTransactionsFromMempool(transactions []*Transaction) {
	txMap := make(map[string]bool)
	for _, tx := range transactions {
		txMap[tx.ID] = true
	}
	
	newMempool := make([]*Transaction, 0)
	for _, tx := range bc.Mempool {
		if !txMap[tx.ID] {
			newMempool = append(newMempool, tx)
		}
	}
	bc.Mempool = newMempool
}

func (bc *Blockchain) GetTransactionByID(txID string) (*Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.ID == txID {
				return tx, nil
			}
		}
	}
	
	for _, tx := range bc.Mempool {
		if tx.ID == txID {
			return tx, nil
		}
	}
	
	return nil, fmt.Errorf("transaction not found")
}

func (bc *Blockchain) GetBlockByIndex(index uint64) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	if index >= uint64(len(bc.Blocks)) {
		return nil, fmt.Errorf("block index out of range")
	}
	
	return bc.Blocks[index], nil
}

func (bc *Blockchain) GetBlockByHash(hash string) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	for _, block := range bc.Blocks {
		if block.Hash == hash {
			return block, nil
		}
	}
	
	return nil, fmt.Errorf("block not found")
}

func (bc *Blockchain) IsValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	if len(bc.Blocks) == 0 {
		return false
	}
	
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]
		
		if currentBlock.PrevBlockHash != prevBlock.Hash {
			return false
		}
		
		if !currentBlock.Validate() {
			return false
		}
	}
	
	return true
}

func (bc *Blockchain) CalculateMerkleRoot(transactions []*Transaction) string {
	if len(transactions) == 0 {
		return ""
	}
	
	var hashes []string
	for _, tx := range transactions {
		hashes = append(hashes, tx.Hash)
	}
	
	for len(hashes) > 1 {
		var newHashes []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				hash := sha256.Sum256([]byte(combined))
				newHashes = append(newHashes, hex.EncodeToString(hash[:]))
			} else {
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}
	
	return hashes[0]
}

func (bc *Blockchain) GetStats() map[string]interface{} {
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
