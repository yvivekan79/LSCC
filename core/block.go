package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	CrossRefs    []string
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
	ShardID      int           `json:"shard_id"`
	Signature    string        `json:"signature"`
}

// BlockHeader contains metadata of a block
type BlockHeader struct {
	Layer         int
	Version       uint32     `json:"version"`
	PreviousHash  string     `json:"previous_hash"`
	MerkleRoot    string     `json:"merkle_root"`
	Timestamp     int64      `json:"timestamp"`
	Difficulty    uint32     `json:"difficulty"`
	Nonce         uint64     `json:"nonce"`
	Height        uint64     `json:"height"`
	ValidatorID   string     `json:"validator_id"`
	CrossRefs     []CrossRef `json:"cross_refs"`
}

// CrossRef represents a reference to a block in another shard
type CrossRef struct {
	ShardID   int    `json:"shard_id"`
	BlockHash string `json:"block_hash"`
	Height    uint64 `json:"height"`
}

// NewBlock creates a new block
func NewBlock(prevBlock *Block, txs []*Transaction, createdBy string) *Block {
	block := &Block{
		Header: BlockHeader{
			PreviousHash: prevBlock.Header.PreviousHash,
			ValidatorID:  createdBy,
			Timestamp:    time.Now().Unix(),
			Height:       prevBlock.Header.Height + 1,
			Layer:        prevBlock.Header.Layer + 1,
		},
		Transactions: convertTxPtrSliceToValue(txs),
	}
	block.Header.MerkleRoot = block.CalculateMerkleRoot()
	return block
}

func convertTxPtrSliceToValue(txs []*Transaction) []Transaction {
	var result []Transaction
	for _, tx := range txs {
		if tx != nil {
			result = append(result, *tx)
		}
	}
	return result
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx Transaction) {
	b.Transactions = append(b.Transactions, tx)
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
func (b *Block) CalculateMerkleRoot() string {
	var txHashes []string
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash)
	}
	if len(txHashes) == 0 {
		return ""
	}
	combined := ""
	for _, h := range txHashes {
		combined += h
	}
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// Hash calculates the hash of the block
func (b *Block) Hash() (string, error) {
	headerBytes, err := json.Marshal(b.Header)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(headerBytes)
	return hex.EncodeToString(hash[:]), nil
}

// Sign signs the block with the provided private key (simulated)
func (b *Block) Sign(privateKey string) error {
	hash, err := b.Hash()
	if err != nil {
		return err
	}
	b.Signature = fmt.Sprintf("signed_by_%s:%s", privateKey[:8], hash)
	return nil
}

// VerifySignature verifies the block's signature (simulated)
func (b *Block) VerifySignature() bool {
	return strings.HasPrefix(b.Signature, "signed_by_")
}

// IsValid checks if the block is valid
func (b *Block) IsValid(prevBlock *Block) bool {
	if !b.VerifySignature() {
		return false
	}
	if b.Header.PreviousHash != prevBlock.Header.PreviousHash {
		return false
	}
	if b.Header.Height != prevBlock.Header.Height+1 {
		return false
	}
	calculatedRoot := b.CalculateMerkleRoot()
	if b.Header.MerkleRoot != calculatedRoot {
		return false
	}
	return true
}
