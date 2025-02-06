package api

//
//import (
//	"fmt"
//	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
//	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
//	"github.com/imrishuroy/legal-referral-fanout-service/util"
//	"github.com/rs/zerolog/log"
//	"time"
//)
//
//func CreateConsumer(store db.Store, config util.Config) error {
//
//	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
//		"bootstrap.servers": config.BootStrapServers,
//		"sasl.username":     config.SASLUsername,
//		"sasl.password":     config.SASLPassword,
//
//		// Fixed properties
//		"security.protocol": "SASL_SSL",
//		"sasl.mechanisms":   "PLAIN",
//		"group.id":          "confluentinc-kafka-go",
//		"auto.offset.reset": "latest"})
//
//	if err != nil {
//		log.Error().Err(err).Msg("cannot create consumer")
//		//os.Exit(1)
//		panic(err)
//	}
//
//	topic := "publish-feed"
//	err = consumer.SubscribeTopics([]string{topic}, nil)
//
//	if err != nil {
//		panic(err)
//	}
//
//	// A signal handler or similar could be used to set this to false to break the loop.
//	run := true
//
//	for run {
//		msg, err := consumer.ReadMessage(time.Second)
//		if err != nil {
//			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
//
//			err = publishNewsFeed(store, string(msg.Key), int32(msg.Value[0]))
//			if err != nil {
//				log.Error().Err(err).Msg("cannot post to news feed")
//				return err
//			}
//
//		} else if !err.(kafka.Error).IsTimeout() {
//			// The client will automatically try to recover from all errors.
//			// Timeout is not considered an error because it is raised by
//			// ReadMessage in absence of messages.
//			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
//		}
//	}
//
//	err = consumer.Close()
//	if err != nil {
//		log.Error().Err(err).Msg("cannot close consumer")
//		return err
//	}
//
//	return nil
//
//}
