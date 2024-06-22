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

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Success")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Success")
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
	http.HandleFunc("/hello", helloHandler)
	fmt.Println("Starting server at port 8000...")
	log.Info().Msg("server address: " + config.ServerAddress)
	if err := http.ListenAndServe(config.ServerAddress, nil); err != nil {
		fmt.Println(err)
	}

}
