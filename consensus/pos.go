package consensus

import (
	"lscc/core"
	"lscc/config"
	"lscc/utils"
	"time"
)

type PoS struct {
	blockchain *core.Blockchain
	config     *config.Config
	logger     *utils.Logger
}

func NewPoS(blockchain *core.Blockchain, cfg *config.Config, logger *utils.Logger) *PoS {
	return &PoS{
		blockchain: blockchain,
		config:     cfg,
		logger:     logger,
	}
}

func (pos *PoS) SelectValidator() string {
	return pos.config.NodeID // mock
}

func (pos *PoS) GenerateBlock(txs []*core.Transaction) *core.Block {
	lastBlock := pos.blockchain.GetLastBlock()
	height := uint64(len(pos.blockchain.Blocks))
	block := core.NewBlock(height, lastBlock.Hash, txs, "", pos.config.ShardID)
	block.Height = height
	block.Header.Timestamp = time.Now().Unix()
	return block
}