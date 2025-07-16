package core

import (
    "crypto/sha256"
    "encoding/hex"
    "time"
)

type BlockHeader struct {
    PreviousHash string
    Timestamp    int64
    Nonce        int
    ShardID      int
    Layer        int
    CrossRefs    []string
}



// Block represents a block in the blockchain.
type Block struct {
    Header       BlockHeader
    Hash         string
    PrevHash     string
    Transactions []*Transaction
    Height       int
    Nonce        int
}
func NewBlock(prevHash string, txs []*Transaction, shardID int, layer int) *Block {
    block := &Block{
        Header: BlockHeader{
            PreviousHash: prevHash,
            Timestamp:    time.Now().Unix(),
            ShardID:      shardID,
            Layer:        layer,
        },
        Transactions: txs,
    }
    block.Hash = block.CalculateHash()
    return block
}

func (b *Block) CalculateHash() string {
    h := sha256.New()
    h.Write([]byte(b.Header.PreviousHash))
    h.Write([]byte(string(b.Header.Timestamp)))
    for _, tx := range b.Transactions {
        h.Write([]byte(tx.Hash))
    }
    return hex.EncodeToString(h.Sum(nil))
}

func (b *Block) IsValid(prevHash string) bool {
    return b.Header.PreviousHash == prevHash && b.Hash == b.CalculateHash()
}

