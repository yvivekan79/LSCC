package core

import (
    "sync"
    "lscc/utils"
    "fmt"
)



type Blockchain struct {
    Blocks       []*Block
    Transactions map[string]*Transaction
    Mu           sync.RWMutex
    logger       *utils.Logger
}

func NewBlockchain(logger *utils.Logger) *Blockchain {
    genesis := NewBlock("genesis", []*Transaction{}, 0, 0)
    return &Blockchain{
        Blocks:       []*Block{genesis},
        Transactions: make(map[string]*Transaction),
        logger:       logger,
    }
}

func (bc *Blockchain) AddBlock(block *Block) {
    bc.Mu.Lock()
    defer bc.Mu.Unlock()
    bc.Blocks = append(bc.Blocks, block)
    for _, tx := range block.Transactions {
        bc.Transactions[tx.Hash] = tx
    }
}

func (bc *Blockchain) GetLastBlock() *Block {
    bc.Mu.RLock()
    defer bc.Mu.RUnlock()
    if len(bc.Blocks) == 0 {
        return nil
    }
    return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) GetBlocks() []*Block {
    bc.Mu.RLock()
    defer bc.Mu.RUnlock()
    return bc.Blocks
}
func (bc *Blockchain) AddTransaction(tx *Transaction) {
    // Ensure the Transactions map is initialized
    if bc.Transactions == nil {
        bc.Transactions = make(map[string]*Transaction)
    }
    // Validate the transaction before adding it        
    bc.logger.Info(fmt.Sprintf("Adding transaction: %v\n", tx))
    bc.Mu.Lock()
    defer bc.Mu.Unlock()
    if tx == nil || tx.Hash == "" {
        bc.logger.Error("Invalid transaction: nil or empty hash")
        return
    }
    if _, exists := bc.Transactions[tx.Hash]; exists {
        bc.logger.Warn("Transaction already exists", "hash", tx.Hash)
        return
    }
    // Add the transaction to the map after validation
    bc.Transactions[tx.Hash] = tx
}



