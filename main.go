package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imrishuroy/legal-referral-fanout-service/api"
	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
	"github.com/imrishuroy/legal-referral-fanout-service/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var (
	config util.Config
	store  db.Store
)

func handleRequest(ctx context.Context, event json.RawMessage) error {
	log.Info().Msg("Lambda function invoked")

	// Connect to consumer
	err := api.ConnectConsumer(config, store)
	if err != nil {
		log.Error().Err(err).Msg("cannot connect consumer")
		return err
	}
	return nil

}

func main() {

	log.Info().Msg("Welcome to LegalReferral Fan-out Service")

	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Error().Err(err).Msg("cannot load config")
	}

	config = cfg

	// db connection
	connPool, err := pgxpool.New(context.Background(), config.DBSource)

	if err != nil {
		fmt.Println("cannot connect to db:", err)
	}
	defer connPool.Close() // close db connection

	store = db.NewStore(connPool)
	// print store object

	log.Printf("store object: %v", store)

	// if running locally
	//var wg sync.WaitGroup
	//wg.Add(1) // Add one counter to wait for goroutine
	//
	//go func() {
	//	defer wg.Done() // Decrement counter when goroutine exits
	//	err := api.ConnectConsumer(config, store)
	//	if err != nil {
	//		log.Error().Err(err).Msg("cannot connect consumer")
	//		panic(err)
	//	}
	//}()
	//
	//// Wait for goroutine
	//wg.Wait()

	lambda.Start(handleRequest)
}
