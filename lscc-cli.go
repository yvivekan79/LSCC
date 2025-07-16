package main

import (
    "bytes"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "net/http"
    "os"
    "time"
)

type Transaction struct {
    From      string  `json:"from"`
    To        string  `json:"to"`
    Amount    float64 `json:"amount"`
    Timestamp int64   `json:"timestamp"`
    Hash      string  `json:"hash"`
}

func main() {
    from := flag.String("from", "", "Sender address")
    to := flag.String("to", "", "Receiver address")
    amount := flag.Float64("amount", 0.0, "Amount to send")
    port := flag.Int("port", 9000, "Port of REST API")
    flag.Parse()

    if *from == "" || *to == "" || *amount <= 0 {
        fmt.Println("Usage: lscc-cli -from Alice -to Bob -amount 10 -port 9000")
        os.Exit(1)
    }

    tx := Transaction{
        From:      *from,
        To:        *to,
        Amount:    *amount,
        Timestamp: time.Now().Unix(),
    }

    data := fmt.Sprintf("%s:%s:%f:%d", tx.From, tx.To, tx.Amount, tx.Timestamp)
    hash := sha256.Sum256([]byte(data))
    tx.Hash = hex.EncodeToString(hash[:])

    jsonData, err := json.Marshal(tx)
    if err != nil {
        fmt.Println("Failed to marshal transaction:", err)
        os.Exit(1)
    }

    url := fmt.Sprintf("http://localhost:%d/send", *port)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("POST failed:", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    fmt.Println("Transaction sent:", resp.Status)
}

