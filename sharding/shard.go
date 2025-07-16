package sharding

type Shard struct {
    ID     int
    Layer  int
    Relay  bool
    Peers  []string
    Active bool
}

func NewShard(id, layer int, relay bool) *Shard {
    return &Shard{
        ID:    id,
        Layer: layer,
        Relay: relay,
        Peers: []string{},
        Active: true,
    }
}

func (s *Shard) AddPeer(peer string) {
    s.Peers = append(s.Peers, peer)
}

func (s *Shard) IsRelayNode() bool {
    return s.Relay
}

