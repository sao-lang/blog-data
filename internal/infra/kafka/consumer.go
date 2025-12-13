package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type HandlerFunc func(msg *sarama.ConsumerMessage) error

type Consumer struct {
	Group sarama.ConsumerGroup
}

func NewConsumer(brokers []string, group string) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_5_0_0
	cfg.Consumer.Return.Errors = true

	g, err := sarama.NewConsumerGroup(brokers, group, cfg)
	if err != nil {
		return nil, err
	}
	return &Consumer{Group: g}, nil
}

func (c *Consumer) Consume(topics []string, fn HandlerFunc) {
	go func() {
		for err := range c.Group.Errors() {
			log.Println("Kafka Consumer Error:", err)
		}
	}()

	handler := groupHandler{fn: fn}

	go func() {
		for {
			_ = c.Group.Consume(context.Background(), topics, handler)
		}
	}()
}

type groupHandler struct{ fn HandlerFunc }

func (h groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h groupHandler) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	for msg := range c.Messages() {
		if err := h.fn(msg); err != nil {
			log.Println("Handle message error:", err)
		}
		s.MarkMessage(msg, "")
	}
	return nil
}
