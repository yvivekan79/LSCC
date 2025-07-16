package consensus

import (
    "lscc/core"
    "lscc/config"
    "lscc/utils"
    "crypto/sha256"
    "encoding/hex"
)

type PoW struct {
    blockchain *core.Blockchain
    config     *config.Config
    logger     *utils.Logger
}

func NewPoW(blockchain *core.Blockchain, cfg *config.Config, logger *utils.Logger) *PoW {
    return &PoW{
        blockchain: blockchain,
        config:     cfg,
        logger:     logger,
    }
}

func (pow *PoW) MineBlock(txs []*core.Transaction) *core.Block {
    lastBlock := pow.blockchain.GetLastBlock()
    var nonce int
    var hash string

    for {
        data := lastBlock.Hash + string(nonce)
        h := sha256.Sum256([]byte(data))
        hash = hex.EncodeToString(h[:])
        if hash[:pow.config.ConsensusParams.Difficulty] == string(make([]byte, pow.config.ConsensusParams.Difficulty)) {
            break
        }
        nonce++
    }

    newBlock := core.NewBlock(lastBlock.Hash, txs, pow.config.ShardID, pow.config.Layer)
    newBlock.Header.Nonce = nonce
    newBlock.Hash = hash
    return newBlock
}

