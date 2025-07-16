package sharding

type CrossReference struct {
    SourceShard int
    TargetShard int
    BlockHash   string
    Timestamp   int64
}

func NewCrossReference(source, target int, hash string, ts int64) *CrossReference {
    return &CrossReference{
        SourceShard: source,
        TargetShard: target,
        BlockHash:   hash,
        Timestamp:   ts,
    }
}

