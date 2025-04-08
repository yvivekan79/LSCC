package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// TransactionType defines the type of transaction
type TransactionType int

const (
	// Regular transaction between accounts
	RegularTransaction TransactionType = iota
	// Cross-shard transaction
	CrossShardTransaction
	// Consensus transaction (validator rewards, staking, etc.)
	ConsensusTransaction
	// Layer-to-layer transaction
	LayerTransaction
)

// Transaction represents a transaction in the blockchain
type Transaction struct {
	Hash        string          `json:"hash"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Amount      float64         `json:"amount"`
	Fee         float64         `json:"fee"`
	Data        []byte          `json:"data"`
	Timestamp   int64           `json:"timestamp"`
	Type        TransactionType `json:"type"`
	Signature   string          `json:"signature"`
	SourceShard int             `json:"source_shard"`
	TargetShard int             `json:"target_shard"`
	Layer       int             `json:"layer"`
	IsConfirmed bool            `json:"is_confirmed"`
	Nonce       uint64          `json:"nonce"`
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount, fee float64, sourceShard, targetShard, layer int, txType TransactionType) (*Transaction, error) {
	tx := &Transaction{
		From:        from,
		To:          to,
		Amount:      amount,
		Fee:         fee,
		Timestamp:   time.Now().Unix(),
		Type:        txType,
		SourceShard: sourceShard,
		TargetShard: targetShard,
		Layer:       layer,
		IsConfirmed: false,
	}

	// Calculate hash
	hash, err := tx.CalculateHash()
	if err != nil {
		return nil, err
	}
	tx.Hash = hash

	return tx, nil
}

// CalculateHash calculates the hash of the transaction
func (tx *Transaction) CalculateHash() (string, error) {
	// Create a temporary copy without the hash and signature fields
	txCopy := *tx
	txCopy.Hash = ""
	txCopy.Signature = ""

	txJSON, err := json.Marshal(txCopy)
	if err != nil {
		return "", err
	}

	hasher := sha256.New()
	hasher.Write(txJSON)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Sign signs the transaction with the provided private key
func (tx *Transaction) Sign(privateKey string) error {
	// In a real implementation, this would use actual cryptographic signing
	// For this example, we'll just set a dummy signature
	tx.Signature = "signed:" + tx.Hash[:8] + ":" + privateKey[:8]
	return nil
}

// VerifySignature verifies the transaction's signature
func (tx *Transaction) VerifySignature() bool {
	// In a real implementation, this would verify the signature cryptographically
	// For this example, we'll just check if the signature exists
	return tx.Signature != ""
}

// IsCrossShard checks if the transaction crosses shard boundaries
func (tx *Transaction) IsCrossShard() bool {
	return tx.SourceShard != tx.TargetShard
}

// IsValid checks if the transaction is valid
func (tx *Transaction) IsValid() bool {
	// Check if the transaction has a valid signature
	if !tx.VerifySignature() {
		return false
	}

	// Check for valid amounts
	if tx.Amount <= 0 || tx.Fee < 0 {
		return false
	}

	// Verify hash integrity
	hash, err := tx.CalculateHash()
	if err != nil {
		return false
	}
	if hash != tx.Hash {
		return false
	}

	return true
}

// Confirm marks the transaction as confirmed
func (tx *Transaction) Confirm() {
	tx.IsConfirmed = true
}

// UnmarshalTransaction deserializes a transaction from JSON
func UnmarshalTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(data, &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// MarshalTransaction serializes a transaction to JSON
func MarshalTransaction(tx *Transaction) ([]byte, error) {
	return json.Marshal(tx)
}
