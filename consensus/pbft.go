package consensus

import (
	"fmt"
	"lscc/core"
	"lscc/utils"
	"sync"
)

type PBFT struct {
	nodeID       string
	validators   []string
	currentView  int
	currentPhase string
	proposals    map[string]*core.Block
	prepares     map[string]map[string]bool
	commits      map[string]map[string]bool
	mu           sync.RWMutex
	logger       *utils.Logger
}

func NewPBFT(nodeID string, validators []string) *PBFT {
	return &PBFT{
		nodeID:       nodeID,
		validators:   validators,
		currentView:  0,
		currentPhase: "prepare",
		proposals:    make(map[string]*core.Block),
		prepares:     make(map[string]map[string]bool),
		commits:      make(map[string]map[string]bool),
		logger:       utils.GetLogger(),
	}
}

func NewPBFTConsensus(cfg interface{}, blockchain interface{}) (*PBFT, error) {
	return NewPBFT("default", []string{}), nil
}

func (pbft *PBFT) Start() error {
	pbft.logger.Info("PBFT consensus engine started", "nodeID", pbft.nodeID)
	return nil
}

func (pbft *PBFT) Stop() error {
	pbft.logger.Info("PBFT consensus engine stopped")
	return nil
}

func (pbft *PBFT) ValidateBlock(block *core.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if !block.Validate() {
		return fmt.Errorf("block validation failed")
	}

	return nil
}

func (pbft *PBFT) ProposeBlock(transactions []*core.Transaction, prevBlockHash string, height uint64, shardID int) (*core.Block, error) {
	pbft.mu.Lock()
	defer pbft.mu.Unlock()

	block := core.NewBlock(height, prevBlockHash, transactions, pbft.nodeID, shardID)

	// Store proposal
	pbft.proposals[block.Hash] = block
	pbft.prepares[block.Hash] = make(map[string]bool)
	pbft.commits[block.Hash] = make(map[string]bool)

	pbft.logger.Info("Block proposed", "height", block.Height, "hash", block.Hash)

	return block, nil
}

func (pbft *PBFT) HandlePrepare(blockHash string, validator string) {
	pbft.mu.Lock()
	defer pbft.mu.Unlock()

	if pbft.prepares[blockHash] == nil {
		pbft.prepares[blockHash] = make(map[string]bool)
	}

	pbft.prepares[blockHash][validator] = true
	pbft.logger.Info("Prepare received", "blockHash", blockHash, "validator", validator)
}

func (pbft *PBFT) HandleCommit(blockHash string, validator string) {
	pbft.mu.Lock()
	defer pbft.mu.Unlock()

	if pbft.commits[blockHash] == nil {
		pbft.commits[blockHash] = make(map[string]bool)
	}

	pbft.commits[blockHash][validator] = true
	pbft.logger.Info("Commit received", "blockHash", blockHash, "validator", validator)
}

func (pbft *PBFT) IsBlockCommitted(blockHash string) bool {
	pbft.mu.RLock()
	defer pbft.mu.RUnlock()

	requiredCommits := minPbft(len(pbft.validators)*2/3+1, len(pbft.validators))
	return len(pbft.commits[blockHash]) >= requiredCommits
}

func minPbft(a, b int) int {
	if a < b {
		return a
	}
	return b
}