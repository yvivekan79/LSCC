package network

import (
    "fmt"
    "net"
    "net/http"
    "lscc/config"
    "lscc/core"
    "lscc/utils"
)

type Node struct {
    Config     *config.Config
    Blockchain *core.Blockchain
    Logger     *utils.Logger
}

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
    bc := core.NewBlockchain(logger)
    return &Node{
        Config:     cfg,
        Blockchain: bc,
        Logger:     logger,
    }, nil
}

func (n *Node) Start() error {
    n.Logger.Info("=== Starting Node ===")
    n.Logger.Info("Node configuration", "shardID", n.Config.ShardID, "layer", n.Config.Layer, "port", n.Config.Port)
    
    // Initialize the blockchain with the genesis block first
    n.Logger.Info("Initializing blockchain...")
    genesisBlock := core.NewBlock(fmt.Sprintf("%d", "0"), []*core.Transaction{}, n.Config.ShardID, n.Config.Layer)
    n.Blockchain.AddBlock(genesisBlock) 
    n.Logger.Info("Genesis block added to the blockchain", "hash", genesisBlock.Hash)
    
    // Bind to 0.0.0.0 to make it accessible externally
    address := fmt.Sprintf("0.0.0.0:%d", n.Config.Port)
    n.Logger.Info("Attempting to bind HTTP server", "address", address)

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

