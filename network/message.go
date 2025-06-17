package network

import (
        "encoding/json"
        "time"

        "lscc/core"
)

// MessageType represents the type of a network message
type MessageType int

const (
        // MessageTypeHandshake is the message for peer handshakes
        MessageTypeHandshake MessageType = iota
        // MessageTypePeerListRequest requests a list of peers
        MessageTypePeerListRequest
        // MessageTypePeerList contains a list of peers
        MessageTypePeerList
        // MessageTypeTransaction contains a transaction
        MessageTypeTransaction
        // MessageTypeBlock contains a block
        MessageTypeBlock
        // MessageTypeBlockRequest requests a block
        MessageTypeBlockRequest
        // MessageTypeBlockResponse responds to a block request
        MessageTypeBlockResponse
        // MessageTypeConsensus contains consensus-specific data
        MessageTypeConsensus
        // MessageTypeCrossShardTx contains a cross-shard transaction
        MessageTypeCrossShardTx
)

// Message represents a network message
type Message struct {
        Type      MessageType     `json:"type"`
        Timestamp int64           `json:"timestamp"`
        Data      json.RawMessage `json:"data,omitempty"`
}

// HandshakeMessage is sent when a peer connects
type HandshakeMessage struct {
        NodeID       string `json:"node_id"`
        Version      string `json:"version"`
        ShardID      int    `json:"shard_id"`
        IsRelay      bool   `json:"is_relay"`
        Port         int    `json:"port"`
        Timestamp    int64  `json:"timestamp"`
}

// PeerInfo contains information about a peer
type PeerInfo struct {
        ID        string `json:"id"`
        Address   string `json:"address"`
        ShardID   int    `json:"shard_id"`
        IsRelay   bool   `json:"is_relay"`
        LastSeen  int64  `json:"last_seen"`
}

// BlockRequestMessage requests a specific block
type BlockRequestMessage struct {
        RequestID string `json:"request_id"`
        Hash      string `json:"hash,omitempty"`
        Height    uint64 `json:"height,omitempty"`
}

// BlockResponseMessage contains a block in response to a request
type BlockResponseMessage struct {
        RequestID string     `json:"request_id"`
        Block     core.Block `json:"block"`
}

// TransactionMessage contains a transaction
type TransactionMessage struct {
        Transaction core.Transaction `json:"transaction"`
        Timestamp   int64            `json:"timestamp"`
}

// CrossShardTransactionMessage contains a cross-shard transaction
type CrossShardTransactionMessage struct {
        Transaction  core.Transaction `json:"transaction"`
        SourceShard  int              `json:"source_shard"`
        TargetShard  int              `json:"target_shard"`
        Timestamp    int64            `json:"timestamp"`
}

// ConsensusMessage contains consensus algorithm specific data
type ConsensusMessage struct {
        Type      string          `json:"type"`
        ShardID   int             `json:"shard_id"`
        NodeID    string          `json:"node_id"`
        Timestamp int64           `json:"timestamp"`
        Data      json.RawMessage `json:"data"`
}

// NewCrossShardTransactionMessage creates a new cross-shard transaction message
func NewCrossShardTransactionMessage(tx *core.Transaction) CrossShardTransactionMessage {
        return CrossShardTransactionMessage{
                Transaction: *tx,
                SourceShard: tx.SourceShard,
                TargetShard: tx.TargetShard,
                Timestamp:   time.Now().Unix(),
        }
}

// NewTransactionMessage creates a new transaction message
func NewTransactionMessage(tx *core.Transaction) TransactionMessage {
        return TransactionMessage{
                Transaction: *tx,
                Timestamp:   time.Now().Unix(),
        }
}

// NewBlockRequestMessage creates a new block request message
func NewBlockRequestMessage(requestID string, hash string, height uint64) BlockRequestMessage {
        return BlockRequestMessage{
                RequestID: requestID,
                Hash:      hash,
                Height:    height,
        }
}
