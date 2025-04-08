package consensus

import (
        "errors"
        "math/rand"
        "sync"
        "time"

        "lscc/config"
        "lscc/core"
        "lscc/utils"
)

// PoSConsensus implements a simple Proof-of-Stake consensus
type PoSConsensus struct {
        blockchain      *core.Blockchain
        config          *config.Config
        validators      map[string]float64 // maps validator ID to stake
        running         bool
        stopChan        chan struct{}
        mu              sync.RWMutex
        pendingBlocks   map[string]*core.Block // blocks waiting for validation
        logger          *utils.Logger
        params          ConsensusParams
        lastBlockTime   time.Time
}

// NewPoSConsensus creates a new PoS consensus engine
func NewPoSConsensus(config *config.Config, blockchain *core.Blockchain) (*PoSConsensus, error) {
        return &PoSConsensus{
                blockchain:    blockchain,
                config:        config,
                validators:    make(map[string]float64),
                running:       false,
                stopChan:      make(chan struct{}),
                pendingBlocks: make(map[string]*core.Block),
                logger:        utils.GetLogger(),
                params: ConsensusParams{
                        BlockTime:        config.BlockTime,
                        MinConfirmations: config.MinConfirmations,
                        MinStake:         1000.0, // Minimum stake to be a validator
                        StakingReward:    5.0,    // Reward per block for validators
                },
                lastBlockTime: time.Now(),
        },
        nil
}

// Start starts the consensus engine
func (pos *PoSConsensus) Start() error {
        pos.mu.Lock()
        defer pos.mu.Unlock()

        if pos.running {
                return errors.New("consensus already running")
        }

        pos.running = true
        go pos.consensusLoop()

        pos.logger.Info("PoS consensus started")
        return nil
}

// Stop stops the consensus engine
func (pos *PoSConsensus) Stop() error {
        pos.mu.Lock()
        defer pos.mu.Unlock()

        if !pos.running {
                return errors.New("consensus not running")
        }

        close(pos.stopChan)
        pos.running = false

        pos.logger.Info("PoS consensus stopped")
        return nil
}

// consensusLoop is the main loop for the consensus engine
func (pos *PoSConsensus) consensusLoop() {
        ticker := time.NewTicker(time.Duration(pos.params.BlockTime) * time.Second)
        defer ticker.Stop()

        for {
                select {
                case <-pos.stopChan:
                        return
                case <-ticker.C:
                        // Check if it's our turn to create a block
                        if pos.isValidatorTurn(pos.config.NodeID) {
                                pos.logger.Info("It's our turn to create a block")
                                block, err := pos.CreateBlock()
                                if err != nil {
                                        pos.logger.Error("Failed to create block", "error", err)
                                        continue
                                }

                                // Process and broadcast the new block
                                err = pos.ProcessBlock(block)
                                if err != nil {
                                        pos.logger.Error("Failed to process block", "error", err)
                                        continue
                                }

                                // Reset the timer for the next block
                                pos.lastBlockTime = time.Now()
                        }
                }
        }
}

// isValidatorTurn checks if it's the validator's turn to create a block
func (pos *PoSConsensus) isValidatorTurn(validatorID string) bool {
        // In a real PoS, this would be determined by stake weight
        // For simplicity, we'll use a simple random selection algorithm
        
        // Get current validators
        validators := pos.getValidators()
        if len(validators) == 0 {
                return false
        }
        
        // For demo purposes, randomly select a validator based on current time
        rand.Seed(time.Now().UnixNano())
        selectedIndex := rand.Intn(len(validators))
        selectedValidator := validators[selectedIndex]
        
        return selectedValidator == validatorID
}

// getValidators returns a list of active validators
func (pos *PoSConsensus) getValidators() []string {
        pos.mu.RLock()
        defer pos.mu.RUnlock()
        
        validators := make([]string, 0, len(pos.validators))
        for validator := range pos.validators {
                validators = append(validators, validator)
        }
        
        // If no validators registered yet, use this node as default
        if len(validators) == 0 {
                validators = append(validators, pos.config.NodeID)
        }
        
        return validators
}

