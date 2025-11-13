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

type DecentralizedMiningConfig struct {
	BlockCount       int
	TxCount          int
	CandidateCount   int
	InitialTimeLimit time.Duration
	TimeMultiplier   float64
	Low              int
	High             int
	Version          uint32
	Difficulty       uint32
}

func DefaultDecentralizedMiningConfig() DecentralizedMiningConfig {
	return DecentralizedMiningConfig{
		BlockCount:       1,
		TxCount:          100,
		CandidateCount:   5,
		InitialTimeLimit: 5 * time.Second,
		TimeMultiplier:   2.0,
		Low:              1,
		High:             1000,
		Version:          1,
		Difficulty:       1,
	}
}

type blockResult struct {
	block    d.Block
	workerId int
}

func (bch *Blockchain) MineBlocksDecentralized(
	parentCtx context.Context,
	users []d.User,
	config DecentralizedMiningConfig,
) error {
	if config.BlockCount <= 0 {
		return nil
	}

	for round := 0; round < config.BlockCount; round++ {
		if err := parentCtx.Err(); err != nil {
			return err
		}

		timeLimit := config.InitialTimeLimit
		miningSuccess := false

		for !miningSuccess {
			if err := parentCtx.Err(); err != nil {
				return err
			}

			candidateBlocks := make([]d.Body, 0, config.CandidateCount)
			for i := 0; i < config.CandidateCount; i++ {
				txs, err := bch.GenerateRandomTransactions(users, config.Low, config.High, config.TxCount)
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
				return d.ErrMiningCanceled
			}

			roundCtx, cancelRound := context.WithCancel(parentCtx)
			timeoutCtx := roundCtx
			var timeoutCancel context.CancelFunc

			if timeLimit > 0 {
				timeoutCtx, timeoutCancel = context.WithTimeout(roundCtx, timeLimit)
			}

			blockResultChan := make(chan blockResult, 1)

			var wg sync.WaitGroup

			for i, body := range candidateBlocks {
				wg.Add(1)
				go func(workerId int, b d.Body) {
					defer wg.Done()
					_ = MineBlockConcurrently(timeoutCtx, workerId, bch, b, config.Version, config.Difficulty, blockResultChan)
				}(i, body)
			}

			select {
			case blkResult := <-blockResultChan:
				cancelRound()
				if timeoutCancel != nil {
					timeoutCancel()
				}

				wg.Wait()

				err := bch.AddBlock(blkResult.block)
				if err != nil {
					log.Printf("Round %d: Failed to add block: %v", round+1, err)
					timeLimit = time.Duration(float64(timeLimit) * config.TimeMultiplier)
					continue
				}
				blockIndex := bch.Len() - 1
				blockTxs := blkResult.block.Body.Transactions
				log.Printf("Round %d: Successfully mined block at index %d (block #%d) with %d transactions and nonce %d by worker %d\n",
					round+1, blockIndex, round+1, len(blockTxs), blkResult.block.Header.Nonce, blkResult.workerId)
				miningSuccess = true
			case <-timeoutCtx.Done():
				cancelRound()
				if timeoutCancel != nil {
					timeoutCancel()
				}
				wg.Wait()
				timeLimit = time.Duration(float64(timeLimit) * config.TimeMultiplier)
				log.Printf("Round %d: No block mined within time limit, retrying with increased time limit: %v", round+1, timeLimit)
			case <-parentCtx.Done():
				cancelRound()
				if timeoutCancel != nil {
					timeoutCancel()
				}
				wg.Wait()
				return parentCtx.Err()
			}
		}
	}

	return nil
}
func MineBlockConcurrently(ctx context.Context, workerId int, bch *Blockchain, body d.Body, version, difficulty uint32, mineChan chan blockResult) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		latestBlock, err := bch.GetLatestBlock()
		if err != nil {
			errChan <- err
			return
		}
		prevBlockHash := bch.CalculateHash(latestBlock)
		header := d.NewHeader(version, uint32(time.Now().Unix()), prevBlockHash, MerkleRootHash(body, bch.hasher), difficulty, 0)
		_, _, err = FindValidNonce(ctx, header, bch.hasher)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
		result := blockResult{block: *d.NewBlock(*header, body), workerId: workerId}
		select {

		case mineChan <- result:
		case <-ctx.Done():
			return
		}
	}(errChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		close(errChan)
		return err
	}
}
