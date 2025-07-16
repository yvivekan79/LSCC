
package consensus

import (
	"fmt"
	"lscc/core"
	"lscc/utils"
	"math/rand"
	"time"
)

type PoS struct {
	validators    map[string]float64 // validator -> stake amount
	totalStake    float64
	currentEpoch  int
	epochDuration time.Duration
	logger        *utils.Logger
}

func NewPoS() *PoS {
	return &PoS{
		validators:    make(map[string]float64),
		totalStake:    0,
		currentEpoch:  0,
		epochDuration: 30 * time.Second,
		logger:        utils.GetLogger(),
	}
}

func NewPoSConsensus(cfg interface{}, blockchain interface{}) (*PoS, error) {
	return NewPoS(), nil
}

func (pos *PoS) Start() error {
	pos.logger.Info("PoS consensus engine started", "epoch", pos.currentEpoch)
	return nil
}

func (pos *PoS) Stop() error {
	pos.logger.Info("PoS consensus engine stopped")
	return nil
}

func (pos *PoS) ValidateBlock(block *core.Block) error {
	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}
	
	// Validate that the block creator has sufficient stake
	if pos.validators[block.Validator] <= 0 {
		return fmt.Errorf("block creator has no stake")
	}
	
	return nil
}

func (pos *PoS) ProposeBlock(transactions []*core.Transaction, prevBlockHash string, height uint64, shardID int) (*core.Block, error) {
	validator := pos.selectValidator()
	if validator == "" {
		return nil, fmt.Errorf("no validator selected")
	}

	block := core.NewBlock(height, prevBlockHash, transactions, validator, shardID)
	pos.logger.Info("Block proposed by validator", "validator", validator, "height", height)
	
	return block, nil
}

func (pos *PoS) AddValidator(nodeID string, stake float64) {
	pos.validators[nodeID] = stake
	pos.totalStake += stake
	pos.logger.Info("Validator added", "nodeID", nodeID, "stake", stake)
}

func (pos *PoS) RemoveValidator(nodeID string) {
	if stake, exists := pos.validators[nodeID]; exists {
		pos.totalStake -= stake
		delete(pos.validators, nodeID)
		pos.logger.Info("Validator removed", "nodeID", nodeID)
	}
}

func (pos *PoS) selectValidator() string {
	if pos.totalStake <= 0 {
		return ""
	}

	// Weighted random selection based on stake
	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Float64() * pos.totalStake
	
	accumulator := 0.0
	for validator, stake := range pos.validators {
		accumulator += stake
		if randomValue <= accumulator {
			return validator
		}
	}
	
	// Fallback to first validator
	for validator := range pos.validators {
		return validator
	}
	
	return ""
}

func (pos *PoS) GetValidators() map[string]float64 {
	result := make(map[string]float64)
	for validator, stake := range pos.validators {
		result[validator] = stake
	}
	return result
}
