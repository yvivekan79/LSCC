
package core

import (
    "fmt"
    "math/rand"
    "sync"
)

type Transaction struct {
    ID       string
    Sender   string
    Receiver string
    Amount   float64
}

type TransactionPool struct {
    txs []*Transaction
    mu  sync.RWMutex
}

func NewTransactionPool() *TransactionPool {
    return &TransactionPool{txs: []*Transaction{}}
}

func (p *TransactionPool) Add(tx *Transaction) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.txs = append(p.txs, tx)
}

func (p *TransactionPool) GetAll() []*Transaction {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.txs
}

func GenerateDeterministicTransactions(seed int64, count int) []*Transaction {
    rand.Seed(seed)
    var txs []*Transaction
    for i := 0; i < count; i++ {
        tx := &Transaction{
            ID:       fmt.Sprintf("tx-%d", i),
            Sender:   fmt.Sprintf("user-%d", rand.Intn(10)),
            Receiver: fmt.Sprintf("user-%d", rand.Intn(10)),
            Amount:   float64(rand.Intn(100)),
        }
        txs = append(txs, tx)
    }
    return txs
}
