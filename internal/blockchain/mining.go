package blockchain

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
	"time"

	c "github.com/Quikmove/blockchain-uzd2/internal/crypto"
	d "github.com/Quikmove/blockchain-uzd2/internal/domain"
)

func (bch *Blockchain) MineBlocks(parentCtx context.Context, blockCount, txCount, low, high int, users []d.User, version, difficulty uint32) error {
	if blockCount <= 0 {
		return nil
	}

	numWorkers := runtime.NumCPU()

	for round := 0; round < blockCount; round++ {
		if parentCtx.Err() != nil {
			return parentCtx.Err()
		}

		roundCtx, cancelRound := context.WithCancel(parentCtx)

		blockChan := make(chan d.Block, 1)

		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(ctx context.Context, workerID int) {
				defer wg.Done()
				for {
					if ctx.Err() != nil {
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

					body := d.NewBody(txs)

					blk, err := bch.generateBlockWithTimestamp(ctx, *body, version, difficulty, uint32(time.Now().Unix())+uint32(workerID))
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

					select {
					case blockChan <- blk:
					default:
					}
					return
				}
			}(roundCtx, i)
		}

		select {
		case blk := <-blockChan:
			blockIndex := bch.Len() - 1
			header := blk.Header
			blockBody := blk.Body
			blockTxs := blockBody.Transactions
			log.Printf("Round %d: Successfully mined block at index %d (block #%d) with %d transactions and nonce %d\n", round+1, blockIndex, round+1, len(blockTxs), header.Nonce)
		case <-parentCtx.Done():
			cancelRound()
			wg.Wait()
			return parentCtx.Err()
		}

		cancelRound()

		wg.Wait()
	}

	if parentCtx.Err() != nil {
		return parentCtx.Err()
	}
	return nil
}

func (bch *Blockchain) generateBlockWithTimestamp(ctx context.Context, body d.Body, version uint32, difficulty uint32, timestamp uint32) (d.Block, error) {
	latestBlock, err := bch.GetLatestBlock()
	if err != nil {
		return d.Block{}, err
	}
	var newHeader d.Header

	newHeader.Version = version
	newHeader.Timestamp = timestamp
	newHeader.PrevHash = bch.CalculateHash(latestBlock)
	newHeader.MerkleRoot = MerkleRootHash(body, bch.hasher)
	newHeader.Difficulty = difficulty

	nonce, _, err := FindValidNonce(ctx, &newHeader, bch.hasher)
	if err != nil {
		return d.Block{}, err
	}
	newHeader.Nonce = nonce

	newBlock := d.Block{
		Header: newHeader,
		Body:   body,
	}

	return newBlock, nil
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
