package consensus

import (
	"fmt"
	"lscc/core"
)

type PoW struct{}

func NewPoWConsensus(cfg interface{}, blockchain interface{}) (*PoW, error) {
	return &PoW{}, nil
}

func (pow *PoW) Start() error {
	return nil
}

func (pow *PoW) Stop() error {
	return nil
}

func (pow *PoW) ValidateBlock(block *core.Block) error {
	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}
	return nil
}

func (pow *PoW) ProposeBlock(transactions []*core.Transaction, prevBlockHash string, height uint64, shardID int) (*core.Block, error) {
	block := core.NewBlock(height, prevBlockHash, transactions, "pow-validator", shardID)
	return block, nil
}