package network

import (
    "encoding/json"
    "net/http"
    "lscc/core"
    "fmt"
     "crypto/sha256"
    "encoding/hex"
  
)

func (n *Node) router() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/status", n.handleStatus)
    mux.HandleFunc("/send", n.handleSend)
    mux.HandleFunc("/chain", n.handleChain)
    return mux
}
func (n *Node) handleStatus(w http.ResponseWriter, r *http.Request) {   
    w.Header().Set("Content-Type", "application/json")
    status := map[string]interface{}{
        "shardID": n.Config.ShardID,
        "layer":   n.Config.Layer,
        "port":    n.Config.Port,
    }
    json.NewEncoder(w).Encode(status)
}
func (n *Node) handleSend(w http.ResponseWriter, r *http.Request) {
    var tx core.Transaction
    if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
        http.Error(w, "Invalid transaction data", http.StatusBadRequest)
        return
    }
    if tx.From == "" || tx.To == "" || tx.Amount <= 0 {
        http.Error(w, "Invalid transaction fields", http.StatusBadRequest)
        return
    }
    
    // Calculate the hash of the transaction
    data := fmt.Sprintf("%s:%s:%f:%d", tx.From, tx.To, tx.Amount, tx.Timestamp)
    hash := sha256.Sum256([]byte(data))
    tx.Hash = hex.EncodeToString(hash[:])
    
    // Add the transaction to the blockchain
    // Ensure Transactions map is initialized
    if n.Blockchain.Transactions == nil {
        n.Blockchain.Transactions = make(map[string]*core.Transaction)
    }
    n.Blockchain.AddTransaction(&tx)
    
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(tx)
}
func (n *Node) handleChain(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    blocks := n.Blockchain.GetBlocks()
    if len(blocks) == 0 {
        http.Error(w, "No blocks found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(blocks)
}   
// Remove AddTransaction from here and define it in the core package (core/blockchain.go).