package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
)

const (
	fileName = "ratingsdata.json"
	topic    = "ratings"
	timeout  = 10 * time.Second
)

func main() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	fmt.Println("reading rating events from file " + fileName)

	ratingEvents, err := readRatingEvents(fileName)
	if err != nil {
		panic(err)
	}

	if err := produceRatingEvents(topic, producer, ratingEvents); err != nil {
		panic(err)
	}

	fmt.Println("waiting " + timeout.String() + " until all events get produced")

	producer.Flush(int(timeout.Milliseconds()))
}

func readRatingEvents(filename string) ([]model.RatingEvent, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ratings []model.RatingEvent
	if err := json.NewDecoder(f).Decode(&ratings); err != nil {
		return nil, err
	}

	return ratings, nil
}

func produceRatingEvents(topic string, producer *kafka.Producer, events []model.RatingEvent) error {
	for _, ratingEvent := range events {
		encodedEvent, err := json.Marshal(ratingEvent)
		if err != nil {
			return err
		}

		if err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          encodedEvent,
		}, nil); err != nil {
			return err
		}
	}

	return nil
}
