package consensus

import (
        "lscc/config"
        "lscc/core"
        "lscc/utils"
)

// ConsensusType represents the type of consensus algorithm
type ConsensusType string

const (
        // ProofOfWork consensus algorithm
        ProofOfWork ConsensusType = "pow"
        // ProofOfStake consensus algorithm
        ProofOfStake ConsensusType = "pos"
        // PracticalByzantineFaultTolerance consensus algorithm
        PBFT ConsensusType = "pbft"
        // Cross-Channel Consensus - our hybrid approach
        CrossChannel ConsensusType = "cross-channel"
)

// ConsensusEngine is the interface for consensus algorithms
type ConsensusEngine interface {
        // Start starts the consensus engine
        Start() error
        
        // Stop stops the consensus engine
        Stop() error
        
        // CreateBlock creates a new block with pending transactions
        CreateBlock() (*core.Block, error)
        
        // ValidateBlock validates a block according to the consensus rules
        ValidateBlock(block *core.Block) bool
        
        // ProcessBlock processes a new block and adds it to the blockchain
        ProcessBlock(block *core.Block) error
        
        // GetType returns the type of consensus algorithm
        GetType() ConsensusType
        
        // GetStatus returns the current status of the consensus engine
        GetStatus() map[string]interface{}
        
        // HandleConsensusMessage handles consensus-specific messages
        HandleConsensusMessage(message []byte) error
}

// NewConsensusEngine creates a new consensus engine based on the config
func NewConsensusEngine(config *config.Config, blockchain *core.Blockchain) (ConsensusEngine, error) {
        logger := utils.GetLogger()
        
        switch ConsensusType(config.ConsensusType) {
        case ProofOfStake:
                logger.Info("Initializing Proof of Stake consensus engine")
                return NewPoSConsensus(config, blockchain)
        case CrossChannel:
                logger.Info("Initializing Cross-Channel consensus engine")
                // Implement Cross-Channel consensus
                return nil, utils.NewError("Cross-Channel consensus not implemented yet")
        case ProofOfWork:
                logger.Info("Initializing Proof of Work consensus engine")
                // Implement PoW consensus
                return nil, utils.NewError("Proof of Work consensus not implemented yet")
        case PBFT:
                logger.Info("Initializing PBFT consensus engine")
                // Implement PBFT consensus
                return nil, utils.NewError("PBFT consensus not implemented yet")
        default:
                logger.Warn("Unknown consensus type, defaulting to PoS", "type", config.ConsensusType)
                return NewPoSConsensus(config, blockchain)
        }
}

// ConsensusParams holds parameters for the consensus algorithm
type ConsensusParams struct {
        // Common parameters for all consensus types
        BlockTime        int // Block time in seconds
        MinConfirmations int // Minimum confirmations required
        
        // PoS specific parameters
        MinStake         float64 // Minimum stake required for validators
        StakingReward    float64 // Reward for staking
        
        // PBFT specific parameters
        Validators       []string // List of validator node IDs
        ViewChangeTimeout int     // Timeout for view change in PBFT
        
        // Cross-Channel specific parameters
        CrossChannelVerify bool  // Enable cross-channel verification
        LayerCount         int   // Number of layers in the network
        CrossLayerTimeout  int   // Timeout for cross-layer communication
}

// DefaultConsensusParams returns default consensus parameters
func DefaultConsensusParams() ConsensusParams {
        return ConsensusParams{
                BlockTime:        5,
                MinConfirmations: 6,
                MinStake:         1000.0,
                StakingReward:    5.0,
                Validators:       []string{},
                ViewChangeTimeout: 30,
                CrossChannelVerify: true,
                LayerCount:         3,
                CrossLayerTimeout:  15,
        }
}
