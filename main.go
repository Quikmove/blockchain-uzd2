package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Quikmove/blockchain-uzd2/internal/api"
	"github.com/Quikmove/blockchain-uzd2/internal/blockchain"
	"github.com/Quikmove/blockchain-uzd2/internal/config"
	"github.com/joho/godotenv"
)

// func (t Transaction) Hash() Hash32 {
// 	var buf bytes.Buffer
// 	buf.Write(t.TxID[:])
// 	buf.Write(t.Sender[:])
// 	buf.Write(t.Receiver[:])
// 	_ = binary.Write(&buf, binary.LittleEndian, t.Amount)

// 	return sha256.Sum256(buf.Bytes())
// }

// func (u Users) GenerateFundTransactions(low, high uint32) (Transactions, error) {
// 	txs := Transactions{}

// 	if len(u) == 0 || high < low {
// 		return txs, nil
// 	}

// 	for _, usr := range u {

// 		b := make([]byte, 16)
// 		_, _ = rand.Read(b)
// 		seed := hex.EncodeToString(b)

// 		sub := seed
// 		if len(sub) > 16 {
// 			sub = sub[:16]
// 		}

// 		amount := low
// 		if v, err := strconv.ParseUint(sub, 16, 32); err == nil {
// 			amount = uint32(v%uint64(high-low+1) + uint64(low))
// 		}

// 		// coinbase transaction creates an unspent output (UTXO) for the user
// 		tx := Transaction{
// 			TxID:     HashString("coinbase" + usr.PublicKey.HexString() + strconv.FormatUint(uint64(amount), 10) + strconv.FormatInt(time.Now().Unix(), 10)),
// 			Sender:   HashString("coinbase"),
// 			Receiver: usr.PublicKey,
// 			Amount:   amount,
// 		}

// 		txs = append(txs, tx)

// 		// small pause to vary timestamp-derived seeds
// 		time.Sleep(time.Second)
// 	}

//		return txs, nil
//	}

func fileToList(path string) []string {
	data, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer data.Close()

	// read by line and append
	var list []string
	for {
		var line string
		_, err := fmt.Fscanln(data, &line)
		if err != nil {
			break
		}
		list = append(list, line)
	}
	return list
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	nameList := fileToList("assets/name_list.txt")
	users := blockchain.GenerateUsers(nameList, 3)
	fundTransactions, err := blockchain.GenerateFundTransactionsForUsers(users, 10, 30)
	if err != nil {
		log.Fatalf("Failed to generate fund transactions: %v", err)
	}

	bch := blockchain.NewBlockchain()
	config := config.LoadConfig()
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	func() {
		genesisBlock, err := blockchain.CreateGenesisBlock(ctx, fundTransactions, config)
		if err != nil {
			log.Fatalf("Failed to create genesis block: %v", err)
		}
		bch.Blocks = append(bch.Blocks, genesisBlock)
	}()
	errChan := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		err := api.Run(ctx, bch, config, started)
		errChan <- err
	}()

	<-started

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println("Press A to add a new random block...")
				var input string
				_, err := fmt.Scanln(&input)
				if err != nil && err.Error() != "unexpected newline" {
					fmt.Println("Error reading input:", err)
					continue
				}
				if input == "a" || input == "A" {
					fmt.Println("Hooray! New block added!")
				}
			}
		}
	}()
	select {
	case <-ctx.Done():
		log.Println("Shutting down main...")
	case err := <-errChan:
		if err != nil {
			log.Fatalln(err)
		}
	}
}
