package blockchain

import (
	"log"
	"sync"
)

type UTXOTracker struct {
	utxoSet   map[Outpoint]UTXO
	UTXOMutex *sync.RWMutex
}

func NewUTXOTracker() *UTXOTracker {
	return &UTXOTracker{
		utxoSet:   make(map[Outpoint]UTXO),
		UTXOMutex: &sync.RWMutex{},
	}
}
func (t *UTXOTracker) reset() {
	t.UTXOMutex.Lock()
	defer t.UTXOMutex.Unlock()
	t.utxoSet = make(map[Outpoint]UTXO)
}
func (t *UTXOTracker) ScanBlockchain(bc *Blockchain) {
	blocks := bc.Blocks()
	t.reset()
	for _, block := range blocks {
		t.ScanBlock(block, bc.hasher)
	}
}
func (t *UTXOTracker) ScanBlock(b Block, hasher Hasher) {
	t.UTXOMutex.Lock()
	defer t.UTXOMutex.Unlock()

	body := b.GetBody()
	txs := body.GetTransactions()
	for _, tx := range txs {
		if len(tx.Inputs) > 0 {
			for _, input := range tx.Inputs {
				delete(t.utxoSet, input.Prev)
			}
		}

		txHash, err := tx.Hash(hasher)
		if err != nil {
			log.Printf("Error hashing transaction: %v\n. continuing...", err)
			continue
		}
		for idx, output := range tx.Outputs {
			outpoint := Outpoint{
				TxID:  txHash,
				Index: uint32(idx),
			}
			utxo := UTXO{
				Out:   outpoint,
				To:    output.To,
				Value: output.Value,
			}
			t.utxoSet[outpoint] = utxo
		}
	}
}

func (t *UTXOTracker) GetUTXO(outpoint Outpoint) (UTXO, bool) {
	t.UTXOMutex.RLock()
	defer t.UTXOMutex.RUnlock()
	utxo, exists := t.utxoSet[outpoint]
	return utxo, exists
}

func (t *UTXOTracker) GetUTXOsForAddress(address Hash32) []UTXO {
	t.UTXOMutex.RLock()
	defer t.UTXOMutex.RUnlock()

	var utxos []UTXO
	for _, utxo := range t.utxoSet {
		if utxo.To == address {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}

func (t *UTXOTracker) GetBalance(address Hash32) uint32 {
	utxos := t.GetUTXOsForAddress(address)
	var balance uint32
	for _, utxo := range utxos {
		if balance > ^uint32(0)-utxo.Value {
			return ^uint32(0)
		}
		balance += utxo.Value
	}
	return balance
}
