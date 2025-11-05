package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
					log.Println("Version:", cfg.Version)
					log.Println("Difficulty:", cfg.Difficulty)
					names := filetolist.FileToList(cfg.NameListPath)
					users := blockchain.GenerateUsers(names, 3)
					log.Println("Generating genesis block...")
					bch := blockchain.InitBlockchainWithFunds(100, 1000000, users, cfg, blockchain.NewArchasHasher())
					genesis, _ := bch.GetLatestBlock()
					log.Println("Found a POW hash successfully with nonce:", genesis.Header.Nonce)

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
