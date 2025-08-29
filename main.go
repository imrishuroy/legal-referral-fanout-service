package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
	"github.com/imrishuroy/legal-referral-fanout-service/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var (
	config    util.Config
	store     db.Store
	sqsClient *sqs.Client
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

type PostFanOutMsg struct {
	OwnerID string `json:"owner_id"`
	PostID  int32  `json:"post_id"`
}

// pollSQS continuously polls SQS and processes messages
func pollSQS(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("SQS poller shutting down")
			return
		default:
		}

		if sqsClient == nil {
			log.Error().Msg("SQS client is not initialized")
			return
		}

		result, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &config.SQSURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
		})
		if err != nil {
			log.Error().Err(err).Msg("Error receiving message")
			continue
		}

		for _, msg := range result.Messages {
			if msg.Body == nil {
				continue
			}
			log.Info().Msgf("Received message: %s", *msg.Body)

			var (
				ownerID string
				postID  int32
				ok      bool
			)

			// Try string post_id
			var raw1 struct {
				OwnerID string `json:"owner_id"`
				PostID  string `json:"post_id"`
			}
			if err := json.Unmarshal([]byte(*msg.Body), &raw1); err == nil && raw1.OwnerID != "" && raw1.PostID != "" {
				if n, err := strconv.Atoi(raw1.PostID); err == nil {
					ownerID, postID, ok = raw1.OwnerID, int32(n), true
				}
			}

			// Try numeric post_id
			if !ok {
				var raw2 struct {
					OwnerID string `json:"owner_id"`
					PostID  int32  `json:"post_id"`
				}
				if err := json.Unmarshal([]byte(*msg.Body), &raw2); err == nil && raw2.OwnerID != "" && raw2.PostID != 0 {
					ownerID, postID, ok = raw2.OwnerID, raw2.PostID, true
				}
			}

			if !ok {
				log.Error().Msg("Invalid message body shape")
				continue
			}

			if err := publishNewsFeed(ctx, store, ownerID, postID); err != nil {
				log.Error().Err(err).Str("owner_id", ownerID).Int32("post_id", postID).Msg("Failed to publish news feed")
				continue
			}

			// Delete after processing
			_, delErr := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      &config.SQSURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if delErr != nil {
				log.Error().Err(delErr).Msg("Failed to delete message")
			}
		}
	}
}

// health handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func main() {
	log.Info().Msg("Starting LegalReferral Fan-out Service (App Runner)")

	// db connection
	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer connPool.Close()
	store = db.NewStore(connPool)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load AWS SDK config")
	}
	sqsClient = sqs.NewFromConfig(cfg)

	// Context canceled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start SQS poller
	go pollSQS(ctx)

	// Health endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	addr := config.ServerAddress
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{Addr: addr, Handler: mux}

	// Start server
	go func() {
		log.Info().Msgf("HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10_000_000_000) // 10s
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info().Msg("Service shut down")
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
