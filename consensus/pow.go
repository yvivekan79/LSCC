
package consensus

import (
    "crypto/sha256"
    "encoding/hex"
    "math/rand"
    "sync"
    "time"

    "lscc/config"
    "lscc/core"
    "lscc/utils"
)

type PoWConsensus struct {
    blockchain *core.Blockchain
    config     *config.Config
    difficulty int
    running    bool
    stopChan   chan struct{}
    mu         sync.RWMutex
    logger     *utils.Logger
}

func NewPoWConsensus(cfg *config.Config, bc *core.Blockchain) *PoWConsensus {
    return &PoWConsensus{
        blockchain: bc,
        config:     cfg,
        difficulty: 4,
        stopChan:   make(chan struct{}),
        logger:     utils.GetLogger(),
    }
}

func (pow *PoWConsensus) Start() error {
    pow.running = true
    go pow.miningLoop()
    return nil
}

func (pow *PoWConsensus) Stop() error {
    pow.running = false
    close(pow.stopChan)
    return nil
}

func (pow *PoWConsensus) CreateBlock() (*core.Block, error) {
    txs := pow.blockchain.CollectPendingTransactions()
    prevBlock := pow.blockchain.GetLatestBlock()
    block := core.NewBlock(prevBlock, txs, pow.config.NodeID)
    return block, nil
}

func (pow *PoWConsensus) ValidateBlock(block *core.Block) bool {
    hash := sha256.Sum256([]byte(block.Hash()))
    return hasLeadingZeros(hex.EncodeToString(hash[:]), pow.difficulty)
}

func (pow *PoWConsensus) ProcessBlock(block *core.Block) error {
    if pow.ValidateBlock(block) {
        return pow.blockchain.AddBlock(block)
    }
    return utils.NewError("Invalid PoW block")
}

func (pow *PoWConsensus) GetType() ConsensusType {
    return ProofOfWork
}

func (pow *PoWConsensus) GetStatus() map[string]interface{} {
    return map[string]interface{}{
        "running":    pow.running,
        "difficulty": pow.difficulty,
    }
}

func (pow *PoWConsensus) HandleConsensusMessage(message []byte) error {
    // Not implemented
    return nil
}

func (pow *PoWConsensus) miningLoop() {
    for pow.running {
        block, _ := pow.CreateBlock()
        nonce := uint64(rand.Intn(1e6))
        for {
            block.Header.Nonce = nonce
            hash := sha256.Sum256([]byte(block.Hash()))
            if hasLeadingZeros(hex.EncodeToString(hash[:]), pow.difficulty) {
                block.Header.Hash = hex.EncodeToString(hash[:])
                break
            }
            nonce++
        }
        pow.blockchain.AddBlock(block)
        time.Sleep(time.Second * 2)
    }
}

func hasLeadingZeros(hash string, difficulty int) bool {
    for i := 0; i < difficulty; i++ {
        if hash[i] != '0' {
            return false
        }
    }
    return true
}
