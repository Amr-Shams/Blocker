package blockchain

import (
	"fmt"
	"sync"
)

type MemPool struct {
	txs     map[string]*Transaction
	mu      sync.RWMutex
	maxSize int
}

func NewMemPool(maxSize int) *MemPool {
	return &MemPool{
		txs:     make(map[string]*Transaction),
		maxSize: maxSize,
	}
}

func (mp *MemPool) AddTx(tx *Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if len(mp.txs) >= mp.maxSize {
		return fmt.Errorf("pool full")
	}

	txID := fmt.Sprintf("%x", tx.ID)
	if _, exists := mp.txs[txID]; exists {
		return fmt.Errorf("tx exists")
	}

	mp.txs[txID] = tx
	fmt.Println("Added tx:", txID)
	return nil
}

func (mp *MemPool) GetTx() *Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	for _, tx := range mp.txs {
		return tx
	}
	return nil
}

func (mp *MemPool) DeleteTx(txID []byte) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	txIDStr := fmt.Sprintf("%x", txID)
	delete(mp.txs, txIDStr)
}

func (mp *MemPool) GetAllTxs() []*Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	txs := make([]*Transaction, 0, len(mp.txs))
	for _, tx := range mp.txs {
		txs = append(txs, tx)
	}
	return txs
}

func (mp *MemPool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.txs = make(map[string]*Transaction)
}

func (mp *MemPool) Size() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return len(mp.txs)
}
