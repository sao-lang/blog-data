package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

func SendDLQ(p *Producer, mainTopic string, msg *sarama.ConsumerMessage) {
	dlqTopic := fmt.Sprintf("%s.dlq", mainTopic)
	_ = p.SendSync(dlqTopic, msg.Value)
}
