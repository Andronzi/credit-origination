package messaging

import (
	"context"
	"log"

	"github.com/Andronzi/credit-origination/pkg/logger"
	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"
)

type KafkaAvroConsumer struct {
	consumer sarama.ConsumerGroup
	codec    *goavro.Codec
	topic    string
	handlers []MessageHandler
}

func NewKafkaAvroConsumer(
	brokers []string,
	groupID string,
	topic string,
	schema string,
	handlers []MessageHandler,
) (*KafkaAvroConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}

	return &KafkaAvroConsumer{
		consumer: consumer,
		codec:    codec,
		topic:    topic,
		handlers: handlers,
	}, nil
}

func (c *KafkaAvroConsumer) Consume(ctx context.Context) error {
	handler := consumerHandler{codec: c.codec, handlers: c.handlers}

	return c.consumer.Consume(ctx, []string{c.topic}, &handler)
}

type consumerHandler struct {
	codec    *goavro.Codec
	handlers []MessageHandler
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Logger.Info("Received Mssage", zap.ByteString("key", msg.Key), zap.Int64("offset", msg.Offset), zap.Int("length", len(msg.Value)))
		if len(msg.Value) < 5 {
			log.Printf("Invalid message length")
			continue
		}
		avroData := msg.Value[5:]

		native, _, err := h.codec.NativeFromBinary(avroData)
		if err != nil {
			log.Printf("Failed to decode Avro: %v", err)
			continue
		}

		data, ok := native.(map[string]interface{})
		if !ok {
			log.Printf("Unexpected message format")
			continue
		}

		logger.Logger.Info("Before handler", zap.Int("length", len(h.handlers)))

		for _, handlers := range h.handlers {
			logger.Logger.Info("Inside handlers")

			if err := handlers.Handle(msg); err != nil {
				logger.Logger.Error("Failed to handle message", zap.Error(err))
			}
		}

		logger.Logger.Info("Received event: %+v\n", zap.Any("event", data))
		session.MarkMessage(msg, "")
	}
	return nil
}
