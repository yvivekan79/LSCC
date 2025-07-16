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
    block := core.NewBlock(lastBlock.Hash, txs, pos.config.ShardID, pos.config.Layer)
    block.Header.Timestamp = time.Now().Unix()
    return block
}

