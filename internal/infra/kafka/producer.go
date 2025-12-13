package kafka

import (
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Producer struct {
	Sync  sarama.SyncProducer
	Async sarama.AsyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = time.Second
	config.Version = sarama.V2_5_0_0

	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	sp, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	ap, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	go func() {
		for err := range ap.Errors() {
			log.Println("Kafka Async Error:", err)
		}
	}()

	return &Producer{Sync: sp, Async: ap}, nil
}

func (p *Producer) SendSync(topic string, data []byte) error {
	_, _, err := p.Sync.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (p *Producer) SendAsync(topic string, data []byte) {
	p.Async.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}
}
