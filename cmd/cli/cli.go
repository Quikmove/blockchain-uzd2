package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/blockchain"
	"github.com/Quikmove/blockchain-uzd2/internal/config"
	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
	"github.com/Quikmove/blockchain-uzd2/internal/domain"
	"github.com/Quikmove/blockchain-uzd2/internal/filetolist"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

func validateBlockIndex(index int, maxHeight int) error {
	if index < 0 || index >= maxHeight {
		return domain.ErrBlockIndexOutOfRange
	}
	return nil
}

func validatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

func validateTransactionValueRange(min, max int) error {
	if min < 0 {
		return errors.New("minimum transaction value must be non-negative")
	}
	if max < 0 {
		return errors.New("maximum transaction value must be non-negative")
	}
	if min > max {
		return errors.New("minimum transaction value cannot exceed maximum")
	}
	return nil
}

func readInt(prompt string, validator func(int) error) (int, error) {
	fmt.Println(prompt)
	var value int
	_, err := fmt.Scanln(&value)
	if err != nil {
		return 0, fmt.Errorf("failed to read number, try again: %w", err)
	}
	if validator != nil {
		if err := validator(value); err != nil {
			return 0, err
		}
	}
	return value, nil
}

func readIntWithDefault(prompt string, defaultValue int, validator func(int) error) (int, error) {
	fmt.Println(prompt)
	var value int
	_, err := fmt.Scanln(&value)
	if err != nil {
		return defaultValue, nil
	}
	if validator != nil {
		if err := validator(value); err != nil {
			return defaultValue, nil
		}
	}
	return value, nil
}

func readString(prompt string) (string, error) {
	fmt.Println(prompt)
	var value string
	_, err := fmt.Scanln(&value)
	if err != nil {
		return "", fmt.Errorf("failed to read input, try again: %w", err)
	}
	return value, nil
}

func readCommand() (string, error) {
	fmt.Print("\nEnter command: ")
	var command string
	_, err := fmt.Scanln(&command)
	if err != nil {
		return "", fmt.Errorf("failed to read command, try again: %w", err)
	}
	return command, nil
}

func getBlockByIndex(bch *blockchain.Blockchain, prompt string) (domain.Block, int, error) {
	index, err := readInt(prompt, func(idx int) error {
		maxHeight := bch.Len()
		return validateBlockIndex(idx, maxHeight)
	})
	if err != nil {
		return domain.Block{}, 0, err
	}
	block, err := bch.GetBlockByIndex(index)
	if err != nil {
		return domain.Block{}, 0, fmt.Errorf("error retrieving block: %w", err)
	}
	return block, index, nil
}

func marshalJSON(data interface{}, fallback func()) ([]byte, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		if fallback != nil {
			fallback()
		}
		return nil, err
	}
	return bytes, nil
}

func findUserByInput(input string, users []domain.User) (user domain.User, address domain.PublicAddress, found bool, err error) {
	for i := range users {
		if users[i].Name == input {
			return users[i], users[i].PublicAddress, true, nil
		}
	}

	hexBytes, err := hex.DecodeString(input)
	if err != nil {
		return domain.User{}, domain.PublicAddress{}, false, fmt.Errorf("input is neither a valid user name nor a valid hex string")
	}

	if len(hexBytes) == 20 {
		var addr domain.PublicAddress
		copy(addr[:], hexBytes)
		for i := range users {
			if users[i].PublicAddress == addr {
				return users[i], addr, true, nil
			}
		}
		return domain.User{}, addr, false, nil
	} else if len(hexBytes) == 33 {
		var pubKey domain.PublicKey
		copy(pubKey[:], hexBytes)
		for i := range users {
			if users[i].PublicKey == pubKey {
				return users[i], users[i].PublicAddress, true, nil
			}
		}
		return domain.User{}, domain.PublicAddress{}, false, fmt.Errorf("no user found with that public key")
	}

	return domain.User{}, domain.PublicAddress{}, false, fmt.Errorf("hex input must be either 40 characters (public address) or 66 characters (public key)")
}

