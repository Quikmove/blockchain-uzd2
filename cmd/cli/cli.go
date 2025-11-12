package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Quikmove/blockchain-uzd2/internal/blockchain"
	"github.com/Quikmove/blockchain-uzd2/internal/config"
	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
	"github.com/Quikmove/blockchain-uzd2/internal/domain"
	"github.com/Quikmove/blockchain-uzd2/internal/filetolist"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	app := &cli.Command{
		Name:  "blockchain-cli",
		Usage: "Interact with blockchain (local or via HTTP API)",
		Commands: []*cli.Command{
			{
				Name:  "local",
				Usage: "Start an interactive blockchain session",
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg := config.LoadConfig()
					hasher := crypto.NewArchasHasher()
					log.Println("Version:", cfg.Version)
					log.Println("Difficulty:", cfg.Difficulty)
					names := filetolist.FileToList(cfg.NameListPath)
					keyGen := crypto.NewKeyGenerator()
					userGen := blockchain.NewUserGeneratorService(keyGen)
					users := userGen.GenerateUsers(names, cfg.UserCount)
					log.Println("User count:", len(users))
					log.Println("Generating genesis block...")
					txSigner := crypto.NewTransactionSigner()
					bch := blockchain.InitBlockchainWithFunds(100, 1000000, users, cfg, hasher, txSigner)
					genesis, _ := bch.GetLatestBlock()
					genesisHeader := genesis.Header
					log.Println("Found a POW hash successfully with nonce:", genesisHeader.Nonce)
					log.Println("Added genesis block successfully")

					if err != nil {
						log.Println(err)
					}
					txsSize := 100
					//AddBlocks(5, ctx, bch, users, txsSize, cfg, hasher)
					err := bch.MineBlocks(ctx, 5, txsSize, 10, 50, users, cfg.Version, cfg.Difficulty)
					if err != nil {
						log.Println("Error mining initial blocks:", err)
					}
					for {
						fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
						fmt.Println("║                    BLOCKCHAIN CLI - AVAILABLE COMMANDS                ║")
						fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
						fmt.Println("║ MINING:                                                               ║")
						fmt.Println("║   mineblocks          - Mine new blocks with random transactions      ║")
						fmt.Println("║                                                                       ║")
						fmt.Println("║ BLOCKCHAIN INFO:                                                      ║")
						fmt.Println("║   height              - Show current blockchain height                ║")
						fmt.Println("║   stats               - Show blockchain statistics                    ║")
						fmt.Println("║   validatechain       - Validate entire blockchain integrity          ║")
						fmt.Println("║                                                                       ║")
						fmt.Println("║ BLOCK QUERIES:                                                        ║")
						fmt.Println("║   getblock            - Get full block details by index               ║")
						fmt.Println("║   getblockheader      - Get block header by index                     ║")
						fmt.Println("║   getblockhash        - Get block hash by index                       ║")
						fmt.Println("║   getblocktransactions- Get block transactions by index               ║")
						fmt.Println("║   getallheaders       - Get all block headers                         ║")
						fmt.Println("║                                                                       ║")
						fmt.Println("║ USER & BALANCE:                                                       ║")
						fmt.Println("║   balance             - Show all user balances (table)                ║")
						fmt.Println("║   getuserbalance      - Get balance by name or public key             ║")
						fmt.Println("║   richlist            - Show top users by balance                     ║")
						fmt.Println("║   getutxos            - Get UTXOs by name or public key               ║")
						fmt.Println("║                                                                       ║")
						fmt.Println("║ OTHER:                                                                ║")
						fmt.Println("║   help                - Show detailed help                            ║")
						fmt.Println("║   exit                - Exit the program                              ║")
						fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
						fmt.Print("\nEnter command: ")
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
						case "height":
							height := bch.Len()
							fmt.Printf("Current blockchain height: %d\n", height)
						case "stats":
							blocks := bch.Blocks()
							totalTxs := 0
							totalDifficulty := uint32(0)
							for _, block := range blocks {
								body := block.Body
								txs := body.Transactions
								totalTxs += len(txs)
								header := block.Header
								totalDifficulty += header.Difficulty
							}
							avgTxPerBlock := 0.0
							if len(blocks) > 0 {
								avgTxPerBlock = float64(totalTxs) / float64(len(blocks))
							}

							fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
							fmt.Println("║                   BLOCKCHAIN STATISTICS                       ║")
							fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
							fmt.Printf("║ Total Blocks:              %34d ║\n", len(blocks))
							fmt.Printf("║ Total Transactions:        %34d ║\n", totalTxs)
							fmt.Printf("║ Avg Transactions/Block:    %34.2f ║\n", avgTxPerBlock)
							fmt.Printf("║ Total Users:               %34d ║\n", len(users))
							fmt.Printf("║ Current Version:           %34d ║\n", cfg.Version)
							fmt.Printf("║ Current Difficulty:        %34d ║\n", cfg.Difficulty)
							fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
						case "validatechain":
							fmt.Println("Validating blockchain...")
							valid := true
							blocks := bch.Blocks()
							for i := 1; i < len(blocks); i++ {
								prevBlock := blocks[i-1]
								currentBlock := blocks[i]

								prevHash := bch.CalculateHash(prevBlock)

								currentHeader := currentBlock.Header
								if prevHash != currentHeader.PrevHash {
									fmt.Printf("❌ Block %d: Previous hash mismatch!\n", i)
									valid = false
								}

								currentHash := bch.CalculateHash(currentBlock)

								if !blockchain.IsHashValid(currentHash, currentHeader.Difficulty) {
									fmt.Printf("❌ Block %d: Hash doesn't meet difficulty requirements!\n", i)
									valid = false
								}
							}

							if valid {
								fmt.Println("✅ Blockchain is valid!")
							} else {
								fmt.Println("❌ Blockchain validation failed!")
							}
						case "getblock":
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
							blockBytes, err := json.MarshalIndent(block, "", "  ")
							if err != nil {
								fmt.Printf("Block at index %d: %+v\n", index, block)
								continue
							}
							fmt.Printf("Block at index %d:\n%s\n", index, string(blockBytes))
						case "getblockhash":
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
							hash := bch.CalculateHash(block)

							fmt.Printf("Block Hash at index %d: %x\n", index, hash)
						case "mineblocks":
							var numBlocks int
							fmt.Println("Please enter number of blocks to mine concurrently:")
							_, err := fmt.Scanln(&numBlocks)
							if err != nil {
								fmt.Println("failed to read number, try again:", err)
								continue
							}
							var numTxs int
							fmt.Println("Please enter number of transactions per block:")
							_, err = fmt.Scanln(&numTxs)
							if err != nil {
								fmt.Println("failed to read number, try again:", err)
								continue
							}
							var minTxValue int
							fmt.Println("Please enter minimum transaction value:")
							_, err = fmt.Scanln(&minTxValue)
							if err != nil {
								fmt.Println("failed to read number, try again:", err)
								continue
							}
							var maxTxValue int
							fmt.Println("Please enter maximum transaction value:")
							_, err = fmt.Scanln(&maxTxValue)
							if err != nil {
								fmt.Println("failed to read number, try again:", err)
								continue
							}

							ctx := context.Background()
							err = bch.MineBlocks(ctx, numBlocks, numTxs, minTxValue, maxTxValue, users, cfg.Version, cfg.Difficulty)
							if err != nil {
								fmt.Println("Error mining blocks:", err)
							}
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
							body := block.Body
							txs := body.Transactions
							bodyBytes, err := json.MarshalIndent(txs, "", "  ")
							if err != nil {
								fmt.Printf("Block Transactions at index %d: %+v\n", index, bodyBytes)
								continue
							}
							fmt.Printf("Block Transactions at index %d:\n%s\n", index, string(bodyBytes))
						case "getuserbalance":
							var input string
							fmt.Println("Please enter user name or public key (hex):")
							_, err := fmt.Scanln(&input)
							if err != nil {
								fmt.Println("failed to read input, try again:", err)
								continue
							}

							var user domain.User
							var found bool

							for i := range users {
								if users[i].Name == input {
									user = users[i]
									found = true
									break
								}
							}

							if !found {
								pubKeyBytes, err := hex.DecodeString(input)
								if err != nil {
									fmt.Println("Error: input is neither a valid user name nor a valid hex public key")
									fmt.Println("Hint: Try using a user name from the 'balance' command")
									continue
								}

								if len(pubKeyBytes) > 32 {
									fmt.Println("Error: public key too long (max 32 bytes / 64 hex characters)")
									continue
								}

								for i := range users {
									userKeyBytes := users[i].PublicKey[:]
									if len(pubKeyBytes) <= len(userKeyBytes) {
										if string(userKeyBytes[:len(pubKeyBytes)]) == string(pubKeyBytes) {
											user = users[i]
											found = true
											break
										}
									}
								}

								if !found {
									fmt.Println("Error: no user found with that public key prefix")
									fmt.Println("Hint: Use the 'balance' command to see all users")
									continue
								}
							}

							if !found {
								fmt.Println("User not found")
								continue
							}

							balance := bch.GetUserBalance(user.Address())
							pubKeyHex := fmt.Sprintf("%x", user.PublicKey)

							fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════════════════════════╗")
							fmt.Printf("║ User:       %-77s ║\n", user.Name)
							fmt.Printf("║ Balance:    %-77d ║\n", balance)
							fmt.Printf("║ Public Key: %-77s ║\n", pubKeyHex)
							fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════════════════╝")
						case "getallheaders":
							headers := make([]domain.Header, 0, bch.Len())
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
							fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════════════════════════╗")
							fmt.Println("║                                     USER BALANCES                                         ║")
							fmt.Println("╠════════════════════════════════╦═══════════════╦══════════════════════════════════════════╣")
							fmt.Println("║            NAME                ║    BALANCE    ║              PUBLIC ADDRESSES            ║")
							fmt.Println("╠════════════════════════════════╬═══════════════╬══════════════════════════════════════════╣")
							for _, user := range users {
								balance := bch.GetUserBalance(user.PublicAddress)
								pubKeyHex := fmt.Sprintf("%x", user.PublicKey)
								pubKeyShort := pubKeyHex[:40]
								fmt.Printf("║ %-30s ║ %13d ║ %40s ║\n", user.Name, balance, pubKeyShort)
							}
							fmt.Println("╚════════════════════════════════╩═══════════════╩══════════════════════════════════════════╝")
						case "richlist":
							type UserBalance struct {
								Name          string
								Balance       uint32
								PublicAddress domain.PublicAddress
							}

							var userBalances []UserBalance
							for _, user := range users {
								balance := bch.GetUserBalance(user.PublicAddress)
								userBalances = append(userBalances, UserBalance{
									Name:          user.Name,
									Balance:       balance,
									PublicAddress: user.PublicAddress,
								})
							}

							for i := 0; i < len(userBalances); i++ {
								for j := i + 1; j < len(userBalances); j++ {
									if userBalances[j].Balance > userBalances[i].Balance {
										userBalances[i], userBalances[j] = userBalances[j], userBalances[i]
									}
								}
							}

							var topN int
							fmt.Println("How many top users to show?")
							_, err := fmt.Scanln(&topN)
							if err != nil || topN <= 0 {
								topN = 10
							}

							if topN > len(userBalances) {
								topN = len(userBalances)
							}

							fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════════════════════════╗")
							fmt.Printf("║                                 TOP %d RICHEST USERS                                      ║\n", topN)
							fmt.Println("╠══════╦═════════════════════════════╦═══════════════╦══════════════════════════════════════╣")
							fmt.Println("║ RANK ║           NAME              ║    BALANCE    ║           PUBLIC ADDRESS             ║")
							fmt.Println("╠══════╬═════════════════════════════╬═══════════════╬══════════════════════════════════════╣")
							for i := 0; i < topN; i++ {
								pubKeyHex := fmt.Sprintf("%x", userBalances[i].PublicAddress)
								pubKeyShort := pubKeyHex[:36]
								fmt.Printf("║  %2d  ║ %-27s ║ %13d ║ %36s ║\n",
									i+1, userBalances[i].Name, userBalances[i].Balance, pubKeyShort)
							}
							fmt.Println("╚══════╩═════════════════════════════╩═══════════════╩══════════════════════════════════════╝")
						case "getutxos":
							var input string
							fmt.Println("Please enter user name or public key (hex):")
							_, err := fmt.Scanln(&input)
							if err != nil {
								fmt.Println("failed to read input, try again:", err)
								continue
							}

							var user domain.User
							var found bool

							for i := range users {
								if users[i].Name == input {
									user = users[i]
									found = true
									break
								}
							}

							if !found {
								pubKeyBytes, err := hex.DecodeString(input)
								if err != nil {
									fmt.Println("Error: input is neither a valid user name nor a valid hex public key")
									fmt.Println("Hint: Try using a user name from the 'balance' command")
									continue
								}

								if len(pubKeyBytes) > 32 {
									fmt.Println("Error: public key too long (max 32 bytes / 64 hex characters)")
									continue
								}

								for i := range users {
									userKeyBytes := users[i].PublicKey[:]
									if len(pubKeyBytes) <= len(userKeyBytes) {
										if string(userKeyBytes[:len(pubKeyBytes)]) == string(pubKeyBytes) {
											user = users[i]
											found = true
											break
										}
									}
								}

								if !found {
									fmt.Println("Error: no user found with that public key prefix")
									fmt.Println("Hint: Use the 'balance' command to see all users and their public keys")
									continue
								}
							}

							if !found {
								fmt.Println("User not found")
								continue
							}

							utxos := bch.GetUTXOsForAddress(user.PublicAddress)

							fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════════════╗\n")
							fmt.Printf("║                            UTXOs for %-51s ║\n", user.Name)
							fmt.Println("╠══════╦═══════════════╦═══════════════════════════════════════════════════════════════════╣")
							fmt.Println("║  #   ║     VALUE     ║                    TRANSACTION ID:INDEX                           ║")
							fmt.Println("╠══════╬═══════════════╬═══════════════════════════════════════════════════════════════════╣")

							totalValue := uint32(0)
							if len(utxos) == 0 {
								fmt.Println("║                         No UTXOs found for this user                                      ║")
							} else {
								for i, utxo := range utxos {
									txIDHex := fmt.Sprintf("%x", utxo.Outpoint.TxID)
									fmt.Printf("║ %4d ║ %13d ║ %s:%-6d ║\n",
										i+1, utxo.Value, txIDHex[:58], utxo.Outpoint.Index)
									totalValue += utxo.Value
								}
							}

							fmt.Println("╠══════╩═══════════════╩═══════════════════════════════════════════════════════════════════╣")
							fmt.Printf("║ Total UTXOs: %-10d                            Total Value: %-24d ║\n",
								len(utxos), totalValue)
							fmt.Println("╚══════════════════════════════════════════════════════════════════════════════════════════╝")
						case "help":
							fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════════════════════════╗")
							fmt.Println("║                              BLOCKCHAIN CLI - HELP                                        ║")
							fmt.Println("╠═══════════════════════════════════════════════════════════════════════════════════════════╣")
							fmt.Println("║                                                                                           ║")
							fmt.Println("║ MINING COMMANDS:                                                                          ║")
							fmt.Println("║   mineblocks - Mines new blocks with random transactions between users                    ║")
							fmt.Println("║                Prompts for: number of blocks, transactions per block, min/max tx value    ║")
							fmt.Println("║                                                                                           ║")
							fmt.Println("║ BLOCKCHAIN INFO:                                                                          ║")
							fmt.Println("║   height     - Displays the current height (number of blocks) in the chain                ║")
							fmt.Println("║   stats      - Shows statistics: blocks, transactions, averages, difficulty               ║")
							fmt.Println("║   validatechain - Validates the entire blockchain integrity (checks hashes & PoW)         ║")
							fmt.Println("║                                                                                           ║")
							fmt.Println("║ BLOCK QUERIES:                                                                            ║")
							fmt.Println("║   getblock   - Get complete block data (header + transactions) by index                   ║")
							fmt.Println("║   getblockheader - Get only the block header by index                                     ║")
							fmt.Println("║   getblockhash - Get the hash of a block by index                                         ║")
							fmt.Println("║   getblocktransactions - Get all transactions in a block by index                         ║")
							fmt.Println("║   getallheaders - Get all block headers in the chain                                      ║")
							fmt.Println("║                                                                                           ║")
							fmt.Println("║ USER & BALANCE:                                                                           ║")
							fmt.Println("║   balance    - Show all users with their balances in a formatted table                    ║")
							fmt.Println("║   getuserbalance - Get balance for a user (by name or public key)                         ║")
							fmt.Println("║   richlist   - Show top N users ranked by balance                                         ║")
							fmt.Println("║   getutxos   - Show all UTXOs for a user (by name or public key)                          ║")
							fmt.Println("║                                                                                           ║")
							fmt.Println("║                                                                                           ║")
							fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════════════════╝")
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
