package consensus

import (
	"fmt"
	"lscc/core"
)

type PoSConsensus struct {
	blockchain core.Blockchain
	validators map[string]float64
}

func NewPoSConsensus(cfg interface{}, blockchain interface{}) (*PoSConsensus, error) {
	return &PoSConsensus{
		validators: make(map[string]float64),
	}, nil
}

func (pos *PoSConsensus) Start() error {
	return nil
}

func (pos *PoSConsensus) Stop() error {
	return nil
}

func (pos *PoSConsensus) ValidateBlock(block *core.Block) error {
	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}
	return nil
}

func (pos *PoSConsensus) ProposeBlock(transactions []*core.Transaction, prevBlockHash string, height uint64, shardID int) (*core.Block, error) {
	block := core.NewBlock(height, prevBlockHash, transactions, "pos-validator", shardID)
	return block, nil
}

func (pos *PoSConsensus) AddValidator(nodeID string, stake float64) {
	pos.validators[nodeID] = stake
}

func (pos *PoSConsensus) GetValidators() map[string]float64 {
	return pos.validators
}