package config

import (
    "encoding/json"
    "io/ioutil"
)

type ConsensusParams struct {
    Difficulty int `json:"difficulty"`
}

type Config struct {
    NodeID         string           `json:"node_id"`
    Port           int              `json:"port"`
    ShardID        int              `json:"shard_id"`
    Layer          int              `json:"layer"`
    IsRelay        bool             `json:"is_relay"`
    BootstrapNodes []string         `json:"bootstrap_nodes"`
    ConsensusType  string           `json:"consensus_type"`
    LoggingLevel   string           `json:"logging_level"`
    ConsensusParams ConsensusParams `json:"consensus_params"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

