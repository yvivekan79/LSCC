
# LSCC Blockchain Benchmarking Suite

This project benchmarks and compares four consensus algorithms in parallel:

- 🧠 **Proof of Stake (PoS)**
- ⛏️ **Proof of Work (PoW)** (Multithreaded)
- 🛡️ **PBFT (Practical Byzantine Fault Tolerance)**
- 🔗 **Cross-Channel Consensus**

Each engine runs independently using a shared transaction pool and logs its performance.

---

## 📦 Project Structure

```
LSCC/
├── main.go                      # Orchestrates all 4 consensus engines
├── core/
│   └── txpool.go                # Shared transaction pool + Transaction struct
├── consensus/
│   └── pow.go                   # Multithreaded PoW implementation
├── metrics/
│   └── recorder.go              # Metrics recording (block time + tx latency)
├── results_block_latency.csv    # Output: block-level latency
├── results_tx_latency.csv       # Output: per-transaction latency
└── Makefile                     # Build & run targets
```

---

## 🚀 How to Run

### 🛠️ Build

```bash
make build
```

### ▶️ Run Default Benchmark

```bash
make run
```

### 🧪 Run Enhanced Benchmark (Live TX Feed + Metrics)

```bash
make run-enhanced
```

### 🧼 Clean

```bash
make clean
```

---

## 📈 Metrics Collected

| Metric              | Description                                 |
|---------------------|---------------------------------------------|
| Block Latency       | Time taken to create/commit each block      |
| Transaction Latency | Time from submission to inclusion in block  |
| TPS                 | Transactions per second (logged to console) |

CSV outputs:
- `results_block_latency.csv`
- `results_tx_latency.csv`

---

## 🔬 Configuration

Modify `config.json` to change:
- Consensus parameters (PoS stake, PoW difficulty, etc.)
- Logging level
- Network settings (optional)

---

## 📊 Visualize Results

You can load the CSV outputs into:
- 📊 Excel or Google Sheets
- 📈 Grafana (CSV-to-Prometheus exporter)
- 📘 Python (pandas + matplotlib)

---

## 📥 Contributing

Feel free to fork, experiment, and submit PRs to improve benchmarking accuracy, visualization, or consensus implementations.

---

## 🧠 Authors

Built to evaluate consensus performance in LSCC-based blockchains for research, education, and system design.
