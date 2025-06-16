
package metrics

import (
    "encoding/csv"
    "fmt"
    "os"
    "sync"
    "time"
)

type MetricEntry struct {
    ConsensusType string
    BlockTime     time.Duration
    Timestamp     time.Time
}

type TxEntry struct {
    ConsensusType string
    TxID          string
    Latency       time.Duration
    Timestamp     time.Time
}

type Recorder struct {
    mu          sync.Mutex
    blockTimes  []MetricEntry
    txLatencies []TxEntry
}

func NewRecorder() *Recorder {
    return &Recorder{}
}

func (r *Recorder) Record(consensus string, duration time.Duration) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.blockTimes = append(r.blockTimes, MetricEntry{
        ConsensusType: consensus,
        BlockTime:     duration,
        Timestamp:     time.Now(),
    })
}

func (r *Recorder) RecordTransactionLatency(consensus, txID string, latency time.Duration) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.txLatencies = append(r.txLatencies, TxEntry{
        ConsensusType: consensus,
        TxID:          txID,
        Latency:       latency,
        Timestamp:     time.Now(),
    })
}

func (r *Recorder) ExportCSV(path string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    writer := csv.NewWriter(f)
    defer writer.Flush()

    writer.Write([]string{"Type", "BlockTime(ms)", "Timestamp"})

    for _, e := range r.blockTimes {
        writer.Write([]string{
            e.ConsensusType,
            fmt.Sprintf("%d", e.BlockTime.Milliseconds()),
            e.Timestamp.Format(time.RFC3339),
        })
    }
    return nil
}

func (r *Recorder) ExportTransactionCSV(path string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    writer := csv.NewWriter(f)
    defer writer.Flush()

    writer.Write([]string{"Type", "TxID", "Latency(ms)", "Timestamp"})

    for _, tx := range r.txLatencies {
        writer.Write([]string{
            tx.ConsensusType,
            tx.TxID,
            fmt.Sprintf("%d", tx.Latency.Milliseconds()),
            tx.Timestamp.Format(time.RFC3339),
        })
    }
    return nil
}
