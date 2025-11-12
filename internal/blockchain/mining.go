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

func (bch *Blockchain) MineBlocksDecentralized(
	parentCtx context.Context,
	blockCount, txCount, candidateCount int,
	initialTimeLimit time.Duration,
	initialAttemptLimit uint64,
	timeMultiplier, attemptMultiplier float64,
	low, high int,
	users []d.User,
	version, difficulty uint32,
) error {
	if blockCount <= 0 {
		return nil
	}

	for round := 0; round < blockCount; round++ {
		if parentCtx.Err() != nil {
			return parentCtx.Err()
		}

		timeLimit := initialTimeLimit
		attemptLimit := initialAttemptLimit
		var miningSuccess bool

		for !miningSuccess {
			if parentCtx.Err() != nil {
				return parentCtx.Err()
			}

			candidateBlocks := make([]d.Body, 0, candidateCount)
			for i := 0; i < candidateCount; i++ {
				txs, err := bch.GenerateRandomTransactions(users, low, high, txCount)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return err
					}
					log.Printf("Warning: failed to generate transactions for candidate %d: %v", i, err)
					continue
				}

				if len(txs) == 0 {
					continue
				}

				body := d.NewBody(txs)
				candidateBlocks = append(candidateBlocks, *body)
			}

			if len(candidateBlocks) == 0 {
				return errors.New("could not generate any candidate blocks")
			}

			roundCtx, cancelRound := context.WithCancel(parentCtx)
			defer cancelRound()

			var timeoutCtx context.Context
			var timeoutCancel context.CancelFunc
			if timeLimit > 0 {
				timeoutCtx, timeoutCancel = context.WithTimeout(roundCtx, timeLimit)
			} else {
				timeoutCtx, timeoutCancel = context.WithCancel(roundCtx)
			}
			defer timeoutCancel()

			blockChan := make(chan d.Block, 1)
			var wg sync.WaitGroup

			for i, body := range candidateBlocks {
				wg.Add(1)
				go func(candidateID int, candidateBody d.Body) {
					defer wg.Done()

					latestBlock, err := bch.GetLatestBlock()
					if err != nil {
						return
					}

					var newHeader d.Header
					newHeader.Version = version
					newHeader.Timestamp = uint32(time.Now().Unix())
					newHeader.PrevHash = bch.CalculateHash(latestBlock)
					newHeader.MerkleRoot = MerkleRootHash(candidateBody, bch.hasher)
					newHeader.Difficulty = difficulty

					var nonce uint32
					var attempts uint64

					for {
						select {
						case <-timeoutCtx.Done():
							return
						default:
						}

						if attemptLimit > 0 && attempts >= attemptLimit {
							return
						}

						newHeader.Nonce = nonce
						hash := bch.hasher.Hash(newHeader.Serialize())
						if IsHashValid(hash, difficulty) {
							newBlock := d.Block{
								Header: newHeader,
								Body:   candidateBody,
							}

							err := bch.AddBlock(newBlock)
							if err != nil {
								return
							}

							select {
							case blockChan <- newBlock:
							default:
							}
							return
						}

						nonce++
						attempts++

						if nonce == ^uint32(0) {
							return
						}
					}
				}(i, body)
			}

			select {
			case blk := <-blockChan:
				miningSuccess = true
				blockIndex := bch.Len() - 1
				blockTxs := blk.Body.Transactions
				log.Printf("Round %d: Successfully mined block at index %d (block #%d) with %d transactions and nonce %d\n",
					round+1, blockIndex, round+1, len(blockTxs), blk.Header.Nonce)
			case <-timeoutCtx.Done():
				if !miningSuccess {
					timeLimit = time.Duration(float64(timeLimit) * timeMultiplier)
					log.Printf("Round %d: No block mined within time limit, retrying with increased time limit: %v", round+1, timeLimit)
				}
			case <-parentCtx.Done():
				wg.Wait()
				return parentCtx.Err()
			}

			wg.Wait()
		}
	}

	return nil
}
