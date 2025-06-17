
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "lscc/config"
    "lscc/consensus"
    "lscc/core"
    "lscc/metrics"
)

func main() {
    cfg := config.LoadConfig("config.json")
    txPool := core.NewTransactionPool()

    // Generate deterministic transactions
    txs := core.GenerateDeterministicTransactions(12345, 1000)
    for _, tx := range txs {
        txPool.Add(tx)
    }

    recorder := metrics.NewRecorder()

    engines := []consensus.ConsensusEngine{
        consensus.NewPoSConsensus(cfg, core.NewBlockchain(cfg, txPool)),
        consensus.NewPoWConsensus(cfg, core.NewBlockchain(cfg, txPool, recorder)),
        consensus.NewPBFTConsensus(cfg, core.NewBlockchain(cfg, txPool)),
        consensus.NewCrossChannelConsensus(cfg, core.NewBlockchain(cfg, txPool)),
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

    recorder.ExportCSV("results.csv")
    fmt.Println("Results exported to results.csv")
}
