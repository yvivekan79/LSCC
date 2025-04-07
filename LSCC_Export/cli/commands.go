package cli

import (
        "encoding/json"
        "flag"
        "fmt"
        "os"
        "strconv"
        "strings"

        "lscc/config"
        "lscc/core"
        "lscc/network"
        "lscc/sharding"
        "lscc/utils"
)

// CLI represents the command-line interface
type CLI struct {
        node         *network.Node
        shardManager *sharding.Manager
        logger       *utils.Logger
}

// NewCLI creates a new CLI instance
func NewCLI(node *network.Node, shardManager *sharding.Manager) *CLI {
        return &CLI{
                node:         node,
                shardManager: shardManager,
                logger:       utils.GetLogger(),
        }
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
        // Define command-line flags
        statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
        
        createTxCmd := flag.NewFlagSet("createtx", flag.ExitOnError)
        createTxFrom := createTxCmd.String("from", "", "Sender address")
        createTxTo := createTxCmd.String("to", "", "Recipient address")
        createTxAmount := createTxCmd.Float64("amount", 0.0, "Amount to send")
        createTxFee := createTxCmd.Float64("fee", 0.001, "Transaction fee")
        createTxShard := createTxCmd.Int("shard", -1, "Target shard (default: auto-assign)")
        
        getBlockCmd := flag.NewFlagSet("getblock", flag.ExitOnError)
        getBlockHeight := getBlockCmd.Uint64("height", 0, "Block height")
        getBlockHash := getBlockCmd.String("hash", "", "Block hash")
        
        configCmd := flag.NewFlagSet("config", flag.ExitOnError)
        configShow := configCmd.Bool("show", false, "Show current configuration")
        configSet := configCmd.String("set", "", "Set configuration value (format: key=value)")
        configSave := configCmd.String("save", "", "Save configuration to file")
        
        peersCmd := flag.NewFlagSet("peers", flag.ExitOnError)
        
        shardsCmd := flag.NewFlagSet("shards", flag.ExitOnError)
        
        // Check command
        if len(os.Args) < 2 {
                cli.printUsage()
                os.Exit(1)
        }
        
        // Process commands
        switch os.Args[1] {
        case "status":
                statusCmd.Parse(os.Args[2:])
                cli.showStatus()
                
        case "createtx":
                createTxCmd.Parse(os.Args[2:])
                cli.createTransaction(*createTxFrom, *createTxTo, *createTxAmount, *createTxFee, *createTxShard)
                
        case "getblock":
                getBlockCmd.Parse(os.Args[2:])
                cli.getBlock(*getBlockHeight, *getBlockHash)
                
        case "config":
                configCmd.Parse(os.Args[2:])
                cli.handleConfig(*configShow, *configSet, *configSave)
                
        case "peers":
                peersCmd.Parse(os.Args[2:])
                cli.showPeers()
                
        case "shards":
                shardsCmd.Parse(os.Args[2:])
                cli.showShards()
                
        case "help":
                cli.printUsage()
                
        default:
                fmt.Println("Unknown command")
                cli.printUsage()
                os.Exit(1)
        }
}

// printUsage prints the command-line usage
func (cli *CLI) printUsage() {
        fmt.Println("Usage:")
        fmt.Println("  help                - Show this help message")
        fmt.Println("  status              - Show node status")
        fmt.Println("  createtx -from ADDR -to ADDR -amount AMT [-fee FEE] [-shard ID] - Create a transaction")
        fmt.Println("  getblock -height N or -hash HASH - Get block information")
        fmt.Println("  config -show        - Show current configuration")
        fmt.Println("  config -set KEY=VAL - Set configuration value")
        fmt.Println("  config -save FILE   - Save configuration to file")
        fmt.Println("  peers               - Show connected peers")
        fmt.Println("  shards              - Show shard information")
}

// showStatus displays the current node status
func (cli *CLI) showStatus() {
        status := cli.node.GetStatus()
        jsonBytes, err := json.MarshalIndent(status, "", "  ")
        if err != nil {
                cli.logger.Error("Failed to marshal status", "error", err)
                return
        }
        
        fmt.Println("Node Status:")
        fmt.Println(string(jsonBytes))
}

