// Applying the provided changes to fix compilation errors and update function signatures.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Height        uint64         `json:"height"`
	Hash          string         `json:"hash"`
	PrevBlockHash string         `json:"prev_block_hash"`
	Transactions  []*Transaction `json:"transactions"`
	Timestamp     int64          `json:"timestamp"`
	Validator     string         `json:"validator"`
	Signature     string         `json:"signature"`
	ShardID       int            `json:"shard_id"`
	Nonce         uint64         `json:"nonce"`
}

func NewBlock(height uint64, prevHash string, transactions []*Transaction, validator string, shardID int) *Block {
	block := &Block{
		Height:        height,
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevHash,
		Transactions:  transactions,
		Validator:     validator,
		ShardID:       shardID,
		Nonce:         0,
	}

	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) CalculateHash() string {
	blockData := fmt.Sprintf("%d:%d:%s:%s:%d:%d",
		b.Height, b.Timestamp, b.PrevBlockHash, b.Validator, b.ShardID, b.Nonce)

	for _, tx := range b.Transactions {
		blockData += ":" + tx.Hash
	}

	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

func (b *Block) Serialize() ([]byte, error) {
	return json.Marshal(b)
}

func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	err := json.Unmarshal(data, &block)
	return &block, err
}

func (b *Block) Validate() bool {
	if b.Height < 0 {
		return false
	}

	if b.Timestamp <= 0 {
		return false
	}

	if b.Hash == "" {
		return false
	}

	// Validate hash
	calculatedHash := b.CalculateHash()
	if calculatedHash != b.Hash {
		return false
	}

	return true
}

func GenesisBlock(shardID int) *Block {
	return &Block{
		Height:        0,
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: "",
		Hash:          "genesis_block_" + fmt.Sprintf("%d", shardID),
		Transactions:  []*Transaction{},
		Validator:     "genesis",
		ShardID:       shardID,
		Nonce:         0,
	}
}
```// Applying the provided changes to fix compilation errors and update function signatures.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Height        uint64         `json:"height"`
	Hash          string         `json:"hash"`
	PrevBlockHash string         `json:"prev_block_hash"`
	Transactions  []*Transaction `json:"transactions"`
	Timestamp     int64          `json:"timestamp"`
	Validator     string         `json:"validator"`
	Signature     string         `json:"signature"`
	ShardID       int            `json:"shard_id"`
	Nonce         uint64         `json:"nonce"`
}

func NewBlock(height uint64, prevHash string, transactions []*Transaction, validator string, shardID int) *Block {
	block := &Block{
		Height:        height,
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevHash,
		Transactions:  transactions,
		Validator:     validator,
		ShardID:       shardID,
		Nonce:         0,
	}

	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) CalculateHash() string {
	blockData := fmt.Sprintf("%d:%d:%s:%s:%d:%d",
		b.Height, b.Timestamp, b.PrevBlockHash, b.Validator, b.ShardID, b.Nonce)

	for _, tx := range b.Transactions {
		blockData += ":" + tx.Hash
	}

	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

func (b *Block) Serialize() ([]byte, error) {
	return json.Marshal(b)
}

func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	err := json.Unmarshal(data, &block)
	return &block, err
}

func (b *Block) Validate() bool {
	if b.Height < 0 {
		return false
	}

	if b.Timestamp <= 0 {
		return false
	}

	if b.Hash == "" {
		return false
	}

	// Validate hash
	calculatedHash := b.CalculateHash()
	if calculatedHash != b.Hash {
		return false
	}

	return true
}

func GenesisBlock(shardID int) *Block {
	return &Block{
		Height:        0,
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: "",
		Hash:          "genesis_block_" + fmt.Sprintf("%d", shardID),
		Transactions:  []*Transaction{},
		Validator:     "genesis",
		ShardID:       shardID,
		Nonce:         0,
	}
}