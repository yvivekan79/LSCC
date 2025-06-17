package config

import (
	"encoding/json"
	"os"
)

// Config holds the configuration for the LSCC node
type Config struct {
	NodeID             string          `json:"node_id"`
	ShardID            int             `json:"shard_id"`
	ShardCount         int             `json:"shard_count"`
	IsRelay            bool            `json:"is_relay"`
	LayerCount         int             `json:"layer_count"`
	MaxTransPerBlock   int             `json:"max_trans_per_block"`
	ShardingStrategy   string          `json:"sharding_strategy"`
	ConsensusType      string          `json:"consensus_type"`
	ConsensusParams    ConsensusParams `json:"consensus_params"`
	Port               int             `json:"port"`
	BootstrapNodes     []string        `json:"bootstrap_nodes"`
	CrossShardRatio    float64         `json:"cross_shard_ratio"`
	RelayNodesRatio    float64         `json:"relay_nodes_ratio"`
	NodesPerShard      int             `json:"nodes_per_shard"`
	NetworkID          string          `json:"network_id"`
	BlockTime          int             `json:"block_time"`
	MinConfirmations   int             `json:"min_confirmations"`
	CrossChannelVerify bool            `json:"cross_channel_verify"`
	ConnectionTimeout  int             `json:"connection_timeout"`
}

type ConsensusParams struct {
	BlockTime          int      `json:"block_time"`
	MinConfirmations   int      `json:"min_confirmations"`
	MinStake           float64  `json:"min_stake"`
	StakingReward      float64  `json:"staking_reward"`
	Validators         []string `json:"validators"`
	ViewChangeTimeout  int      `json:"view_change_timeout"`
	CrossChannelVerify bool     `json:"cross_channel_verify"`
	CrossLayerTimeout  int      `json:"cross_layer_timeout"`
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