func printMenu() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    BLOCKCHAIN CLI - AVAILABLE COMMANDS                ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ MINING:                                                               ║")
	fmt.Println("║   mineblocks          - Mine new blocks with random transactions      ║")
	fmt.Println("║   simulatedecentralizedmining - Simulate decentralized mining         ║")
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
	fmt.Println("║   getuserbalance      - Get balance by name, public key, or address   ║")
	fmt.Println("║   richlist            - Show top users by balance                     ║")
	fmt.Println("║   getutxos            - Get UTXOs by name, public key, or address     ║")
	fmt.Println("║                                                                       ║")
	fmt.Println("║ OTHER:                                                                ║")
	fmt.Println("║   help                - Show detailed help                            ║")
	fmt.Println("║   exit                - Exit the program                              ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
}

func calculateStats(bch *blockchain.Blockchain, users []domain.User, cfg *config.Config) (totalBlocks, totalTxs, totalUsers int, avgTxPerBlock float64, version, difficulty uint32) {
	blocks := bch.Blocks()
	totalBlocks = len(blocks)
	totalUsers = len(users)
	version = cfg.Version
	difficulty = cfg.Difficulty

	for _, block := range blocks {
		totalTxs += len(block.Body.Transactions)
	}

	if totalBlocks > 0 {
		avgTxPerBlock = float64(totalTxs) / float64(totalBlocks)
	}

	return
}

