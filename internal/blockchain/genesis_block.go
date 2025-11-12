package blockchain

import (
	"context"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func CreateGenesisBlock(ctx context.Context, txs Transactions, conf *config.Config, hasher c.Hasher) (d.Block, error) {
	t := time.Now()
	merkleRoot := merkleRootHash(txs, hasher)
	header := d.NewHeader(
		conf.Version,
		uint32(t.Unix()),
		d.Hash32{},
		merkleRoot,
		conf.Difficulty,
		0,
	)
	body := d.NewBody(txs)
	genesisBlock := d.NewBlock(*header, *body)
	nonce, _, err := FindValidNonce(ctx, &genesisBlock.Header, hasher)
	if err != nil {
		return d.Block{}, err
	}
	header.Nonce = nonce
	genesisBlock.Header = *header
	return *genesisBlock, nil
}
