package main

import (
	"context"
	"fmt"
	"github.com/imrishuroy/legal-referral-fanout-service/api"
	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
	"github.com/imrishuroy/legal-referral-fanout-service/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"net/http"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func main() {

	log.Info().Msg("Welcome to LegalReferral Fan-out Service")

	config, err := util.LoadConfig(".")
	if err != nil {

		log.Fatal().Msg("cannot connect to db:")
	}

	log.Info().Msg("DB Source: " + config.DBSource)
	// db connection
	connPool, err := pgxpool.New(context.Background(), config.DBSource)

	if err != nil {
		fmt.Println("cannot connect to db:", err)
	}
	defer connPool.Close() // close db connection

	store := db.NewStore(connPool)

	go func() {
		err := api.CreateConsumer(context.Background(), store, config)
		if err != nil {
			log.Error().Err(err).Msg("cannot create consumer")
		}
	}()

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/", healthCheck)
	fmt.Println("Starting server at " + config.ServerAddress)
	log.Info().Msg("server address: " + config.ServerAddress)
	if err := http.ListenAndServe(config.ServerAddress, nil); err != nil {
		log.Info().Err(err).Msg("cannot start server")
	}

}
