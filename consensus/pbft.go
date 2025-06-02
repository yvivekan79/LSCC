
package consensus

import (
    "math/rand"
    "sync"
    "time"

    "lscc/config"
    "lscc/core"
    "lscc/utils"
)

type PBFTConsensus struct {
    blockchain *core.Blockchain
    config     *config.Config
    logger     *utils.Logger
    mu         sync.RWMutex
    quorumSize int
    stopChan   chan struct{}
}

func NewPBFTConsensus(cfg *config.Config, bc *core.Blockchain) *PBFTConsensus {
    return &PBFTConsensus{
        blockchain: bc,
        config:     cfg,
        quorumSize: 2,
        stopChan:   make(chan struct{}),
        logger:     utils.GetLogger(),
    }
}

func (pbft *PBFTConsensus) Start() error {
    go pbft.consensusLoop()
    return nil
}

func (pbft *PBFTConsensus) Stop() error {
    close(pbft.stopChan)
    return nil
}

func (pbft *PBFTConsensus) CreateBlock() (*core.Block, error) {
    txs := pbft.blockchain.CollectPendingTransactions()
    prev := pbft.blockchain.GetLatestBlock()
    return core.NewBlock(prev, txs, pbft.config.NodeID), nil
}

func (pbft *PBFTConsensus) ValidateBlock(block *core.Block) bool {
    return rand.Float64() > 0.1 // simulate successful quorum
}

func (pbft *PBFTConsensus) ProcessBlock(block *core.Block) error {
    if pbft.ValidateBlock(block) {
        return pbft.blockchain.AddBlock(block)
    }
    return utils.NewError("PBFT: Block validation failed")
}

func (pbft *PBFTConsensus) GetType() ConsensusType {
    return PBFT
}

func (pbft *PBFTConsensus) GetStatus() map[string]interface{} {
    return map[string]interface{}{
        "quorum": pbft.quorumSize,
    }
}

func (pbft *PBFTConsensus) HandleConsensusMessage(message []byte) error {
    // Not implemented
    return nil
}

func (pbft *PBFTConsensus) consensusLoop() {
    for {
        select {
        case <-pbft.stopChan:
            return
        default:
            block, _ := pbft.CreateBlock()
            if pbft.ValidateBlock(block) {
                pbft.blockchain.AddBlock(block)
                pbft.logger.Info("PBFT: Block committed")
            }
            time.Sleep(5 * time.Second)
        }
    }
}
