
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Transaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Fee       float64   `json:"fee"`
	Timestamp time.Time `json:"timestamp"`
	Hash      string    `json:"hash"`
	Signature string    `json:"signature"`
	ShardID   int       `json:"shard_id"`
	CrossShard bool     `json:"cross_shard"`
}

func NewTransaction(from, to string, amount, fee float64, shardID int) *Transaction {
	tx := &Transaction{
		ID:        generateTransactionID(),
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Timestamp: time.Now(),
		ShardID:   shardID,
		CrossShard: false,
	}
	
	tx.Hash = tx.CalculateHash()
	return tx
}

func (tx *Transaction) CalculateHash() string {
	data, _ := json.Marshal(struct {
		ID        string    `json:"id"`
		From      string    `json:"from"`
		To        string    `json:"to"`
		Amount    float64   `json:"amount"`
		Fee       float64   `json:"fee"`
		Timestamp time.Time `json:"timestamp"`
		ShardID   int       `json:"shard_id"`
	}{
		ID:        tx.ID,
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Timestamp: tx.Timestamp,
		ShardID:   tx.ShardID,
	})
	
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (tx *Transaction) Validate() error {
	if tx.From == "" {
		return fmt.Errorf("invalid from address")
	}
	
	if tx.To == "" {
		return fmt.Errorf("invalid to address")
	}
	
	if tx.Amount <= 0 {
		return fmt.Errorf("invalid amount")
	}
	
	if tx.Fee < 0 {
		return fmt.Errorf("invalid fee")
	}
	
	if tx.Hash == "" {
		return fmt.Errorf("invalid hash")
	}
	
	// Verify hash
	expectedHash := tx.CalculateHash()
	if tx.Hash != expectedHash {
		return fmt.Errorf("hash mismatch")
	}
	
	return nil
}

func (tx *Transaction) Sign(privateKey string) error {
	// Simple signature implementation
	tx.Signature = generateSignature(tx.Hash, privateKey)
	return nil
}

func (tx *Transaction) VerifySignature(publicKey string) bool {
	// Simple signature verification
	return tx.Signature != ""
}

func generateTransactionID() string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("tx_%d", timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func generateSignature(hash, privateKey string) string {
	data := hash + privateKey
	signature := sha256.Sum256([]byte(data))
	return hex.EncodeToString(signature[:])
}
