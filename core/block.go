package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
	ShardID      int           `json:"shard_id"`
	Signature    string        `json:"signature"`
}

// BlockHeader contains metadata of a block
type BlockHeader struct {
	Version        uint32    `json:"version"`
	PreviousHash   string    `json:"previous_hash"`
	MerkleRoot     string    `json:"merkle_root"`
	Timestamp      int64     `json:"timestamp"`
	Difficulty     uint32    `json:"difficulty"`
	Nonce          uint64    `json:"nonce"`
	Height         uint64    `json:"height"`
	ValidatorID    string    `json:"validator_id"`
	CrossRefs      []CrossRef `json:"cross_refs"`
	Layer          int       `json:"layer"`
}

// CrossRef represents a reference to a block in another shard
type CrossRef struct {
	ShardID   int    `json:"shard_id"`
	BlockHash string `json:"block_hash"`
	Height    uint64 `json:"height"`
}

// NewBlock creates a new block
func NewBlock(prevHash string, height uint64, shardID int, layer int, validatorID string) *Block {
	block := &Block{
		Header: BlockHeader{
			Version:      1,
			PreviousHash: prevHash,
			Timestamp:    time.Now().Unix(),
			Height:       height,
			ValidatorID:  validatorID,
			CrossRefs:    []CrossRef{},
			Layer:        layer,
		},
		Transactions: []Transaction{},
		ShardID:      shardID,
	}
	return block
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx Transaction) {
	b.Transactions = append(b.Transactions, tx)
	// Update merkle root after adding transaction
	b.Header.MerkleRoot = b.CalculateMerkleRoot()
}

// AddCrossReference adds a cross-shard reference to the block
func (b *Block) AddCrossReference(shardID int, blockHash string, height uint64) {
	crossRef := CrossRef{
		ShardID:   shardID,
		BlockHash: blockHash,
		Height:    height,
	}
	b.Header.CrossRefs = append(b.Header.CrossRefs, crossRef)
}

// CalculateMerkleRoot calculates the merkle root of the transactions
// This is a simplified implementation - in a real system, you would use a proper Merkle tree
func (b *Block) CalculateMerkleRoot() string {
	var txHashes []string
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash)
	}
	
	// If no transactions, return empty hash
	if len(txHashes) == 0 {
		return ""
	}
	
	// Combine all transaction hashes and hash them together
	// Note: This is not a true Merkle tree, just a simplification
	combined := ""
	for _, hash := range txHashes {
		combined += hash
	}
	
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Hash calculates the hash of the block
func (b *Block) Hash() (string, error) {
	headerBytes, err := json.Marshal(b.Header)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	hasher.Write(headerBytes)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Sign signs the block with the provided private key
func (b *Block) Sign(privateKey string) error {
	// In a real implementation, this would use actual cryptographic signing
	// For now, we'll just simulate signing with a placeholder
	hash, err := b.Hash()
	if err != nil {
		return err
	}
	
	// Simulate signature (in reality, would use crypto library)
	b.Signature = fmt.Sprintf("signed:%s:%s", privateKey[:8], hash)
	return nil
}

// VerifySignature verifies the block's signature
func (b *Block) VerifySignature(publicKey string) bool {
	// In a real implementation, this would verify the signature cryptographically
	// For now, we'll just return true as a placeholder
	return len(b.Signature) > 0
}

// IsValid checks if the block is valid
func (b *Block) IsValid(prevBlock *Block) bool {
	// Check if previous hash matches
	if prevBlock != nil {
		prevHash, err := prevBlock.Hash()
		if err != nil {
			return false
		}
		if b.Header.PreviousHash != prevHash {
			return false
		}
		
		// Check if height is correct
		if b.Header.Height != prevBlock.Header.Height+1 {
			return false
		}
	}
	
	// Verify merkle root
	calculatedRoot := b.CalculateMerkleRoot()
	if b.Header.MerkleRoot != calculatedRoot {
		return false
	}
	
	// Additional validation can be added here
	
	return true
}
