package network

import (
    "encoding/json"
    "net/http"
    "lscc/core"
    "fmt"
    "crypto/sha256"
    "encoding/hex"
    "time"
)

func (n *Node) router() http.Handler {
    n.Logger.Info("Setting up HTTP router...")
    mux := http.NewServeMux()
    
    n.Logger.Info("Registering endpoint: /status")
    mux.HandleFunc("/status", n.handleStatus)
    
    n.Logger.Info("Registering endpoint: /send")
    mux.HandleFunc("/send", n.handleSend)
    
    n.Logger.Info("Registering endpoint: /chain")
    mux.HandleFunc("/chain", n.handleChain)
    
    n.Logger.Info("HTTP router setup completed")
    return mux
}
func (n *Node) handleStatus(w http.ResponseWriter, r *http.Request) {
    n.Logger.Info("=== Status Request ===", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)
    
    w.Header().Set("Content-Type", "application/json")
    status := map[string]interface{}{
        "shardID": n.Config.ShardID,
        "layer":   n.Config.Layer,
        "port":    n.Config.Port,
        "status":  "running",
        "blocks":  len(n.Blockchain.GetBlocks()),
    }
    
    n.Logger.Info("Sending status response", "status", status)
    if err := json.NewEncoder(w).Encode(status); err != nil {
        n.Logger.Error("Failed to encode status response", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    n.Logger.Info("Status request completed successfully")
}
func (n *Node) handleSend(w http.ResponseWriter, r *http.Request) {
    n.Logger.Info("=== Transaction Request ===", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)
    
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
    
    // Calculate the hash of the transaction
    n.Logger.Info("Calculating transaction hash...")
    data := fmt.Sprintf("%s:%s:%f:%d", tx.From, tx.To, tx.Amount, tx.Timestamp)
    hash := sha256.Sum256([]byte(data))
    tx.Hash = hex.EncodeToString(hash[:])
    n.Logger.Info("Transaction hash calculated", "hash", tx.Hash)
    
    // Add the transaction to the blockchain
    n.Logger.Info("Adding transaction to blockchain...")
    if n.Blockchain.Transactions == nil {
        n.Blockchain.Transactions = make(map[string]*core.Transaction)
        n.Logger.Info("Initialized blockchain transactions map")
    }
    n.Blockchain.AddTransaction(&tx)
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
    n.Logger.Info("=== Chain Request ===", "method", r.Method, "url", r.URL.Path, "remote", r.RemoteAddr)
    
    n.Logger.Info("Retrieving blockchain blocks...")
    blocks := n.Blockchain.GetBlocks()
    n.Logger.Info("Retrieved blocks from blockchain", "count", len(blocks))
    
    if len(blocks) == 0 {
        n.Logger.Warn("No blocks found in blockchain")
        http.Error(w, "No blocks found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(blocks); err != nil {
        n.Logger.Error("Failed to encode blocks response", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    n.Logger.Info("Chain request completed successfully", "blocks_sent", len(blocks))
}   
// Remove AddTransaction from here and define it in the core package (core/blockchain.go).