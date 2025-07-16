
package consensus

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "lscc/config"
    "lscc/core"
    "lscc/utils"
    "strconv"
    "strings"
    "sync"
    "time"
)

type PoWConsensus struct {
    blockchain  *core.Blockchain
    nodeID      string
    difficulty  int
    target      string
    mining      bool
    mu          sync.RWMutex
    logger      *utils.Logger
    config      *config.Config
}

func NewPoWConsensus(cfg *config.Config, blockchain *core.Blockchain) (*PoWConsensus, error) {
    logger := utils.InitLoggerLevel(cfg.LoggingLevel)
    
    difficulty := cfg.ConsensusParams.Difficulty
    if difficulty == 0 {
        difficulty = 2 // Default difficulty
    }
    
    target := strings.Repeat("0", difficulty)
    
    return &PoWConsensus{
        blockchain: blockchain,
        nodeID:     cfg.NodeID,
        difficulty: difficulty,
        target:     target,
        mining:     false,
        logger:     logger,
        config:     cfg,
    }, nil
}

func (p *PoWConsensus) Start() error {
    p.logger.Info("Starting PoW consensus", "node", p.nodeID, "difficulty", p.difficulty)
    
    go p.miningLoop()
    
    return nil
}

func (p *PoWConsensus) Stop() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.mining = false
    p.logger.Info("Stopping PoW consensus")
    return nil
}

func (p *PoWConsensus) CreateBlock() (*core.Block, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    transactions := p.blockchain.GetPendingTransactions()
    if len(transactions) == 0 {
        return nil, fmt.Errorf("no pending transactions")
    }
    
    lastBlock := p.blockchain.GetLastBlock()
    height := uint64(0)
    prevHash := ""
    
    if lastBlock != nil {
        height = lastBlock.Height + 1
        prevHash = lastBlock.Hash
    }
    
    block := &core.Block{
        Height:        height,
        Timestamp:     time.Now().Unix(),
        PrevBlockHash: prevHash,
        Transactions:  transactions[:min(len(transactions), 10)],
        Validator:     p.nodeID,
        ShardID:       p.config.ShardID,
    }
    
    // Mine the block
    nonce, hash := p.mineBlock(block)
    block.Nonce = nonce
    block.Hash = hash
    
    p.logger.Info("PoW block mined", "height", block.Height, "hash", block.Hash, "nonce", nonce, "txCount", len(block.Transactions))
    
    return block, nil
}

func (p *PoWConsensus) mineBlock(block *core.Block) (uint64, string) {
    var nonce uint64 = 0
    var hash string
    
    startTime := time.Now()
    
    for {
        blockData := fmt.Sprintf("%d:%d:%s:%s:%d", block.Height, block.Timestamp, block.PrevBlockHash, block.Validator, nonce)
        hashBytes := sha256.Sum256([]byte(blockData))
        hash = hex.EncodeToString(hashBytes[:])
        
        if strings.HasPrefix(hash, p.target) {
            duration := time.Since(startTime)
            p.logger.Info("PoW mining completed", "nonce", nonce, "hash", hash, "duration", duration)
            break
        }
        
        nonce++
        
        // Check if we should stop mining
        p.mu.RLock()
        if !p.mining {
            p.mu.RUnlock()
            break
        }
        p.mu.RUnlock()
        
        // Prevent infinite mining in case of high difficulty
        if nonce%100000 == 0 {
            p.logger.Debug("PoW mining progress", "nonce", nonce, "target", p.target)
        }
    }
    
    return nonce, hash
}

func (p *PoWConsensus) ValidateBlock(block *core.Block) bool {
    p.logger.Debug("PoW validating block", "height", block.Height, "hash", block.Hash)
    
    // Validate block structure
    if block.Height == 0 {
        return true // Genesis block
    }
    
    lastBlock := p.blockchain.GetLastBlock()
    if lastBlock == nil && block.Height != 0 {
        p.logger.Error("No genesis block found")
        return false
    }
    
    if lastBlock != nil && block.Height != lastBlock.Height+1 {
        p.logger.Error("Invalid block height", "expected", lastBlock.Height+1, "got", block.Height)
        return false
    }
    
    if lastBlock != nil && block.PrevBlockHash != lastBlock.Hash {
        p.logger.Error("Invalid previous block hash")
        return false
    }
    
    // Validate proof of work
    if !p.validateProofOfWork(block) {
        p.logger.Error("Invalid proof of work", "hash", block.Hash, "target", p.target)
        return false
    }
    
    return true
}

func (p *PoWConsensus) validateProofOfWork(block *core.Block) bool {
    blockData := fmt.Sprintf("%d:%d:%s:%s:%d", block.Height, block.Timestamp, block.PrevBlockHash, block.Validator, block.Nonce)
    hashBytes := sha256.Sum256([]byte(blockData))
    calculatedHash := hex.EncodeToString(hashBytes[:])
    
    return calculatedHash == block.Hash && strings.HasPrefix(block.Hash, p.target)
}

func (p *PoWConsensus) ProcessBlock(block *core.Block) error {
    if !p.ValidateBlock(block) {
        return fmt.Errorf("block validation failed")
    }
    
    err := p.blockchain.AddBlock(block)
    if err != nil {
        return fmt.Errorf("failed to add block to blockchain: %v", err)
    }
    
    // Remove processed transactions from mempool
    for _, tx := range block.Transactions {
        p.blockchain.RemoveFromMempool(tx.Hash)
    }
    
    p.logger.Info("PoW block processed and added to blockchain", "height", block.Height, "hash", block.Hash)
    
    return nil
}

func (p *PoWConsensus) miningLoop() {
    p.mu.Lock()
    p.mining = true
    p.mu.Unlock()
    
    ticker := time.NewTicker(5 * time.Second) // Check for new transactions every 5 seconds
    defer ticker.Stop()
    
    for range ticker.C {
        p.mu.RLock()
        if !p.mining {
            p.mu.RUnlock()
            break
        }
        p.mu.RUnlock()
        
        block, err := p.CreateBlock()
        if err != nil {
            continue // No transactions to mine
        }
        
        err = p.ProcessBlock(block)
        if err != nil {
            p.logger.Error("Failed to process mined block", "error", err)
        }
    }
}

func (p *PoWConsensus) GetType() string {
    return "pow"
}

func (p *PoWConsensus) GetStatus() map[string]interface{} {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    return map[string]interface{}{
        "type":       "pow",
        "node_id":    p.nodeID,
        "difficulty": p.difficulty,
        "target":     p.target,
        "mining":     p.mining,
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
