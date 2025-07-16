
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
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

func NewTransaction(from, to string, amount, fee float64, sourceShard, targetShard int) *Transaction {
	tx := &Transaction{
		From:        from,
		To:          to,
		Amount:      amount,
		Fee:         fee,
		Timestamp:   time.Now().Unix(),
		SourceShard: sourceShard,
		TargetShard: targetShard,
		Nonce:       0,
		Type:        0,
	}

	tx.Hash = tx.CalculateHash()
	return tx
}

func (tx *Transaction) CalculateHash() string {
	txData := fmt.Sprintf("%s:%s:%.8f:%.8f:%d:%d:%d:%d:%d",
		tx.From, tx.To, tx.Amount, tx.Fee, tx.Timestamp, tx.SourceShard, tx.TargetShard, tx.Nonce, tx.Type)

	hash := sha256.Sum256([]byte(txData))
	return hex.EncodeToString(hash[:])
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

	if tx.Timestamp <= 0 {
		return false
	}

	if tx.Hash == "" {
		return false
	}

	// Validate hash
	calculatedHash := tx.CalculateHash()
	if calculatedHash != tx.Hash {
		return false
	}

	return true
}

func (tx *Transaction) Serialize() ([]byte, error) {
	return json.Marshal(tx)
}

func DeserializeTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(data, &tx)
	return &tx, err
}

func (tx *Transaction) IsCrossShardTransaction() bool {
	return tx.SourceShard != tx.TargetShard
}

func (tx *Transaction) Sign(signature string) {
	tx.Signature = signature
}

func (tx *Transaction) IsValid() bool {
	return tx.Validate()
}

func (tx *Transaction) Confirm() {
	// Mark transaction as confirmed - this could be expanded
	// with additional confirmation logic
}
