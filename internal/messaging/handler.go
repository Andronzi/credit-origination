package messaging

import "github.com/IBM/sarama"

type MessageHandler interface {
	Handle(message *sarama.ConsumerMessage) error
}