// CreateBlock creates a new block with pending transactions
func (pos *PoSConsensus) CreateBlock() (*core.Block, error) {
        // Get pending transactions
        pendingTxs := pos.blockchain.GetPendingTransactions()
        
        // Get the latest block
        latestBlock := pos.blockchain.GetLatestBlock()
        if latestBlock == nil {
                return nil, errors.New("no blocks in blockchain")
        }
        
        lastHash, err := latestBlock.Hash()
        if err != nil {
                return nil, err
        }
        
        // Create new block
        newBlock := core.NewBlock(
                lastHash,
                latestBlock.Header.Height+1,
                pos.config.ShardID,
                latestBlock.Header.Layer,
                pos.config.NodeID,
        )
        
        // Add transactions to the block (up to max limit)
        txCount := 0
        for _, tx := range pendingTxs {
                if txCount >= pos.config.MaxTransPerBlock {
                        break
                }
                
                // Skip cross-shard transactions for now, they need special handling
                if tx.IsCrossShard() {
                        continue
                }
                
                newBlock.AddTransaction(*tx)
                txCount++
        }
        
        // Add cross-shard references if any
        // In a real implementation, this would pull from other shards
        
        // Sign the block
        err = newBlock.Sign(pos.config.NodeID)
        if err != nil {
                return nil, err
        }
        
        return newBlock, nil
}

// ValidateBlock validates a block according to PoS rules
func (pos *PoSConsensus) ValidateBlock(block *core.Block) bool {
        // Check if the block is from a valid validator
        if !pos.isValidator(block.Header.ValidatorID) {
                pos.logger.Warn("Block from non-validator", "validator", block.Header.ValidatorID)
                return false
        }
        
        // Verify block signature
        if !block.VerifySignature(block.Header.ValidatorID) {
                pos.logger.Warn("Invalid block signature")
                return false
        }
        
        // Validate block structure and transactions
        latestBlock := pos.blockchain.GetLatestBlock()
        if !block.IsValid(latestBlock) {
                pos.logger.Warn("Invalid block structure")
                return false
        }
        
        // Validate all transactions in the block
        for _, tx := range block.Transactions {
                if !tx.IsValid() {
                        pos.logger.Warn("Invalid transaction in block", "txHash", tx.Hash)
                        return false
                }
        }
        
        return true
}

// isValidator checks if a node is a validator
func (pos *PoSConsensus) isValidator(nodeID string) bool {
        pos.mu.RLock()
        defer pos.mu.RUnlock()
        
        // For demo purposes, all nodes are validators
        return true
}

// ProcessBlock processes a new block and adds it to the blockchain
func (pos *PoSConsensus) ProcessBlock(block *core.Block) error {
        // Validate the block
        if !pos.ValidateBlock(block) {
                return errors.New("invalid block")
        }
        
        // Add the block to the blockchain
        err := pos.blockchain.AddBlock(block)
        if err != nil {
                return err
        }
        
        // Update last block time
        pos.lastBlockTime = time.Now()
        
        // Mark transactions as confirmed
        for _, tx := range block.Transactions {
                txCopy := tx
                txCopy.Confirm()
        }
        
        pos.logger.Info("Block processed successfully", 
                "height", block.Header.Height, 
                "validator", block.Header.ValidatorID,
                "transactions", len(block.Transactions))
        
        return nil
}

// GetType returns the type of consensus algorithm
func (pos *PoSConsensus) GetType() ConsensusType {
        return ProofOfStake
}

// GetStatus returns the current status of the consensus engine
func (pos *PoSConsensus) GetStatus() map[string]interface{} {
        pos.mu.RLock()
        defer pos.mu.RUnlock()
        
        validators := make([]string, 0, len(pos.validators))
        for validator := range pos.validators {
                validators = append(validators, validator)
        }
        
        return map[string]interface{}{
                "type":              string(ProofOfStake),
                "running":           pos.running,
                "validators":        validators,
                "validator_count":   len(validators),
                "last_block_time":   pos.lastBlockTime,
                "pending_blocks":    len(pos.pendingBlocks),
                "block_time":        pos.params.BlockTime,
                "min_confirmations": pos.params.MinConfirmations,
        }
}

// HandleConsensusMessage handles consensus-specific messages
func (pos *PoSConsensus) HandleConsensusMessage(message []byte) error {
        // In a real implementation, this would parse and handle different consensus messages
        // For now, we'll just return nil
        return nil
}

// RegisterValidator registers a new validator with the specified stake
func (pos *PoSConsensus) RegisterValidator(nodeID string, stake float64) error {
        pos.mu.Lock()
        defer pos.mu.Unlock()
        
        if stake < pos.params.MinStake {
                return errors.New("stake amount below minimum required")
        }
        
        pos.validators[nodeID] = stake
        pos.logger.Info("New validator registered", "nodeID", nodeID, "stake", stake)
        
        return nil
}

// UnregisterValidator removes a validator
func (pos *PoSConsensus) UnregisterValidator(nodeID string) {
        pos.mu.Lock()
        defer pos.mu.Unlock()
        
        delete(pos.validators, nodeID)
        pos.logger.Info("Validator unregistered", "nodeID", nodeID)
}
