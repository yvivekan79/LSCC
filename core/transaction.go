
package core

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

type Transaction struct {
    From        string  `json:"from"`
    To          string  `json:"to"`
    Amount      float64 `json:"amount"`
    Fee         float64 `json:"fee"`
    Timestamp   int64   `json:"timestamp"`
    Hash        string  `json:"hash"`
    Signature   string  `json:"signature"`
    SourceShard int     `json:"source_shard"`
    TargetShard int     `json:"target_shard"`
    Nonce       uint64  `json:"nonce"`
    Type        int     `json:"type"`
}

func (tx *Transaction) Validate() bool {
    if tx.From == "" || tx.To == "" {
        return false
    }
    if tx.Amount <= 0 {
        return false
    }
    if tx.Fee < 0 {
        return false
    }
    return true
}

func (tx *Transaction) CalculateHash() string {
    data := fmt.Sprintf("%s:%s:%f:%f:%d:%d:%d:%d", 
        tx.From, tx.To, tx.Amount, tx.Fee, tx.Timestamp, 
        tx.SourceShard, tx.TargetShard, tx.Nonce)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func NewTransaction(from, to string, amount, fee float64, sourceShardID, targetShardID int) *Transaction {
    tx := &Transaction{
        From:        from,
        To:          to,
        Amount:      amount,
        Fee:         fee,
        SourceShard: sourceShardID,
        TargetShard: targetShardID,
        Timestamp:   0, // Will be set when processed
        Nonce:       0, // Will be set based on sender's transaction count
        Type:        0, // 0 = regular, 1 = cross-shard
    }
    
    if sourceShardID != targetShardID {
        tx.Type = 1
    }
    
    return tx
}