func validateChain(bch *blockchain.Blockchain) bool {
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

	return valid
}

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
					bch.RegisterUsers(users)
					genesis, _ := bch.GetLatestBlock()
					genesisHeader := genesis.Header
					log.Println("Found a POW hash successfully with nonce:", genesisHeader.Nonce)
					log.Println("Added genesis block successfully")

					txsSize := 100
					config := blockchain.DefaultDecentralizedMiningConfig()
					config.BlockCount = 5
					config.TxCount = txsSize
					config.Low = 10
					config.High = 50
					config.Version = cfg.Version
					config.Difficulty = cfg.Difficulty
					err := bch.MineBlocksDecentralized(ctx, users, config)
					if err != nil {
						log.Println("Error mining initial blocks:", err)
					}
					for {
						printMenu()
						command, err := readCommand()
						if err != nil {
							fmt.Println(err)
							continue
						}
						switch command {
						case "getblockheader":
							block, index, err := getBlockByIndex(bch, "Please enter block index:")
							if err != nil {
								fmt.Printf("Invalid block index: %v\n", err)
								continue
							}
							headBytes, err := marshalJSON(block.Header, func() {
								fmt.Printf("Block Header at index %d: %+v\n", index, block.Header)
							})
							if err != nil {
								continue
							}
							fmt.Printf("Block Header at index %d:\n%s\n", index, string(headBytes))
						case "height":
							height := bch.Len()
							fmt.Printf("Current blockchain height: %d\n", height)
						case "stats":
							totalBlocks, totalTxs, totalUsers, avgTxPerBlock, version, difficulty := calculateStats(bch, users, cfg)
							fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
							fmt.Println("║                   BLOCKCHAIN STATISTICS                       ║")
							fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
							fmt.Printf("║ Total Blocks:              %34d ║\n", totalBlocks)
							fmt.Printf("║ Total Transactions:        %34d ║\n", totalTxs)
							fmt.Printf("║ Avg Transactions/Block:    %34.2f ║\n", avgTxPerBlock)
							fmt.Printf("║ Total Users:               %34d ║\n", totalUsers)
							fmt.Printf("║ Current Version:           %34d ║\n", version)
							fmt.Printf("║ Current Difficulty:        %34d ║\n", difficulty)
							fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
						case "validatechain":
							validateChain(bch)
						case "getblock":
							block, index, err := getBlockByIndex(bch, "Please enter block index:")
							if err != nil {
								fmt.Printf("Invalid block index: %v\n", err)
								continue
							}
							blockBytes, err := marshalJSON(block, func() {
								fmt.Printf("Block at index %d: %+v\n", index, block)
							})
							if err != nil {
								continue
							}
							fmt.Printf("Block at index %d:\n%s\n", index, string(blockBytes))
						case "getblockhash":
							block, index, err := getBlockByIndex(bch, "Please enter block index:")
							if err != nil {
								fmt.Printf("Invalid block index: %v\n", err)
								continue
							}
							hash := bch.CalculateHash(block)
							fmt.Printf("Block Hash at index %d: %x\n", index, hash)
						case "mineblocks":
							numBlocks, err := readInt("Please enter number of blocks to mine concurrently:", func(v int) error {
								return validatePositiveInt(v, "number of blocks")
							})
							if err != nil {
								fmt.Printf("Invalid input: %v\n", err)
								continue
							}
							numTxs, err := readInt("Please enter number of transactions per block:", func(v int) error {
								return validatePositiveInt(v, "number of transactions")
							})
							if err != nil {
								fmt.Printf("Invalid input: %v\n", err)
								continue
							}
							minTxValue, err := readInt("Please enter minimum transaction value:", nil)
							if err != nil {
								fmt.Printf("Invalid input: %v\n", err)
								continue
							}
							maxTxValue, err := readInt("Please enter maximum transaction value:", nil)
							if err != nil {
								fmt.Printf("Invalid input: %v\n", err)
								continue
							}
							if err := validateTransactionValueRange(minTxValue, maxTxValue); err != nil {
								fmt.Printf("Invalid transaction value range: %v\n", err)
								continue
							}

							ctx := context.Background()
							err = bch.MineBlocks(ctx, numBlocks, numTxs, minTxValue, maxTxValue, users, cfg.Version, cfg.Difficulty)
							if err != nil {
								fmt.Println("Error mining blocks:", err)
							}
						case "simulatedecentralizedmining":
							config := blockchain.DefaultDecentralizedMiningConfig()
							config.Version = cfg.Version
							config.Difficulty = cfg.Difficulty

							numBlocks, _ := readIntWithDefault("Please enter number of blocks to mine:", 1, func(v int) error {
								return validatePositiveInt(v, "number of blocks")
							})
							config.BlockCount = numBlocks

							numTxs, _ := readIntWithDefault("Please enter number of transactions per candidate block (default: 100):", 100, func(v int) error {
								return validatePositiveInt(v, "number of transactions")
							})
							config.TxCount = numTxs

							candidateCount, _ := readIntWithDefault("Please enter number of candidate blocks to generate (default: 5):", 5, func(v int) error {
								return validatePositiveInt(v, "number of candidates")
							})
							config.CandidateCount = candidateCount

							timeLimitSeconds, _ := readIntWithDefault("Please enter initial time limit in seconds (default: 5):", 5, func(v int) error {
								return validatePositiveInt(v, "time limit")
							})
							config.InitialTimeLimit = time.Duration(timeLimitSeconds) * time.Second

							minTxValue, _ := readIntWithDefault("Please enter minimum transaction value (default: 1):", 1, nil)
							maxTxValue, _ := readIntWithDefault("Please enter maximum transaction value (default: 1000):", 1000, nil)
							if err := validateTransactionValueRange(minTxValue, maxTxValue); err != nil {
								fmt.Printf("Invalid transaction value range: %v, using defaults\n", err)
								minTxValue = 1
								maxTxValue = 1000
							}
							config.Low = minTxValue
							config.High = maxTxValue

							fmt.Println("\nStarting decentralized mining simulation...")
							fmt.Printf("Configuration: %d blocks, %d candidates per round, %d tx per candidate, %v initial time limit\n",
								config.BlockCount, config.CandidateCount, config.TxCount, config.InitialTimeLimit)

							ctx := context.Background()
							err := bch.MineBlocksDecentralized(ctx, users, config)
							if err != nil {
								fmt.Println("Error in decentralized mining:", err)
							} else {
								fmt.Println("Decentralized mining completed successfully!")
							}
						case "getblocktransactions":
							block, index, err := getBlockByIndex(bch, "Please enter block index:")
							if err != nil {
								fmt.Printf("Invalid block index: %v\n", err)
								continue
							}
							bodyBytes, err := marshalJSON(block.Body.Transactions, func() {
								fmt.Printf("Block Transactions at index %d: %+v\n", index, block.Body.Transactions)
							})
							if err != nil {
								continue
							}
							fmt.Printf("Block Transactions at index %d:\n%s\n", index, string(bodyBytes))
						case "getuserbalance":
							input, err := readString("Please enter user name, public key (hex), or public address (hex):")
							if err != nil {
								fmt.Println(err)
								continue
							}

							user, address, found, err := findUserByInput(input, users)
							if err != nil {
								fmt.Println("Error:", err)
								if err.Error() == "input is neither a valid user name nor a valid hex string" {
									fmt.Println("Hint: Try using a user name, public key (66 hex chars), or public address (40 hex chars) from the 'balance' command")
								} else if err.Error() == "no user found with that public key" {
									fmt.Println("Hint: Use the 'balance' command to see all users and their public keys")
								} else {
									fmt.Println("Hint: Public address = 40 hex chars, Public key = 66 hex chars")
								}
								continue
							}

							balance := bch.GetUserBalance(address)
							addressHex := fmt.Sprintf("%x", address)

							fmt.Println("\n╔═══════════════════════════════════════════════════════════════════════════════════════════╗")
							if found {
								fmt.Printf("║ User:       %-77s ║\n", user.Name)
							} else {
								fmt.Printf("║ User:       %-77s ║\n", "Unknown")
							}
							fmt.Printf("║ Balance:    %-77d ║\n", balance)
							fmt.Printf("║ Address:    %-77s ║\n", addressHex)
							if found {
								pubKeyHex := fmt.Sprintf("%x", user.PublicKey)
								fmt.Printf("║ Public Key: %-77s ║\n", pubKeyHex)
							}
							fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════════════════╝")
						case "getallheaders":
							headers := make([]domain.Header, 0, bch.Len())
							for _, blk := range bch.Blocks() {
								headers = append(headers, blk.Header)
							}
							headersBytes, err := marshalJSON(headers, nil)
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

							topN, _ := readIntWithDefault("How many top users to show?", 10, func(v int) error {
								return validatePositiveInt(v, "number of top users")
							})
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
							input, err := readString("Please enter user name, public key (hex), or public address (hex):")
							if err != nil {
								fmt.Println(err)
								continue
							}

							user, address, found, err := findUserByInput(input, users)
							if err != nil {
								fmt.Println("Error:", err)
								if err.Error() == "input is neither a valid user name nor a valid hex string" {
									fmt.Println("Hint: Try using a user name, public key (66 hex chars), or public address (40 hex chars) from the 'balance' command")
								} else if err.Error() == "no user found with that public key" {
									fmt.Println("Hint: Use the 'balance' command to see all users and their public keys")
								} else {
									fmt.Println("Hint: Public address = 40 hex chars, Public key = 66 hex chars")
								}
								continue
							}

							utxos := bch.GetUTXOsForAddress(address)
							addressHex := fmt.Sprintf("%x", address)
							displayName := addressHex
							if found {
								displayName = user.Name
							}

							fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════════════╗\n")
							fmt.Printf("║                            UTXOs for %-51s ║\n", displayName)
							fmt.Println("╠══════╦═══════════════╦═══════════════════════════════════════════════════════════════════╣")
							fmt.Println("║  #   ║     VALUE     ║                    TRANSACTION ID:INDEX                           ║")
							fmt.Println("╠══════╬═══════════════╬═══════════════════════════════════════════════════════════════════╣")

							totalValue := uint32(0)
							if len(utxos) == 0 {
								fmt.Println("║                         No UTXOs found for this address                                    ║")
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
							fmt.Println("║   simulatedecentralizedmining - Simulates decentralized mining with multiple candidates   ║")
							fmt.Println("║                Generates multiple candidate blocks and mines them with time limits        ║")
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
							fmt.Println("║   getuserbalance - Get balance (by name, public key, or public address)                   ║")
							fmt.Println("║   richlist   - Show top N users ranked by balance                                         ║")
							fmt.Println("║   getutxos   - Show all UTXOs (by name, public key, or public address)                    ║")
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
