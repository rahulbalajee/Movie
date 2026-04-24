package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
)

type Ingester struct {
	consumer *kafka.Consumer
	topic    string
}

func NewIngester(addr, groupId, topic string) (*Ingester, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		return nil, err
	}

	return &Ingester{consumer: consumer, topic: topic}, nil
}

func (i *Ingester) Ingest(ctx context.Context) (chan model.RatingEvent, error) {
	fmt.Println("starting kafka ingester")

	if err := i.consumer.SubscribeTopics([]string{i.topic}, nil); err != nil {
		return nil, err
	}

	ch := make(chan model.RatingEvent, 1)
	go func() {
		defer close(ch)
		defer i.consumer.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, err := i.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.IsTimeout() {
					continue
				}
				fmt.Println("consumer error: ", err.Error())
				continue
			}

			fmt.Println("processing a message")
			var event model.RatingEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				fmt.Println("unmarshal error: ", err.Error())
				continue
			}

			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
