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

		// Unmarshal the message body safely
		var messageBody map[string]interface{}
		if err := json.Unmarshal([]byte(*msg.Body), &messageBody); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal message body")
			continue
		}

		// Extract owner_id safely
		userID, ok := messageBody["owner_id"].(string)
		if !ok || userID == "" {
			log.Error().Msg("Invalid or missing owner_id in message")
			continue
		}

		// Extract post_id safely
		postIDFloat, ok := messageBody["post_id"].(float64)
		if !ok {
			log.Error().Msg("Invalid or missing post_id in message")
			continue
		}
		postID := int32(postIDFloat) // Safe conversion
		log.Info().Msgf("Owner ID: %s, Post ID: %d", userID, postID)

		// Process the message (e.g., store it in DB)
		if err := publishNewsFeed(ctx, store, userID, postID); err != nil {
			log.Error().Err(err).Msg("Failed to publish news feed")
		}

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

func publishNewsFeed(ctx context.Context, store db.Store, userID string, postID int32) error {
	// Initialize the userIDs slice with the given userID as the first element
	userIDs := []string{userID}

	// Get the list of connected user IDs
	connectedUserIDs, err := store.ListConnectedUserIDs(context.Background(), userID)
	if err != nil {
		log.Error().Err(err).Msg("cannot list connected user IDs")
		return err
	}

	// Append the connected user IDs to the userIDs slice
	userIDs = append(userIDs, connectedUserIDs...)

	// Post to the news feed for each user ID
	for _, id := range userIDs {
		arg := db.PostToNewsFeedParams{
			UserID: id,
			PostID: postID,
		}

		if err := store.PostToNewsFeed(ctx, arg); err != nil {
			log.Error().Err(err).Str("userID", id).Msg("cannot post to news feed")
			return err
		}
	}
	return nil
}
