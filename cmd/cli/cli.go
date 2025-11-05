package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/blockchain"
	"github.com/Quikmove/blockchain-uzd2/internal/config"
	"github.com/Quikmove/blockchain-uzd2/internal/filetolist"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// (&cli.Command{}).Run(context.Background(), os.Args)
	//choose to start an api or to connect to running api
	//therefore need to determine how to check if api endpoint is valid
	app := &cli.Command{
		Name:  "blockchain-cli",
		Usage: "Interact with blockchain (local or via HTTP API)",
		Commands: []*cli.Command{
			//{
			//	Name:  "serve",
			//	Usage: "Start HTTP API server",
			//	Action: func(context.Context, *cli.Command) error {
			//		cfg := config.LoadConfig()
			//		names := filetolist.FileToList(cfg.NameListPath)
			//		users := blockchain.GenerateUsers(names, 3)
			//		bch := blockchain.InitBlockchainWithFunds(100, 1000000, users, cfg, blockchain.NewArchasHasher())
			//		ctx := context.Background()
			//		if err := api.Run(ctx, bch, cfg, nil); err != nil {
			//			log.Fatalln(err)
			//		}
			//		return nil
			//	},
			//},
			{
				Name:  "local",
				Usage: "Start an interactive blockchain session",
				Action: func(ctx context.Context, c *cli.Command) error {
					hasher := blockchain.NewArchasHasher()
					cfg := config.LoadConfig()
					log.Println("Version:", cfg.Version)
					log.Println("Difficulty:", cfg.Difficulty)
					names := filetolist.FileToList(cfg.NameListPath)
					users := blockchain.GenerateUsers(names, 50)
					log.Println("Generating genesis block...")
					bch := blockchain.InitBlockchainWithFunds(100, 1000000, users, cfg, hasher)
					genesis, _ := bch.GetLatestBlock()
					log.Println("Found a POW hash successfully with nonce:", genesis.Header.Nonce)
					if err != nil {
						log.Println(err)
					}
					txsSize := 100
					//totalTxs := 10_000
					//blocksToAdd := totalTxs / txsSize
					for range 5 {
						txs, err := bch.GenerateRandomTransactions(users, 10, 50, txsSize)
						if err != nil {
							log.Println("Error generating transactions:", err)
							continue
						}
						latestBlock, err := bch.GetLatestBlock()
						if err != nil {
							log.Println("Error getting latest block:", err)
							continue
						}
						hash, err := bch.CalculateHash(latestBlock)
						if err != nil {
							log.Println("Error calculating previous block hash:", err)
							continue
						}
						var header blockchain.Header
						var body blockchain.Body
						body.Transactions = txs
						header.Version = cfg.Version
						header.MerkleRoot = body.MerkleRootHash(hasher)
						header.PrevHash = hash
						header.Timestamp = uint32(time.Now().Unix())
						header.Difficulty = cfg.Difficulty
						nonce, _, err := header.FindValidNonce(ctx, hasher)
						if err != nil {
							log.Println("Error finding valid nonce:", err)
							continue
						}
						header.Nonce = nonce
						newBlock := blockchain.Block{
							Header: header,
							Body:   body,
						}
						if err := bch.AddBlock(newBlock); err != nil {
							log.Println("Error adding new block to blockchain:", err)
							continue
						}
						log.Println("Added new block with nonce:", newBlock.Header.Nonce)
					}
					return nil
				},
			},
			{
				Name:  "blocks",
				Usage: "Fetch and display blockchain blocks",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println("Not implemented yet")
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalln(err)
	}
}
