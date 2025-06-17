package core

import (
	"sync"
	"time"
)

// TransactionPool manages the pool of pending transactions
type TransactionPool struct {
	mu           sync.Mutex
	transactions []*Transaction
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: []*Transaction{},
	}
}

// AddTransaction adds a new transaction to the pool
func (tp *TransactionPool) AddTransaction(tx *Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.transactions = append(tp.transactions, tx)
}

// GetPendingTransactions returns all pending transactions and clears the pool
func (tp *TransactionPool) GetPendingTransactions() []*Transaction {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	txs := tp.transactions
	tp.transactions = []*Transaction{}
	return txs
}

// GenerateSampleTransaction generates a dummy transaction
func GenerateSampleTransaction(from string, to string, amount float64) *Transaction {
	return &Transaction{
		Hash:     generateTxHash(), // Implement a hash generator
		From:     from,
		To:       to,
		Amount:   amount,
		Fee:      0.01,
		SubmitAt: time.Now(),
	}
}
