package blockchain

import (
	"math/rand"
	"sort"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func GenerateFundTransactionsForUsers(users []d.User, low, high uint32, hasher c.Hasher) (Transactions, error) {
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
		var utxos []uint32
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
		var outputs []d.TxOutput
		for _, v := range utxos {
			txOut := d.TxOutput{
				Value: v,
				To:    usr.PublicKey,
			}
			outputs = append(outputs, txOut)
		}
		tx := d.Transaction{
			Inputs:  nil,
			Outputs: outputs,
		}
		hash := hasher.Hash(tx.Serialize())

		tx.TxID = hash
		txs = append(txs, tx)
	}
	return txs, nil
}
