package consensus

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "lscc/config"
    "lscc/core"
    "lscc/utils"
    "sync"
    "time"
)

type PBFTPhase int

const (
    PrePrepare PBFTPhase = iota
    Prepare
    Commit
)

type PBFTMessage struct {
    Phase       PBFTPhase `json:"phase"`
    View        int       `json:"view"`
    SequenceNum int       `json:"sequence_num"`
    BlockHash   string    `json:"block_hash"`
    NodeID      string    `json:"node_id"`
    Timestamp   int64     `json:"timestamp"`
}

type PBFTConsensus struct {
    blockchain    *core.Blockchain
    nodeID        string
    validators    []string
    currentView   int
    sequenceNum   int
    isPrimary     bool
    messages      map[string]map[PBFTPhase][]*PBFTMessage
    committedBlocks map[string]bool
    mu            sync.RWMutex
    logger        *utils.Logger
    config        *config.Config
    pendingBlock  *core.Block
}

func NewPBFTConsensus(cfg *config.Config, blockchain *core.Blockchain) (*PBFTConsensus, error) {
    logger := utils.InitLoggerLevel(cfg.LoggingLevel)

    validators := []string{cfg.NodeID}
    if len(cfg.BootstrapNodes) > 0 {
        for _, node := range cfg.BootstrapNodes {
            validators = append(validators, fmt.Sprintf("node_%s", node))
        }
    }

    return &PBFTConsensus{
        blockchain:      blockchain,
        nodeID:          cfg.NodeID,
        validators:      validators,
        currentView:     0,
        sequenceNum:     0,
        isPrimary:       cfg.NodeID == validators[0],
        messages:        make(map[string]map[PBFTPhase][]*PBFTMessage),
        committedBlocks: make(map[string]bool),
        logger:          logger,
        config:          cfg,
    }, nil
}

func (p *PBFTConsensus) Start() error {
    p.logger.Info("Starting PBFT consensus", "node", p.nodeID, "isPrimary", p.isPrimary)

    if p.isPrimary {
        go p.primaryLoop()
    }

    return nil
}

func (p *PBFTConsensus) Stop() error {
    p.logger.Info("Stopping PBFT consensus")
    return nil
}

func (p *PBFTConsensus) CreateBlock() (*core.Block, error) {
    if !p.isPrimary {
        return nil, fmt.Errorf("only primary can create blocks")
    }

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

    blockData := fmt.Sprintf("%d:%d:%s:%s", block.Height, block.Timestamp, block.PrevBlockHash, block.Validator)
    hash := sha256.Sum256([]byte(blockData))
    block.Hash = hex.EncodeToString(hash[:])

    p.logger.Info("PBFT block created", "height", block.Height, "hash", block.Hash, "txCount", len(block.Transactions))

    return block, nil
}

func (p *PBFTConsensus) ValidateBlock(block *core.Block) bool {
    p.logger.Debug("PBFT validating block", "height", block.Height, "hash", block.Hash)

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

    return true
}

func (p *PBFTConsensus) ProcessBlock(block *core.Block) error {
    if !p.ValidateBlock(block) {
        return fmt.Errorf("block validation failed")
    }

    // Initiate PBFT consensus for this block
    return p.initiateConsensus(block)
}

func (p *PBFTConsensus) initiateConsensus(block *core.Block) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.pendingBlock = block
    p.sequenceNum++

    // Pre-prepare phase
    msg := &PBFTMessage{
        Phase:       PrePrepare,
        View:        p.currentView,
        SequenceNum: p.sequenceNum,
        BlockHash:   block.Hash,
        NodeID:      p.nodeID,
        Timestamp:   time.Now().Unix(),
    }

    p.logger.Info("PBFT Pre-prepare phase initiated", "blockHash", block.Hash, "sequence", p.sequenceNum)

    // Store our own pre-prepare message
    p.storeMessage(msg)

    // Broadcast to other validators (simulation)
    go p.simulateValidatorResponses(block.Hash)

    return nil
}

