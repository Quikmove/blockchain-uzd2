package blockchain

import "sync"

type UTXOTracker struct {
	UTXOSet   map[Outpoint]UTXO
	UTXOMutex *sync.RWMutex
}

func NewUTXOTracker() *UTXOTracker {
	return &UTXOTracker{
		UTXOSet:   make(map[Outpoint]UTXO),
		UTXOMutex: &sync.RWMutex{},
	}
}

func (t *UTXOTracker) ScanBlock(b Block) {
	t.UTXOMutex.Lock()
	defer t.UTXOMutex.Unlock()

	for _, tx := range b.Body.Transactions {
		if len(tx.Inputs) > 0 {
			for _, input := range tx.Inputs {
				delete(t.UTXOSet, input.Prev)
			}
		}

		txHash := tx.Hash()
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
			t.UTXOSet[outpoint] = utxo
		}
	}
}

func (t *UTXOTracker) GetUTXO(outpoint Outpoint) (UTXO, bool) {
	t.UTXOMutex.RLock()
	defer t.UTXOMutex.RUnlock()
	utxo, exists := t.UTXOSet[outpoint]
	return utxo, exists
}

func (t *UTXOTracker) GetUTXOsForAddress(address Hash32) []UTXO {
	t.UTXOMutex.RLock()
	defer t.UTXOMutex.RUnlock()

	var utxos []UTXO
	for _, utxo := range t.UTXOSet {
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
		balance += utxo.Value
	}
	return balance
}
