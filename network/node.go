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
    n.Logger.Info("Starting node...")
    n.Logger.Info(fmt.Sprintf("Trying to bind HTTP server to :%d", n.Config.Port))

    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.Config.Port))
    if err != nil {
		 n.Logger.Error(fmt.Sprintf("Failed to bind to port %d: %v", n.Config.Port, err))
        return err
    }
    n.Logger.Info(fmt.Sprintf("Successfully bound to port %d", n.Config.Port))  
    n.Logger.Info("Starting HTTP server...")
    // Start the HTTP server in a separate goroutine

    go func() {
        if err := http.Serve(listener, n.router()); err != nil {
            n.Logger.Error("Failed to start HTTP server", "error", err)
        }
    }()
    n.Logger.Info(fmt.Sprintf("Node started on port %d", n.Config.Port))
    n.Logger.Info(fmt.Sprintf("Shard ID: %d, Layer: %d", n.Config.ShardID, n.Config.Layer))
    n.Logger.Info("Node is ready to accept transactions")       
    // Initialize the blockchain with the genesis block
    // Adjust the parameters below to match the actual NewBlock function signature in your core package
    genesisBlock := core.NewBlock(fmt.Sprintf("%d", "0"), []*core.Transaction{}, n.Config.ShardID, n.Config.Layer)
    n.Blockchain.AddBlock(genesisBlock) 
    n.Logger.Info("Genesis block added to the blockchain", "hash", genesisBlock.Hash)           
    // Start the REST API server
    n.Logger.Info("REST API server is running", "port", n.Config.Port)
        

    return nil
}

