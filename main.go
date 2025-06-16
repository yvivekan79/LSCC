
package main

import (
    "fmt"
    "math/rand"
    "os"
    "os/signal"
    "syscall"
    "time"

    "lscc/config"
    "lscc/consensus"
    "lscc/core"
    "lscc/metrics"
)

var txCount int

func main() {
    cfg := config.LoadConfig("config.json")
    txPool := core.NewTransactionPool()
    recorder := metrics.NewRecorder()

    // Start live transaction feeder: 50 tx/sec for 2 minutes
    go StartTransactionFeeder(txPool, 50, 2*time.Minute)

    // Start TPS monitor
    go MonitorTPS()

    // Initialize and start consensus engines
    engines := []consensus.ConsensusEngine{
        consensus.NewPoSConsensus(cfg, core.NewBlockchain(cfg, txPool, recorder)),
        consensus.NewPoWConsensus(cfg, core.NewBlockchain(cfg, txPool, recorder)),
        consensus.NewPBFTConsensus(cfg, core.NewBlockchain(cfg, txPool, recorder)),
        consensus.NewCrossChannelConsensus(cfg, core.NewBlockchain(cfg, txPool, recorder)),
    }

    for _, engine := range engines {
        go engine.Start()
    }

    fmt.Println("All consensus engines started...")

    // Wait for termination
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    for _, engine := range engines {
        engine.Stop()
    }

    recorder.ExportCSV("results_block_latency.csv")
    recorder.ExportTransactionCSV("results_tx_latency.csv")
    fmt.Println("Results exported: results_block_latency.csv, results_tx_latency.csv")
}

func StartTransactionFeeder(pool *core.TransactionPool, rate int, duration time.Duration) {
    ticker := time.NewTicker(time.Second / time.Duration(rate))
    timeout := time.After(duration)
    id := 0

    for {
        select {
        case <-ticker.C:
            tx := &core.Transaction{
                ID:       fmt.Sprintf("tx-%d", id),
                Sender:   fmt.Sprintf("user-%d", rand.Intn(10)),
                Receiver: fmt.Sprintf("user-%d", rand.Intn(10)),
                Amount:   float64(rand.Intn(100)),
                SubmitAt: time.Now(),
            }
            pool.Add(tx)
            id++
        case <-timeout:
            return
        }
    }
}

func MonitorTPS() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        fmt.Printf("[TPS] Last second: %d txs\n", txCount)
        txCount = 0
    }
}