func (p *PBFTConsensus) storeMessage(msg *PBFTMessage) {
    if p.messages[msg.BlockHash] == nil {
        p.messages[msg.BlockHash] = make(map[PBFTPhase][]*PBFTMessage)
    }

    p.messages[msg.BlockHash][msg.Phase] = append(p.messages[msg.BlockHash][msg.Phase], msg)

    // Check if we have enough messages to proceed
    p.checkConsensusProgress(msg.BlockHash)
}

func (p *PBFTConsensus) checkConsensusProgress(blockHash string) {
    requiredVotes := (2 * len(p.validators)) / 3 + 1

    blockMsgs := p.messages[blockHash]

    // Check prepare phase
    if len(blockMsgs[PrePrepare]) >= 1 && len(blockMsgs[Prepare]) >= requiredVotes-1 {
        // Move to commit phase
        commitMsg := &PBFTMessage{
            Phase:       Commit,
            View:        p.currentView,
            SequenceNum: p.sequenceNum,
            BlockHash:   blockHash,
            NodeID:      p.nodeID,
            Timestamp:   time.Now().Unix(),
        }
        p.storeMessage(commitMsg)
        p.logger.Info("PBFT Commit phase", "blockHash", blockHash)
    }

    // Check commit phase
    if len(blockMsgs[Commit]) >= requiredVotes {
        // Consensus reached
        p.finalizeBlock(blockHash)
    }
}

func (p *PBFTConsensus) finalizeBlock(blockHash string) {
    if p.committedBlocks[blockHash] {
        return // Already committed
    }

    if p.pendingBlock != nil && p.pendingBlock.Hash == blockHash {
        err := p.blockchain.AddBlock(p.pendingBlock)
        if err != nil {
            p.logger.Error("Failed to add block to blockchain", "error", err)
            return
        }

        p.committedBlocks[blockHash] = true
        p.logger.Info("PBFT consensus reached - block committed", "blockHash", blockHash, "height", p.pendingBlock.Height)

        // Remove processed transactions from mempool
        for _, tx := range p.pendingBlock.Transactions {
            p.blockchain.RemoveFromMempool(tx.Hash)
        }

        p.pendingBlock = nil
    }
}

func (p *PBFTConsensus) simulateValidatorResponses(blockHash string) {
    time.Sleep(100 * time.Millisecond)

    // Simulate prepare messages from other validators
    for i, validator := range p.validators {
        if validator == p.nodeID {
            continue
        }

        time.Sleep(time.Duration(i*50) * time.Millisecond)

        prepareMsg := &PBFTMessage{
            Phase:       Prepare,
            View:        p.currentView,
            SequenceNum: p.sequenceNum,
            BlockHash:   blockHash,
            NodeID:      validator,
            Timestamp:   time.Now().Unix(),
        }

        p.mu.Lock()
        p.storeMessage(prepareMsg)
        p.mu.Unlock()

        p.logger.Debug("Simulated PBFT prepare message", "from", validator, "blockHash", blockHash)
    }
}

func (p *PBFTConsensus) primaryLoop() {
    ticker := time.NewTicker(time.Duration(p.config.ConsensusParams.Difficulty) * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        block, err := p.CreateBlock()
        if err != nil {
            continue
        }

        err = p.ProcessBlock(block)
        if err != nil {
            p.logger.Error("Failed to process block in primary loop", "error", err)
        }
    }
}

func (p *PBFTConsensus) GetType() string {
    return "pbft"
}

func (p *PBFTConsensus) GetStatus() map[string]interface{} {
    p.mu.RLock()
    defer p.mu.RUnlock()

    return map[string]interface{}{
        "type":            "pbft",
        "node_id":         p.nodeID,
        "is_primary":      p.isPrimary,
        "current_view":    p.currentView,
        "sequence_num":    p.sequenceNum,
        "validators":      p.validators,
        "pending_block":   p.pendingBlock != nil,
        "committed_count": len(p.committedBlocks),
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}