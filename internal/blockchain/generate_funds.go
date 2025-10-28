package blockchain

import (
	"math/rand"
	"sort"
)

// go through users and select a random amount between 100 and 1000 to get transferred as UTXOs in exponential sizes: 1, 2, 4, 8, 16, etc.
// afterwards, construct the genesis block - that's the only block with coinbase-like transactions

func GenerateFundTransactionsForUsers(users []User, low, high uint32) (Transactions, error) {
	var txs Transactions
	for _, usr := range users {
		var amount uint32
		if high <= low {
			amount = low
		} else {
			delta := int(high - low + 1)
			if delta <= 0 {
				amount = low
			} else {
				amount = low + uint32(rand.Intn(delta))
			}
		}
		utxos := []uint32{}
		remaining := amount
		size := uint32(1)
		for remaining > 0 {
			if remaining >= size {
				utxos = append(utxos, size)
				remaining -= size
				size *= 2
			} else {
				utxos = append(utxos, remaining)
				remaining = 0
			}
		}
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i] < utxos[j]
		})
		var outputs []TxOutput
		for _, v := range utxos {
			txOut := TxOutput{
				Value: v,
				To:    usr.PublicKey,
			}
			outputs = append(outputs, txOut)
		}
		tx := Transaction{
			Inputs:  nil,
			Outputs: outputs,
		}
		tx.TxID = tx.Hash()
		txs = append(txs, tx)
	}
	return txs, nil
}
