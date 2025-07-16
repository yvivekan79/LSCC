package main

import (
    "flag"
    "lscc/config"
    "lscc/network"
    "lscc/utils"
)

func main() {
    configPath := flag.String("config", "config/config.json", "Path to config file")
    flag.Parse()

    cfg, err := config.LoadConfig(*configPath)
    if err != nil {
        panic(err)
    }

    logger := utils.InitLoggerLevel(cfg.LoggingLevel)
    node, err := network.NewNode(cfg, logger)
    if err != nil {
        logger.Error("Failed to create node", "error", err)
        return
    }

    err = node.Start()
    if err != nil {
        logger.Error("Node failed to start", "error", err)
    }
    logger.Info("Node started successfully", "shardID", cfg.ShardID, "layer", cfg.Layer)
    select {} // Keep the main goroutine running    
}

