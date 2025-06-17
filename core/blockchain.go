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
	Blocks           []*Block
	Transactions     map[string]*Transaction
	TransactionPool  *TransactionPool
	Config           *config.Config
	mu               sync.RWMutex
	logger           *utils.Logger
}

func NewBlockchain(cfg *config.Config) *Blockchain {
	logger := utils.GetLogger()
	bc := &Blockchain{
		Blocks:          []*Block{},
		Transactions:    make(map[string]*Transaction),
		TransactionPool: NewTransactionPool(),
		Config:          cfg,
		logger:          logger,
	}
	genesis := createGenesisBlock(cfg.NodeID)
	bc.Blocks = append(bc.Blocks, genesis)
	logger.Info("Blockchain initialized with genesis block", "shardID", cfg.ShardID)
	return bc
}

func createGenesisBlock(createdBy string) *Block {
	genesisBlock := &Block{
		Header: BlockHeader{
			PreviousHash: "0",
			Timestamp:    time.Now().Unix(),
			Layer:        0,
			Height:       0,
			ValidatorID:  createdBy,
		},
		Transactions: []Transaction{},
		ShardID:      0,
	}
	genesisBlock.Header.MerkleRoot = genesisBlock.CalculateMerkleRoot()
	return genesisBlock
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.Blocks) > 0 && !block.IsValid() {
		return errors.New("invalid block")
	}

	for _, tx := range block.Transactions {
		bc.Transactions[tx.Hash] = &tx
	}

	bc.Blocks = append(bc.Blocks, block)
	return nil
}

func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) CollectPendingTransactions(max int) []*Transaction {
	return bc.TransactionPool.GetPendingTransactions(max)
}
