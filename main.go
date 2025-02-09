package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
	"github.com/imrishuroy/legal-referral-fanout-service/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"time"
)

var (
	config    util.Config
	store     db.Store
	sqsClient *sqs.SQS
)

//func handleRequest(ctx context.Context, event json.RawMessage) error {
//	log.Info().Msg("Lambda function invoked")
//
//	// Connect to consumer
//	err := api.ConnectConsumer(config, store)
//	if err != nil {
//		log.Error().Err(err).Msg("cannot connect consumer")
//		return err
//	}
//	return nil
//
//}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	log.Info().Msg("Lambda function invoked")

	// print SQS URL
	fmt.Println("SQS URL: ", config.SQSURL)

	// Continuous polling loop
	for {
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(config.SQSURL),
			MaxNumberOfMessages: aws.Int64(10), // Fetch multiple messages
			WaitTimeSeconds:     aws.Int64(20), // Long polling to reduce API calls
		})

		if err != nil {
			log.Printf("Error receiving message: %v", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		for _, msg := range result.Messages {
			fmt.Println("Received message:", *msg.Body)

			// Process the message (store in DB, etc.)

			// Delete the message after processing
			_, delErr := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(config.SQSURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if delErr != nil {
				log.Printf("Failed to delete message: %v", delErr)
			}
		}
	}

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
		log.Error().Err(err).Msg("cannot connect to db")
	}
	defer connPool.Close()
	store = db.NewStore(connPool)

	log.Printf("store object initialized: %v", store)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//svc := sqs.New(sess)
	sqsClient = sqs.New(sess)

	lambda.Start(handleRequest)
}

func main12() {

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

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: aws.String(config.SQSURL),
	})
	if err != nil {
		// Handle error
	}

	for _, msg := range result.Messages {
		// Process the received message
		fmt.Println(*msg.Body)
	}

}

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

//lambda.Start(handleRequest)
