package main

import (
	"context"
	"encoding/json"
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
					cfg := config.LoadConfig()
					log.Println("Please select an option:")
					hasher := blockchain.NewArchasHasher()
					log.Println("Version:", cfg.Version)
					log.Println("Difficulty:", cfg.Difficulty)
					names := filetolist.FileToList(cfg.NameListPath)
					users := blockchain.GenerateUsers(names, 50)
					log.Println("Generating genesis block...")
					bch := blockchain.InitBlockchainWithFunds(100, 1000000, users, cfg, hasher)
					genesis, _ := bch.GetLatestBlock()
					log.Println("Found a POW hash successfully with nonce:", genesis.Header.Nonce)
					log.Println("Added genesis block successfully")

					if err != nil {
						log.Println(err)
					}
					txsSize := 100
					AddBlocks(5, ctx, bch, users, txsSize, cfg, hasher)
					for {
						fmt.Println("-------------------------")
						fmt.Println("Available commands:")
						fmt.Println("addblocks - Add new blocks with random transactions")
						fmt.Println("getallheaders - Get all block headers")
						fmt.Println("getblockheader - Get block header by index")
						fmt.Println("getblocktransactions - Get block transactions by index")
						fmt.Println("balance - Show user balances")
						fmt.Println("exit - Exit the program")
						fmt.Println("-------------------------")
						fmt.Println("Please enter a command")
						var command string
						_, err := fmt.Scanln(&command)
						if err != nil {
							fmt.Println("failed to read command, try again:", err)
							continue
						}
						switch command {
						case "getblockheader":
							var index int
							fmt.Println("Please enter block index:")
							_, err := fmt.Scanln(&index)
							if err != nil {
								fmt.Println("failed to read index, try again:", err)
								continue
							}
							block, err := bch.GetBlockByIndex(index)
							if err != nil {
								fmt.Println("Error:", err)
								continue
							}
							headBytes, err := json.MarshalIndent(block.Header, "", "  ")
							if err != nil {
								fmt.Printf("Block Header at index %d: %+v\n", index, block.Header)
								continue
							}
							fmt.Printf("Block Header at index %d:\n%s\n", index, string(headBytes))
						case "mine":
							var numBlocks int
							fmt.Println("Please enter number of blocks to mine concurrently:")
							_, err := fmt.Scanln(&numBlocks)
							if err != nil {
								fmt.Println("failed to read number, try again:", err)
								continue
							}
							ctx := context.Background()
							err = bch.MineBlocks(ctx, users, cfg.Version, cfg.Difficulty, numBlocks)
							if err != nil {
								fmt.Println("Error mining blocks:", err)
							}
						case "addblocks":
							var numBlocks int
							fmt.Println("Please enter number of blocks to add:")
							_, err := fmt.Scanln(&numBlocks)
							if err != nil {
								fmt.Println("failed to read number, try again:", err)
								continue
							}
							if numBlocks <= 0 {
								fmt.Println("Number of blocks must be positive")
								continue
							}
							AddBlocks(numBlocks, ctx, bch, users, txsSize, cfg, hasher)

						case "getblocktransactions":
							var index int
							fmt.Println("Please enter block index:")
							_, err := fmt.Scanln(&index)
							if err != nil {
								fmt.Println("failed to read index, try again:", err)
								continue
							}
							block, err := bch.GetBlockByIndex(index)
							if err != nil {
								fmt.Println("Error:", err)
								continue
							}
							bodyBytes, err := json.MarshalIndent(block.Body.Transactions, "", "  ")
							if err != nil {
								fmt.Printf("Block Transactions at index %d: %+v\n", index, bodyBytes)
								continue
							}
						case "getallheaders":
							headers := make([]blockchain.Header, 0, bch.Len())
							for _, blk := range bch.Blocks() {
								headers = append(headers, blk.Header)
							}
							headersBytes, err := json.MarshalIndent(headers, "", "  ")
							if err != nil {
								fmt.Println("Error marshaling headers:", err)
								continue
							}
							fmt.Printf("All Block Headers:\n%s\n", string(headersBytes))
						case "balance":
							for _, user := range users {
								balance := bch.GetUserBalance(user.PublicKey)
								fmt.Printf("User %s has balance: %d\n", user.Name, balance)
							}
						case "exit":
							fmt.Println("Exiting...")
							return nil
						default:
							fmt.Println("Unknown command")
						}
					}
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalln(err)
	}
}

func AddBlocks(blockNum int, ctx context.Context, bch *blockchain.Blockchain, users []blockchain.User, txsSize int, cfg *config.Config, hasher blockchain.Hasher) {
	for range blockNum {
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
		log.Printf("Added a new block with %d transactions and nonce: %d", len(txs), newBlock.Header.Nonce)
	}
}
