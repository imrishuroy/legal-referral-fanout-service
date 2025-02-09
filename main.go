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
	"os"
)

var (
	config    util.Config
	store     db.Store
	sqsClient *sqs.SQS
)

func init() {
	log.Info().Msg("Initializing LegalReferral Fan-out Service")

	var err error
	config, err = util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
		os.Exit(1)
	}
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	log.Info().Msg("Lambda function invoked")

	// Ensure sqsClient is initialized
	if sqsClient == nil {
		log.Error().Msg("SQS client is not initialized")
		return fmt.Errorf("SQS client is not initialized")
	}

	log.Info().Msgf("SQS URL: %s", config.SQSURL)

	// Receive messages from SQS
	result, err := sqsClient.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(config.SQSURL),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(20),
	})
	if err != nil {
		log.Error().Err(err).Msg("Error receiving message")
		return err
	}

	for _, msg := range result.Messages {
		log.Info().Msgf("Received message: %s", *msg.Body)

		// Process the message (e.g., store it in DB)

		// Delete the message after processing
		_, delErr := sqsClient.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(config.SQSURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if delErr != nil {
			log.Error().Err(delErr).Msg("Failed to delete message")
		}
	}

	return nil
}

func main() {
	log.Info().Msg("Welcome to LegalReferral Fan-out Service")

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
