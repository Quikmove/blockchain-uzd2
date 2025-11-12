package blockchain

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func (bch *Blockchain) MineBlocks(parentCtx context.Context, blockCount, txCount, low, high int, users []d.User, version, difficulty uint32) error {
	if blockCount <= 0 {
		return nil
	}

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	totalMined := atomic.Int64{}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	for i := range numWorkers {
		go func(ctx context.Context, workerID int) {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				if totalMined.Load() >= int64(blockCount) {
					return
				}
				txs, err := bch.GenerateRandomTransactions(users, low, high, txCount)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					continue
				}

				if len(txs) == 0 {
					continue
				}

				if len(txs) < txCount {
				}

				body := d.NewBody(txs)
				blk, err := bch.GenerateBlock(ctx, *body, version, difficulty)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					continue
				}
				err = bch.AddBlock(blk)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					continue
				}
				mined := totalMined.Add(1)
				blockIndex := bch.Len() - 1
				header := blk.Header
				blockBody := blk.Body
				blockTxs := blockBody.Transactions
				log.Printf("Worker %d mined block at index %d (block #%d) with %d transactions and nonce %d\n", workerID, blockIndex, mined, len(blockTxs), header.Nonce)
				if mined >= int64(blockCount) {
					cancel()
					return
				}
			}

		}(ctx, i)
	}
	wg.Wait()

	if parentCtx.Err() != nil {
		return parentCtx.Err()
	}
	return nil
}

func FindValidNonce(ctx context.Context, header *d.Header, hasher c.Hasher) (uint32, d.Hash32, error) {

	if header.Difficulty == 0 {
		hash := hasher.Hash(header.Serialize())
		return header.Nonce, hash, nil
	}
	if header.MerkleRoot.IsZero() {
		return 0, d.Hash32{}, d.ErrInvalidMerkleRoot
	}

	var nonce uint32

	for {
		if err := ctx.Err(); err != nil {
			return 0, d.Hash32{}, err
		}
		header.Nonce = nonce
		hash := hasher.Hash(header.Serialize())
		if IsHashValid(hash, header.Difficulty) {
			return nonce, hash, nil
		}
		nonce++

		if nonce == ^uint32(0) {
			return 0, d.Hash32{}, d.ErrNoValidNonce
		}
	}
}
