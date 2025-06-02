
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

type Recorder struct {
    mu      sync.Mutex
    entries []MetricEntry
}

func NewRecorder() *Recorder {
    return &Recorder{}
}

func (r *Recorder) Record(t string, duration time.Duration) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.entries = append(r.entries, MetricEntry{
        ConsensusType: t,
        BlockTime:     duration,
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

    for _, e := range r.entries {
        writer.Write([]string{
            e.ConsensusType,
            fmt.Sprintf("%d", e.BlockTime.Milliseconds()),
            e.Timestamp.Format(time.RFC3339),
        })
    }
    return nil
}
