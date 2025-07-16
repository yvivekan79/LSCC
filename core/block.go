
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type Block struct {
	Index         uint64         `json:"index"`
	PrevBlockHash string         `json:"prev_block_hash"`
	Timestamp     time.Time      `json:"timestamp"`
	Transactions  []*Transaction `json:"transactions"`
	Validator     string         `json:"validator"`
	ShardID       int            `json:"shard_id"`
	Hash          string         `json:"hash"`
	MerkleRoot    string         `json:"merkle_root"`
	Layer         int            `json:"layer"`
}

func NewBlock(index uint64, prevBlockHash string, transactions []*Transaction, validator string, shardID int) *Block {
	block := &Block{
		Index:         index,
		PrevBlockHash: prevBlockHash,
		Timestamp:     time.Now(),
		Transactions:  transactions,
		Validator:     validator,
		ShardID:       shardID,
		Layer:         0,
	}
	
	block.MerkleRoot = block.calculateMerkleRoot()
	block.Hash = block.CalculateHash()
	
	return block
}

func (b *Block) CalculateHash() string {
	data, _ := json.Marshal(struct {
		Index         uint64    `json:"index"`
		PrevBlockHash string    `json:"prev_block_hash"`
		Timestamp     time.Time `json:"timestamp"`
		MerkleRoot    string    `json:"merkle_root"`
		Validator     string    `json:"validator"`
		ShardID       int       `json:"shard_id"`
	}{
		Index:         b.Index,
		PrevBlockHash: b.PrevBlockHash,
		Timestamp:     b.Timestamp,
		MerkleRoot:    b.MerkleRoot,
		Validator:     b.Validator,
		ShardID:       b.ShardID,
	})
	
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (b *Block) calculateMerkleRoot() string {
	if len(b.Transactions) == 0 {
		return ""
	}
	
	var hashes []string
	for _, tx := range b.Transactions {
		hashes = append(hashes, tx.Hash)
	}
	
	for len(hashes) > 1 {
		var newHashes []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				hash := sha256.Sum256([]byte(combined))
				newHashes = append(newHashes, hex.EncodeToString(hash[:]))
			} else {
				newHashes = append(newHashes, hashes[i])
			}
		}
		hashes = newHashes
	}
	
	return hashes[0]
}

func (b *Block) Validate() bool {
	// Basic validation
	if b.Index < 0 {
		return false
	}
	
	if b.Hash == "" {
		return false
	}
	
	// Verify hash
	expectedHash := b.CalculateHash()
	if b.Hash != expectedHash {
		return false
	}
	
	// Verify merkle root
	expectedMerkleRoot := b.calculateMerkleRoot()
	if b.MerkleRoot != expectedMerkleRoot {
		return false
	}
	
	// Validate all transactions
	for _, tx := range b.Transactions {
		if err := tx.Validate(); err != nil {
			return false
		}
	}
	
	return true
}

func (b *Block) AddTransaction(tx *Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	
	b.Transactions = append(b.Transactions, tx)
	b.MerkleRoot = b.calculateMerkleRoot()
	b.Hash = b.CalculateHash()
	
	return nil
}

func (b *Block) GetTransactionCount() int {
	return len(b.Transactions)
}

func (b *Block) GetSize() int {
	data, _ := json.Marshal(b)
	return len(data)
}
