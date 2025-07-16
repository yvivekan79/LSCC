package network

import (
	"encoding/json"
	"fmt"
	"lscc/core"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (n *Node) router() http.Handler {
	n.Logger.Info("Setting up HTTP router...")
	mux := http.NewServeMux()

	n.Logger.Info("Registering endpoint: /")
	mux.HandleFunc("/", n.handleDashboard)

	n.Logger.Info("Registering endpoint: /status")
	mux.HandleFunc("/status", n.handleStatus)

	n.Logger.Info("Registering endpoint: /send")
	mux.HandleFunc("/send", n.handleSend)

	n.Logger.Info("Registering endpoint: /chain")
	mux.HandleFunc("/chain", n.handleChain)

	n.Logger.Info("Registering endpoint: /mempool")
	mux.HandleFunc("/mempool", n.handleMempool)

	n.Logger.Info("Registering endpoint: /shard-info")
	mux.HandleFunc("/shard-info", n.handleShardInfo)

	n.Logger.Info("HTTP router setup completed")
	return mux
}
func (n *Node) handleDashboard(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info("=== Dashboard Request ===", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)

	// Serve the dashboard HTML file
	dashboardPath := filepath.Join("web", "index.html")
	if _, err := os.Stat(dashboardPath); os.IsNotExist(err) {
		n.Logger.Error("Dashboard file not found", "path", dashboardPath)
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, dashboardPath)
	n.Logger.Info("Dashboard served successfully")
}

func (n *Node) handleStatus(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info("Received status request")

	// Get blockchain info
	blockchainInfo := n.Blockchain.GetBlockchainInfo()

	// Calculate network statistics
	totalTxs := 0
	for _, block := range n.Blockchain.Blocks {
		totalTxs += len(block.Transactions)
	}

	status := map[string]interface{}{
		"nodeID":               n.Config.NodeID,
		"shardID":              n.Config.ShardID,
		"layer":                n.Config.Layer,
		"port":                 n.Config.Port,
		"consensus":            n.Config.ConsensusType,
		"height":               n.Blockchain.GetHeight(),
		"blocks":               len(n.Blockchain.Blocks),
		"pending_transactions": len(n.Blockchain.Mempool),
		"total_transactions":   totalTxs,
		"status":               "running",
		"timestamp":            time.Now().Unix(),
		"uptime":               time.Now().Unix(),
		"blockchain_info":      blockchainInfo,
		"shard_details": map[string]interface{}{
			"shard_id":     n.Config.ShardID,
			"layer":        n.Config.Layer,
			"is_relay":     false, // You can add this to config if needed
			"consensus":    n.Config.ConsensusType,
			"node_count":   1, // In a real implementation, this would be dynamic
		},
		"network_info": map[string]interface{}{
			"listening_address": fmt.Sprintf("0.0.0.0:%d", n.Config.Port),
			"api_endpoints": []string{
				"/status",
				"/send",
				"/chain",
				"/mempool",
				"/shard-info",
			},
		},
	}

	n.Logger.Info("Status request processed successfully",
		"height", status["height"],
		"shard", status["shardID"],
		"layer", status["layer"])

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(status)
}
func (n *Node) handleSend(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info("=== Transaction Request ===", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		n.Logger.Warn("Invalid method for /send endpoint", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tx core.Transaction
	n.Logger.Info("Decoding transaction data...")
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		n.Logger.Error("Failed to decode transaction data", "error", err)
		http.Error(w, "Invalid transaction data", http.StatusBadRequest)
		return
	}

	n.Logger.Info("Validating transaction", "from", tx.From, "to", tx.To, "amount", tx.Amount)
	if tx.From == "" || tx.To == "" || tx.Amount <= 0 {
		n.Logger.Warn("Invalid transaction fields", "from", tx.From, "to", tx.To, "amount", tx.Amount)
		http.Error(w, "Invalid transaction fields", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if tx.Timestamp == 0 {
		tx.Timestamp = time.Now().Unix()
	}

	// Set default values for missing fields
	if tx.Fee == 0 {
		tx.Fee = 0.01 // Default fee
	}
	if tx.SourceShard == 0 {
		tx.SourceShard = n.Config.ShardID
	}
	if tx.TargetShard == 0 {
		tx.TargetShard = n.Config.ShardID
	}

	// Calculate the hash of the transaction
	n.Logger.Info("Calculating transaction hash...")
	tx.Hash = tx.CalculateHash()
	n.Logger.Info("Transaction hash calculated", "hash", tx.Hash)

	// Add the transaction to the blockchain
	n.Logger.Info("Adding transaction to blockchain...")
	err := n.Blockchain.AddTransaction(&tx)
	if err != nil {
		n.Logger.Error("Failed to add transaction to blockchain", "error", err)
		http.Error(w, fmt.Sprintf("Failed to add transaction: %v", err), http.StatusBadRequest)
		return
	}
	n.Logger.Info("Transaction added to blockchain successfully", "hash", tx.Hash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(tx); err != nil {
		n.Logger.Error("Failed to encode transaction response", "error", err)
		return
	}
	n.Logger.Info("Transaction request completed successfully", "hash", tx.Hash)
}
func (n *Node) handleChain(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info("Received blockchain request")

	blocks := n.Blockchain.GetBlocks()

	// Enrich blocks with additional information
	enrichedBlocks := make([]map[string]interface{}, len(blocks))
	for i, block := range blocks {
		enrichedBlocks[i] = map[string]interface{}{
			"height":        block.Height,
			"hash":          block.Hash,
			"prevBlockHash": block.PrevBlockHash,
			"merkleRoot":    block.Hash, // Use block hash as merkle root
			"timestamp":     block.Timestamp,
			"shardID":       block.ShardID,
			"layer":         0, // Default layer
			"transactions":  block.Transactions,
			"tx_count":      len(block.Transactions),
			"size":          len(fmt.Sprintf("%+v", block)), // Approximate size
		}
	}

	

	n.Logger.Info("Blockchain request processed successfully",
		"blocks_count", len(blocks),
		"height", n.Blockchain.GetHeight())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(enrichedBlocks)
}

func (n *Node) handleMempool(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info("Received mempool request")

	pendingTxs := n.Blockchain.GetPendingTransactions()

	response := map[string]interface{}{
		"pending_transactions": pendingTxs,
		"count":               len(pendingTxs),
		"shard_id":            n.Config.ShardID,
		"node_id":             n.Config.NodeID,
		"timestamp":           time.Now().Unix(),
	}

	n.Logger.Info("Mempool request processed", "pending_count", len(pendingTxs))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(response)
}

func (n *Node) handleShardInfo(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info("Received shard-info request")

	// Get cross-shard transactions
	crossShardTxs := n.Blockchain.GetCrossShardTransactions()

	shardInfo := map[string]interface{}{
		"shard_id":              n.Config.ShardID,
		"layer":                 n.Config.Layer,
		"node_id":               n.Config.NodeID,
		"consensus_type":        n.Config.ConsensusType,
		"is_relay":              false, // Add to config if needed
		"total_blocks":          len(n.Blockchain.Blocks),
		"blockchain_height":     n.Blockchain.GetHeight(),
		"pending_transactions":  len(n.Blockchain.Mempool),
		"cross_shard_txs":       len(crossShardTxs),
		"cross_shard_details":   crossShardTxs,
		"network_info": map[string]interface{}{
			"port":            n.Config.Port,
			"listening_addr":  fmt.Sprintf("0.0.0.0:%d", n.Config.Port),
		},
		"performance_metrics": map[string]interface{}{
			"uptime":          time.Now().Unix(),
			"last_block_time": func() int64 {
				lastBlock := n.Blockchain.GetLastBlock()
				if lastBlock != nil {
					return lastBlock.Timestamp
				}
				return 0
			}(),
		},
		"timestamp": time.Now().Unix(),
	}

	n.Logger.Info("Shard info request processed",
		"shard_id", n.Config.ShardID,
		"layer", n.Config.Layer)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(shardInfo)
}

// Remove AddTransaction from here and define it in the core package (core/blockchain.go).