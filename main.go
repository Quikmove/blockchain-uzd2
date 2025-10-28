package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/Quikmove/blockchain-uzd2/internal/api"
	"github.com/Quikmove/blockchain-uzd2/internal/blockchain"
	"github.com/Quikmove/blockchain-uzd2/internal/config"
	"github.com/Quikmove/blockchain-uzd2/internal/filetolist"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.LoadConfig()
	log.Println(cfg.NameListPath)
	nameList := filetolist.FileToList(cfg.NameListPath)
	users := blockchain.GenerateUsers(nameList, 3)
	var bch *blockchain.Blockchain
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	bch = blockchain.InitBlockchainWithFunds(100, 1000000, users, cfg)
	if err != nil {
		log.Fatalf("Failed to create genesis block: %v", err)
	}
	errChan := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		err := api.Run(ctx, bch, cfg, started)
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
