
package core

// Consensus represents the interface for consensus algorithms
type Consensus interface {
	Start() error
	Stop() error
	ValidateBlock(block *Block) error
	ProposeBlock(transactions []*Transaction, prevBlockHash string, height uint64, shardID int) (*Block, error)
}
