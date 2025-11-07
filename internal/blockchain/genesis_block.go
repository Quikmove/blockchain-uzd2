package blockchain

import (
	"context"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
)

func CreateGenesisBlock(ctx context.Context, txs Transactions, conf *config.Config, hasher Hasher) (Block, error) {
	t := time.Now()
	merkleRoot := merkleRootHash(txs, hasher)
	header := NewHeader(
		conf.Version,
		uint32(t.Unix()),
		Hash32{},
		merkleRoot,
		conf.Difficulty,
		0,
	)
	body := NewBody(txs)
	genesisBlock := NewBlock(header, body)
	nonce, _, err := genesisBlock.GetHeader().FindValidNonce(ctx, hasher)
	if err != nil {
		return Block{}, err
	}
	header.SetNonce(nonce)
	genesisBlock.SetHeader(header)
	return genesisBlock, nil
}
