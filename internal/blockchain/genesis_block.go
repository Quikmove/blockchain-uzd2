package blockchain

import (
	"context"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
)

func CreateGenesisBlock(ctx context.Context, txs Transactions, conf *config.Config) (Block, error) {
	t := time.Now()
	merkleRoot := MerkleRootHash(txs)
	genesisBlock := Block{
		Header: Header{
			Version:    conf.Version,
			Timestamp:  uint32(t.Unix()),
			PrevHash:   Hash32{},
			MerkleRoot: merkleRoot,
			Difficulty: conf.Difficulty,
			Nonce:      0,
		},
		Body: Body{
			Transactions: txs,
		},
	}
	nonce, _, err := genesisBlock.Header.FindValidNonce(ctx)
	if err != nil {
		return Block{}, err
	}
	genesisBlock.Header.Nonce = nonce
	return genesisBlock, nil
}
