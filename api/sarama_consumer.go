package api

import (
	"crypto/tls"
	"fmt"
	"github.com/IBM/sarama"
	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
	"github.com/imrishuroy/legal-referral-fanout-service/util"
	"os"
	"os/signal"
	"syscall"
)

func ConnectConsumer(config util.Config, store db.Store) error {

	topic := "publish-feed"
	worker, err := createConsumer(config)
	if err != nil {
		return err
	}
	// Calling ConsumePartition. It will open one connection per broker
	// and share it for all partitions that live on it.
	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetNewest) // sarama.OffsetOldest
	if err != nil {
		return err
	}
	fmt.Println("Consumer started ")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	// Count how many message processed
	msgCount := 0

	// Get signal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("Received message Count %d: | Topic(%s) | Message(%s) \n", msgCount, string(msg.Topic), string(msg.Key))
				err = publishNewsFeed(store, string(msg.Key), int32(msg.Value[0]))
				if err != nil {
					fmt.Println("cannot post to news feed")
				}
			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")

	if err := worker.Close(); err != nil {
		return err
	}

	return nil
}

func createConsumer(config util.Config) (sarama.Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true

	cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext //"PLAIN"
	cfg.Net.SASL.Enable = true

	cfg.Net.SASL.User = config.SASLUsername
	cfg.Net.SASL.Password = config.SASLPassword

	tlsConfig := tls.Config{}
	cfg.Net.TLS.Enable = true
	cfg.Net.TLS.Config = &tlsConfig

	// Create new consumer
	conn, err := sarama.NewConsumer([]string{config.BootStrapServers}, cfg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
