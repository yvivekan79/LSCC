package config

import (
	"encoding/json"
	"os"
)

// Config holds the configuration for the LSCC node
type Config struct {
	// Node configuration
	NodeID         string   `json:"node_id"`
	Port           int      `json:"port"`
	BootstrapNodes []string `json:"bootstrap_nodes"`
	IsRelay        bool     `json:"is_relay"`

	// Sharding configuration
	ShardID          int `json:"shard_id"`
	ShardCount       int `json:"shard_count"`
	CrossShardRatio  int `json:"cross_shard_ratio"`
	RelayNodesRatio  int `json:"relay_nodes_ratio"`
	LayerCount       int `json:"layer_count"`
	NodesPerShard    int `json:"nodes_per_shard"`
	ShardingStrategy int `json:"sharding_strategy"`

	// Consensus configuration
	ConsensusType      string `json:"consensus_type"`
	BlockTime          int    `json:"block_time"`
	MinConfirmations   int    `json:"min_confirmations"`
	MaxTransPerBlock   int    `json:"max_transactions_per_block"`
	CrossChannelVerify bool   `json:"cross_channel_verify"`

	// Network configuration
	ConnectionTimeout int    `json:"connection_timeout"`
	SyncInterval      int    `json:"sync_interval"`
	PeerLimit         int    `json:"peer_limit"`
	DataDir           string `json:"data_dir"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		NodeID:            "",
		Port:              8000,
		BootstrapNodes:    []string{},
		IsRelay:           false,
		ShardID:           -1, // auto-assign
		ShardCount:        4,
		CrossShardRatio:   20, // 20% of transactions are cross-shard
		RelayNodesRatio:   10, // 10% of nodes are relay nodes
		LayerCount:        3,
		NodesPerShard:     10,
		ShardingStrategy:  0, // 0 = static, 1 = dynamic
		ConsensusType:     "pos",
		BlockTime:         5, // seconds
		MinConfirmations:  6,
		MaxTransPerBlock:  1000,
		CrossChannelVerify: true,
		ConnectionTimeout: 30, // seconds
		SyncInterval:      60, // seconds
		PeerLimit:         50,
		DataDir:           "./data",
	}
}
