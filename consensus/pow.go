package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"lscc/core"
	"lscc/utils"
	"math/big"
)

type PoW struct {
	difficulty int
	logger     *utils.Logger
}

func NewPoW(difficulty int) *PoW {
	return &PoW{
		difficulty: difficulty,
		logger:     utils.GetLogger(),
	}
}

func NewPoWConsensus(cfg interface{}, blockchain interface{}) (*PoW, error) {
	return NewPoW(2), nil
}

func (pow *PoW) Start() error {
	pow.logger.Info("PoW consensus engine started", "difficulty", pow.difficulty)
	return nil
}

func (pow *PoW) Stop() error {
	pow.logger.Info("PoW consensus engine stopped")
	return nil
}

func (pow *PoW) ValidateBlock(block *core.Block) error {
	if !pow.isValidProof(block) {
		return fmt.Errorf("invalid proof of work")
	}
	return nil
}

func (pow *PoW) ProposeBlock(transactions []*core.Transaction, prevBlockHash string, height uint64, shardID int) (*core.Block, error) {
	block := core.NewBlock(height, prevBlockHash, transactions, "pow-miner", shardID)

	// Mine the block
	pow.mineBlock(block)

	return block, nil
}

func (pow *PoW) mineBlock(block *core.Block) {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-pow.difficulty))

	var hashInt big.Int
	var hash [32]byte

	pow.logger.Info("Mining block", "height", block.Height, "difficulty", pow.difficulty)

	for block.Nonce < ^uint64(0) {
		data := pow.prepareData(block)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(target) == -1 {
			break
		} else {
			block.Nonce++
		}
	}

	block.Hash = hex.EncodeToString(hash[:])
	pow.logger.Info("Block mined", "height", block.Height, "hash", block.Hash, "nonce", block.Nonce)
}

func (pow *PoW) isValidProof(block *core.Block) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-pow.difficulty))

	data := pow.prepareData(block)
	hash := sha256.Sum256(data)
	var hashInt big.Int
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(target) == -1
}

func (pow *PoW) prepareData(block *core.Block) []byte {
	data := fmt.Sprintf("%d:%d:%s:%s:%d:%d",
		block.Height, block.Timestamp, block.PrevBlockHash, block.Validator, block.ShardID, block.Nonce)

	for _, tx := range block.Transactions {
		data += ":" + tx.Hash
	}

	return []byte(data)
}

func powMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}