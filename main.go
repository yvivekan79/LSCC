package main

import (
        "flag"
        "fmt"
        "os"
        "os/signal"
        "syscall"

        "lscc/config"
        "lscc/network"
        "lscc/sharding"
        "lscc/utils"
)

// Command line flags
var (
        configFile  = flag.String("config", "config.json", "Path to configuration file")
        nodeID      = flag.String("nodeid", "", "Node ID (if empty, will be randomly generated)")
        port        = flag.Int("port", 8000, "Port to listen on")
        bootstrapIP = flag.String("bootstrap", "", "Bootstrap node IP:port")
        isRelay     = flag.Bool("relay", false, "Run as a relay node")
        shardID     = flag.Int("shard", -1, "Shard ID (-1 for automatic assignment)")
        verbosity   = flag.Int("verbosity", 3, "Log verbosity (0-5)")
)

func main() {
        flag.Parse()

        // Initialize logger
        utils.InitLogger(*verbosity)
        logger := utils.GetLogger()
        logger.Info("Starting LSCC Node...")

        // Load configuration
        cfg, err := config.LoadConfig(*configFile)
        if err != nil {
                if os.IsNotExist(err) {
                        logger.Info("Config file not found, using defaults and command line arguments")
                        cfg = config.DefaultConfig()
                } else {
                        logger.Error("Failed to load config", "error", err)
                        os.Exit(1)
                }
        }

        // Override config with command line args
        if *nodeID != "" {
                cfg.NodeID = *nodeID
        }
        if *port != 8000 {
                cfg.Port = *port
        }
        if *bootstrapIP != "" {
                cfg.BootstrapNodes = []string{*bootstrapIP}
        }
        if *isRelay {
                cfg.IsRelay = true
        }
        if *shardID != -1 {
                cfg.ShardID = *shardID
        }

        // Generate node ID if not provided
        if cfg.NodeID == "" {
                id, err := utils.GenerateNodeID()
                if err != nil {
                        logger.Error("Failed to generate node ID", "error", err)
                        os.Exit(1)
                }
                cfg.NodeID = id
                logger.Info("Generated node ID", "id", id)
        }

        // Create sharding manager
        shardManager := sharding.NewManager(cfg)

        // Initialize and start the node
        node, err := network.NewNode(cfg, shardManager)
        if err != nil {
                logger.Error("Failed to create node", "error", err)
                os.Exit(1)
        }

        err = node.Start()
        if err != nil {
                logger.Error("Failed to start node", "error", err)
                os.Exit(1)
        }

        // Handle graceful shutdown
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
        <-c

        fmt.Println("\nShutting down...")
        node.Stop()
        logger.Info("Node stopped")
}
