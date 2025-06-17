
package consensus

import (
    "crypto/sha256"
    "encoding/hex"
    "math/rand"
    "sync"
    "time"

    "lscc/config"
    "lscc/core"
    "lscc/metrics"
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
    recorder   *metrics.Recorder
}

func NewPoWConsensus(cfg *config.Config, bc *core.Blockchain, recorder *metrics.Recorder) *PoWConsensus {
    return &PoWConsensus{
        blockchain: bc,
        config:     cfg,
        difficulty: 4,
        stopChan:   make(chan struct{}),
        logger:     utils.GetLogger(),
        recorder:   recorder,
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
    txs := pow.blockchain.TxPool.GetAll()
    prevBlock := pow.blockchain.GetLatestBlock()
    block := core.NewBlock(prevBlock, txs, pow.config.NodeID)
    return block, nil
}

func (pow *PoWConsensus) ValidateBlock(block *core.Block) bool {
    hash := sha256.Sum256([]byte([]byte(block.Hash())))
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
    return nil
}

func (pow *PoWConsensus) miningLoop() {
    threadCount := 4
    var wg sync.WaitGroup
    for i := 0; i < threadCount; i++ {
        wg.Add(1)
        go func(threadID int) {
            defer wg.Done()
            for pow.running {
                start := time.Now()
                block, _ := pow.CreateBlock()
                mined := pow.mineBlock(block)
                if mined {
                    pow.blockchain.AddBlock(block)
                    pow.recorder.Record("pow", time.Since(start))
                }
                time.Sleep(1 * time.Second)
            }
        }(i)
    }
    wg.Wait()
}

func (pow *PoWConsensus) mineBlock(block *core.Block) bool {
    for nonce := uint64(0); nonce < 1e6; nonce++ {
        block.Header.Nonce = nonce
        hash := sha256.Sum256([]byte([]byte(block.Hash())))
        if hasLeadingZeros(hex.EncodeToString(hash[:]), pow.difficulty) {
            block.Header.Hash = hex.EncodeToString(hash[:])
            return true
        }
    }
    return false
}

func hasLeadingZeros(hash string, difficulty int) bool {
    for i := 0; i < difficulty; i++ {
        if hash[i] != '0' {
            return false
        }
    }
    return true
}