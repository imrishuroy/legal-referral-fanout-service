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

func init() {
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

}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	err := api.ConnectConsumer(config, store)
	if err != nil {
		log.Error().Err(err).Msg("cannot connect consumer")
		return err
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}

//
//func healthCheck(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprint(w, "OK")
//}
//
//func main() {
//
//	log.Info().Msg("Welcome to LegalReferral Fan-out Service")
//	config, err := util.LoadConfig(".")
//	if err != nil {
//		log.Error().Err(err).Msg("cannot load config")
//	}
//
//	// db connection
//	connPool, err := pgxpool.New(context.Background(), config.DBSource)
//
//	if err != nil {
//		fmt.Println("cannot connect to db:", err)
//	}
//	defer connPool.Close() // close db connection
//
//	store := db.NewStore(connPool)
//
//	go func() {
//		err := api.ConnectConsumer(config, store)
//		if err != nil {
//			log.Error().Err(err).Msg("cannot connect consumer")
//			panic(err)
//		}
//	}()
//
//	srv := &http.Server{
//		Addr:    config.ServerAddress,
//		Handler: nil,
//	}
//
//	go func() {
//		http.HandleFunc("/health", healthCheck)
//		http.HandleFunc("/", healthCheck)
//		fmt.Println("Starting server at " + config.ServerAddress)
//		log.Info().Msg("server address: " + config.ServerAddress)
//		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
//			log.Error().Err(err).Msg("cannot start server")
//		}
//	}()
//
//	// Graceful shutdown
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//	log.Info().Msg("Shutting down server...")
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := srv.Shutdown(ctx); err != nil {
//		log.Error().Err(err).Msg("Server forced to shutdown")
//	}
//
//	log.Info().Msg("Server exiting")
//
//}
