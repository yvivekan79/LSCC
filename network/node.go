
package network

import (
	"fmt"
	"net"
	"net/http"
	"lscc/config"
	"lscc/core"
	"lscc/utils"
	"lscc/consensus"
	"sync"
	"time"
)

// Node struct to represent a network node in the blockchain network.
type Node struct {
	Config     *config.Config
	Blockchain *core.Blockchain
	Logger     *utils.Logger
	consensus  core.Consensus
	mu         sync.Mutex
}

// NewNode creates a new network node.
func NewNode(cfg *config.Config, logger *utils.Logger) (*Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid port number: %d", cfg.Port)
	}
	if cfg.ShardID < 0 {
		return nil, fmt.Errorf("shard ID cannot be negative: %d", cfg.ShardID)
	}
	bc := core.NewBlockchain(cfg.ShardID, cfg.NodeID)
	node := &Node{
		Config:     cfg,
		Blockchain: bc,
		Logger:     logger,
	}

	var err error
	switch cfg.ConsensusType {
	case "pos":
		node.consensus, err = consensus.NewPoSConsensus(cfg, node.Blockchain)
	case "pow":
		node.consensus, err = consensus.NewPoWConsensus(cfg, node.Blockchain)
	case "pbft":
		node.consensus, err = consensus.NewPBFTConsensus(cfg, node.Blockchain)
	default:
		return nil, fmt.Errorf("unsupported consensus type: %s", cfg.ConsensusType)
	}
	if err != nil {
		return nil, err
	}
	return node, nil
}

// CreateBlock creates a new block with pending transactions.
func (n *Node) CreateBlock() *core.Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get pending transactions from mempool
	pendingTxs := n.Blockchain.GetPendingTransactions()

	// Limit transactions per block (e.g., max 10)
	maxTxsPerBlock := 10
	if len(pendingTxs) > maxTxsPerBlock {
		pendingTxs = pendingTxs[:maxTxsPerBlock]
	}

	// Get last block
	lastBlock := n.Blockchain.GetLastBlock()
	prevHash := ""
	height := uint64(1)

	if lastBlock != nil {
		prevHash = lastBlock.Hash
		height = lastBlock.Height + 1
	}

	// Create new block
	block := core.NewBlock(height, prevHash, pendingTxs, n.Config.NodeID, n.Config.ShardID)

	return block
}

// Start starts the network node.
func (n *Node) Start() error {
	n.Logger.Info("=== Starting Node ===")
	n.Logger.Info("Node configuration", "shardID", n.Config.ShardID, "layer", n.Config.Layer, "port", n.Config.Port)

	n.Logger.Info("Initializing blockchain...")
	n.Logger.Info("Genesis block already added during blockchain creation")

	// Create HTTP server with timeout configurations
	address := fmt.Sprintf("0.0.0.0:%d", n.Config.Port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		n.Logger.Error("Failed to bind to address", "address", address, "error", err)
		return err
	}
	n.Logger.Info("Successfully bound to address", "address", address)

	// Create router and log available endpoints
	router := n.router()
	n.Logger.Info("REST API endpoints configured:")
	n.Logger.Info("  GET  /status - Node status information")
	n.Logger.Info("  POST /send   - Submit new transaction")
	n.Logger.Info("  GET  /chain  - Get blockchain blocks")

	// Start the HTTP server in a separate goroutine
	n.Logger.Info("Starting HTTP server...")
	go func() {
		n.Logger.Info("HTTP server listening for requests", "address", address)
		if err := http.Serve(listener, router); err != nil {
			n.Logger.Error("HTTP server error", "error", err)
		}
	}()

	n.Logger.Info("=== Node Started Successfully ===")
	n.Logger.Info("Node details", "shardID", n.Config.ShardID, "layer", n.Config.Layer, "port", n.Config.Port)
	n.Logger.Info("Node is ready to accept transactions")
	n.Logger.Info("Access the node at:", "url", fmt.Sprintf("http://0.0.0.0:%d", n.Config.Port))

	return nil
}

// startBlockCreation periodically creates and adds new blocks to the blockchain.
func (n *Node) startBlockCreation() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get pending transactions
			pendingTxs := n.Blockchain.GetPendingTransactions()

			if len(pendingTxs) > 0 {
				// Create new block
				block := n.CreateBlock()

				// Add block to blockchain
				if err := n.Blockchain.AddBlock(block); err != nil {
					n.Logger.Error("Failed to add block", "error", err)
					continue
				}

				// Remove processed transactions from mempool
				for _, tx := range block.Transactions {
					n.Blockchain.RemoveFromMempool(tx.Hash)
				}

				n.Logger.Info("Block created and added",
					"height", block.Height,
					"hash", block.Hash,
					"transactions", len(block.Transactions))
			}
		}
	}
}
