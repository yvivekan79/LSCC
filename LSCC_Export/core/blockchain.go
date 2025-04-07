package core

import (
        "errors"
        "sync"
        "time"

        "lscc/config"
        "lscc/utils"
)

// Blockchain represents the main blockchain data structure
type Blockchain struct {
        Blocks       []*Block
        Transactions map[string]*Transaction // Map of transaction hash to transaction
        Config       *config.Config
        mu           sync.RWMutex
        logger       *utils.Logger
}

// NewBlockchain creates a new blockchain with a genesis block
func NewBlockchain(cfg *config.Config) *Blockchain {
        logger := utils.GetLogger()
        bc := &Blockchain{
                Blocks:       []*Block{},
                Transactions: make(map[string]*Transaction),
                Config:       cfg,
                logger:       logger,
        }

        // Create genesis block
        genesis := createGenesisBlock(cfg.ShardID)
        bc.AddBlock(genesis)

        logger.Info("Blockchain initialized with genesis block", "shardID", cfg.ShardID)
        return bc
}

// createGenesisBlock creates the genesis block for the blockchain
func createGenesisBlock(shardID int) *Block {
        genesisBlock := NewBlock("0", 0, shardID, 0, "genesis")
        genesisBlock.Header.Timestamp = time.Now().Unix()
        genesisBlock.Header.MerkleRoot = genesisBlock.CalculateMerkleRoot()
        return genesisBlock
}

// AddBlock adds a block to the blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
        bc.mu.Lock()
        defer bc.mu.Unlock()

        // Validate block before adding
        if len(bc.Blocks) > 0 {
                lastBlock := bc.Blocks[len(bc.Blocks)-1]
                if !block.IsValid(lastBlock) {
                        return errors.New("invalid block")
                }
        }

        // Add transactions to the transaction pool
        for _, tx := range block.Transactions {
                bc.Transactions[tx.Hash] = &tx
        }

        // Add block to the chain
        bc.Blocks = append(bc.Blocks, block)
        bc.logger.Info("Added new block to the chain", 
                "height", block.Header.Height,
                "hash", block.Hash,
                "txs", len(block.Transactions),
                "shardID", block.ShardID)
        
        return nil
}

// GetBlockByHeight retrieves a block by its height
func (bc *Blockchain) GetBlockByHeight(height uint64) *Block {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        for _, block := range bc.Blocks {
                if block.Header.Height == height {
                        return block
                }
        }
        return nil
}

// GetBlockByHash retrieves a block by its hash
func (bc *Blockchain) GetBlockByHash(hash string) *Block {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        for _, block := range bc.Blocks {
                blockHash, err := block.Hash()
                if err != nil {
                        continue
                }
                if blockHash == hash {
                        return block
                }
        }
        return nil
}

// GetLatestBlock returns the latest block in the chain
func (bc *Blockchain) GetLatestBlock() *Block {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        if len(bc.Blocks) == 0 {
                return nil
        }
        return bc.Blocks[len(bc.Blocks)-1]
}

// GetHeight returns the current height of the blockchain
func (bc *Blockchain) GetHeight() uint64 {
        latestBlock := bc.GetLatestBlock()
        if latestBlock == nil {
                return 0
        }
        return latestBlock.Header.Height
}

// ValidateChain checks if the blockchain is valid
func (bc *Blockchain) ValidateChain() bool {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        for i := 1; i < len(bc.Blocks); i++ {
                if !bc.Blocks[i].IsValid(bc.Blocks[i-1]) {
                        return false
                }
        }
        return true
}

// GetTransaction gets a transaction by its hash
func (bc *Blockchain) GetTransaction(hash string) *Transaction {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        tx, exists := bc.Transactions[hash]
        if !exists {
                return nil
        }
        return tx
}

// AddTransaction adds a transaction to the pool
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
        bc.mu.Lock()
        defer bc.mu.Unlock()

        // Check if transaction already exists
        if _, exists := bc.Transactions[tx.Hash]; exists {
                return errors.New("transaction already exists")
        }

        // Validate transaction
        if !tx.IsValid() {
                return errors.New("invalid transaction")
        }

        bc.Transactions[tx.Hash] = tx
        bc.logger.Info("Added new transaction to pool", "hash", tx.Hash)
        return nil
}

// GetPendingTransactions returns all pending transactions
func (bc *Blockchain) GetPendingTransactions() []*Transaction {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        var pendingTxs []*Transaction
        for _, tx := range bc.Transactions {
                if !tx.IsConfirmed {
                        pendingTxs = append(pendingTxs, tx)
                }
        }
        return pendingTxs
}

// SerializeBlockchain exports the blockchain data
func (bc *Blockchain) SerializeBlockchain() ([]byte, error) {
        bc.mu.RLock()
        defer bc.mu.RUnlock()

        // Implement serialization logic here
        // This is a placeholder
        return nil, errors.New("not implemented")
}

// DeserializeBlockchain imports blockchain data
func DeserializeBlockchain(data []byte, cfg *config.Config) (*Blockchain, error) {
        // Implement deserialization logic here
        // This is a placeholder
        return nil, errors.New("not implemented")
}