// createTransaction creates a new transaction
func (cli *CLI) createTransaction(from, to string, amount, fee float64, targetShard int) {
        if from == "" || to == "" || amount <= 0 {
                fmt.Println("Error: Sender, recipient, and amount are required")
                return
        }
        
        // Get the node's shard ID
        sourceShard := cli.node.Config.ShardID
        
        // If target shard not specified, auto-assign
        if targetShard == -1 {
                // Use the same shard as source for simplicity
                targetShard = sourceShard
        }
        
        // Determine transaction type
        txType := core.RegularTransaction
        if sourceShard != targetShard {
                txType = core.CrossShardTransaction
        }
        
        // Create the transaction
        tx, err := core.NewTransaction(from, to, amount, fee, sourceShard, targetShard, 0, txType)
        if err != nil {
                cli.logger.Error("Failed to create transaction", "error", err)
                fmt.Println("Error creating transaction:", err)
                return
        }
        
        // Sign the transaction (in a real implementation, this would use actual signing)
        err = tx.Sign(from)
        if err != nil {
                cli.logger.Error("Failed to sign transaction", "error", err)
                fmt.Println("Error signing transaction:", err)
                return
        }
        
        // Process the transaction
        if tx.IsCrossShard() {
                err = cli.shardManager.ProcessCrossShardTransaction(tx)
        } else {
                err = cli.node.Blockchain.AddTransaction(tx)
        }
        
        if err != nil {
                cli.logger.Error("Failed to process transaction", "error", err)
                fmt.Println("Error processing transaction:", err)
                return
        }
        
        // Broadcast the transaction
        cli.node.BroadcastTransaction(tx)
        
        fmt.Printf("Transaction created and broadcast: %s\n", tx.Hash)
        if tx.IsCrossShard() {
                fmt.Printf("Cross-shard transaction: Shard %d -> Shard %d\n", sourceShard, targetShard)
        }
}

// getBlock retrieves and displays block information
func (cli *CLI) getBlock(height uint64, hash string) {
        var block *core.Block
        
        if hash != "" {
                block = cli.node.Blockchain.GetBlockByHash(hash)
        } else if height > 0 {
                block = cli.node.Blockchain.GetBlockByHeight(height)
        } else {
                fmt.Println("Error: Either block height or hash must be specified")
                return
        }
        
        if block == nil {
                fmt.Println("Block not found")
                return
        }
        
        jsonBytes, err := json.MarshalIndent(block, "", "  ")
        if err != nil {
                cli.logger.Error("Failed to marshal block", "error", err)
                fmt.Println("Error formatting block:", err)
                return
        }
        
        fmt.Println("Block Information:")
        fmt.Println(string(jsonBytes))
}

// handleConfig handles configuration commands
func (cli *CLI) handleConfig(show bool, set string, save string) {
        if show {
                jsonBytes, err := json.MarshalIndent(cli.node.Config, "", "  ")
                if err != nil {
                        cli.logger.Error("Failed to marshal config", "error", err)
                        fmt.Println("Error formatting configuration:", err)
                        return
                }
                
                fmt.Println("Current Configuration:")
                fmt.Println(string(jsonBytes))
        }
        
        if set != "" {
                parts := strings.SplitN(set, "=", 2)
                if len(parts) != 2 {
                        fmt.Println("Error: Configuration must be in format key=value")
                        return
                }
                
                key, value := parts[0], parts[1]
                
                // Convert config to map for easier manipulation
                configMap := make(map[string]interface{})
                configBytes, _ := json.Marshal(cli.node.Config)
                json.Unmarshal(configBytes, &configMap)
                
                // Convert value to appropriate type if needed
                if intVal, err := strconv.Atoi(value); err == nil {
                        configMap[key] = intVal
                } else if boolVal, err := strconv.ParseBool(value); err == nil {
                        configMap[key] = boolVal
                } else {
                        configMap[key] = value
                }
                
                // Convert back to config
                configBytes, _ = json.Marshal(configMap)
                json.Unmarshal(configBytes, cli.node.Config)
                
                fmt.Printf("Set %s = %s\n", key, value)
        }
        
        if save != "" {
                err := config.SaveConfig(cli.node.Config, save)
                if err != nil {
                        cli.logger.Error("Failed to save config", "error", err)
                        fmt.Println("Error saving configuration:", err)
                        return
                }
                
                fmt.Printf("Configuration saved to %s\n", save)
        }
}

// showPeers displays information about connected peers
func (cli *CLI) showPeers() {
        peers := cli.node.GetPeers()
        
        fmt.Printf("Connected Peers (%d):\n", len(peers))
        fmt.Println("-----------------------------")
        
        for i, peer := range peers {
                fmt.Printf("%d. ID: %s\n   Address: %s\n   Connected: %v\n\n", 
                        i+1, peer.ID, peer.Address, peer.IsConnected())
        }
}

// showShards displays information about shards
func (cli *CLI) showShards() {
        status := cli.shardManager.GetStatus()
        
        jsonBytes, err := json.MarshalIndent(status, "", "  ")
        if err != nil {
                cli.logger.Error("Failed to marshal shard status", "error", err)
                fmt.Println("Error formatting shard information:", err)
                return
        }
        
        fmt.Println("Shard Information:")
        fmt.Println(string(jsonBytes))
}
