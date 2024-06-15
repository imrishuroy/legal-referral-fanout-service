package main

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog/log"
	db "legal-referral-fanout-service/db/sqlc"
	"legal-referral-fanout-service/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CreateConsumer(ctx context.Context, store db.Store, config util.Config) error {

	// sets the consumer group ID and offset
	//conf["group.id"] = "go-group-1"
	//conf["auto.offset.reset"] = "earliest"

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		// User-specific properties that you must set
		"bootstrap.servers": config.BootStrapServers,
		"sasl.username":     config.SASLUsername,
		"sasl.password":     config.SASLPassword,

		// Fixed properties
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"group.id":          "confluentinc-kafka-go",
		"auto.offset.reset": "latest"})

	if err != nil {
		log.Error().Err(err).Msg("cannot create consumer")
		os.Exit(1)
	}

	// creates a new consumer and subscribes to your topic
	//consumer, _ := kafka.NewConsumer(&conf)
	//consumer.SubscribeTopics([]string{topic}, nil)

	topic := "publish-feed"
	err = consumer.SubscribeTopics([]string{topic}, nil)
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Info().Msgf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				//log.Error().Err(err).Msg("cannot read message")
				// Errors are informational and automatically handled by the consumer
				continue
			}
			//fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
			//	*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))

			log.Info().Msgf("Consumed event from topic %s: key = %-10s value = %s\n",
				*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))

			err = postToNewsFeed(ctx, store, string(ev.Key), int32(ev.Value[0]))
			if err != nil {
				log.Error().Err(err).Msg("cannot post to news feed")
				return err
			}

			//arg := db.PostToNewsFeedParams{
			//	UserID: string(ev.Key),
			//	PostID: int32(ev.Value[0]),
			//}
			//
			//err = store.PostToNewsFeed(ctx, arg)
			//if err != nil {
			//	return
			//}

		}
	}

	// Process messages
	//run := true
	//for run {
	//	// consumes messages from the subscribed topic and prints them to the console
	//	e := consumer.Poll(1000)
	//	switch ev := e.(type) {
	//	case *kafka.Message:
	//		log.Info().Msgf("Consumed event from topic %s: key = %-10s value = %s\n",
	//			*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
	//		// application-specific processing
	//
	//		arg := db.PostToNewsFeedParams{
	//			UserID: "DlQrTA39q7aI8twLglKknFmWDMF2",
	//			PostID: 12,
	//		}
	//
	//		err := server.store.PostToNewsFeed(ctx, arg)
	//		if err != nil {
	//			return
	//		}
	//
	//		fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
	//			*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
	//	case kafka.Error:
	//		_, err2 := fmt.Fprintf(os.Stderr, "%% Error: %v\n", ev)
	//		if err2 != nil {
	//			return
	//		}
	//		run = false
	//	}
	//}

	// closes the consumer connection
	err = consumer.Close()
	if err != nil {
		log.Error().Err(err).Msg("cannot close consumer")
		return err
	}

	return nil

}

func postToNewsFeed(ctx context.Context, store db.Store, userID string, postID int32) error {
	// Initialize the userIDs slice with the given userID as the first element
	userIDs := []string{userID}

	// Get the list of connected user IDs
	connectedUserIDs, err := store.ListConnectedUserIDs2(ctx, userID)
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
