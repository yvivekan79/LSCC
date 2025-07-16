package consensus

import (
	"fmt"
	"lscc/core"
)

type PBFT struct{}

func NewPBFTConsensus(cfg interface{}, blockchain interface{}) (*PBFT, error) {
	return &PBFT{}, nil
}

func (pbft *PBFT) Start() error {
	return nil
}

func (pbft *PBFT) Stop() error {
	return nil
}

func (pbft *PBFT) ValidateBlock(block *core.Block) error {
	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}
	return nil
}

func (pbft *PBFT) ProposeBlock(transactions []*core.Transaction, prevBlockHash string, height uint64, shardID int) (*core.Block, error) {
	block := core.NewBlock(height, prevBlockHash, transactions, "pbft-validator", shardID)
	return block, nil
}