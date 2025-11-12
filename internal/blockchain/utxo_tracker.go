package blockchain

import (
	"sync"

	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

type UTXOTracker struct {
	utxoSet   map[d.Outpoint]d.UTXO
	UTXOMutex *sync.RWMutex
}

func NewUTXOTracker() *UTXOTracker {
	return &UTXOTracker{
		utxoSet:   make(map[d.Outpoint]d.UTXO),
		UTXOMutex: &sync.RWMutex{},
	}
}
func (t *UTXOTracker) reset() {
	t.UTXOMutex.Lock()
	defer t.UTXOMutex.Unlock()
	t.utxoSet = make(map[d.Outpoint]d.UTXO)
}
func (t *UTXOTracker) ScanBlockchain(bc *Blockchain) {
	blocks := bc.Blocks()
	t.reset()
	for _, block := range blocks {
		t.ScanBlock(block, bc.hasher)
	}
}
func (t *UTXOTracker) ScanBlock(b d.Block, hasher crypto.Hasher) {
	t.UTXOMutex.Lock()
	defer t.UTXOMutex.Unlock()

	body := b.Body
	txs := body.Transactions
	for _, tx := range txs {
		if len(tx.Inputs) > 0 {
			for _, input := range tx.Inputs {
				delete(t.utxoSet, input.Prev)
			}
		}

		txHash := hasher.Hash(tx.Serialize())

		for idx, output := range tx.Outputs {
			outpoint := d.Outpoint{
				TxID:  txHash,
				Index: uint32(idx),
			}
			utxo := d.UTXO{
				Outpoint: outpoint,
				To:       output.To,
				Value:    output.Value,
			}
			t.utxoSet[outpoint] = utxo
		}
	}
}

func (t *UTXOTracker) GetUTXO(outpoint d.Outpoint) (d.UTXO, bool) {
	t.UTXOMutex.RLock()
	defer t.UTXOMutex.RUnlock()
	utxo, exists := t.utxoSet[outpoint]
	return utxo, exists
}

func (t *UTXOTracker) GetUTXOsForAddress(address d.PublicAddress) []d.UTXO {
	t.UTXOMutex.RLock()
	defer t.UTXOMutex.RUnlock()

	var utxos []d.UTXO
	for _, utxo := range t.utxoSet {
		if utxo.To == address {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}

func (t *UTXOTracker) GetBalance(address d.PublicAddress) uint32 {
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
