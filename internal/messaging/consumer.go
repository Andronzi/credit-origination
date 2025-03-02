package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
)

type KafkaAvroConsumer struct {
	consumer sarama.ConsumerGroup
	codec    *goavro.Codec
	topic    string
}

func NewKafkaAvroConsumer(brokers []string, groupID string, topic string, schema string) (*KafkaAvroConsumer, error) {
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
	}, nil
}

func (c *KafkaAvroConsumer) Consume(ctx context.Context) error {
	handler := consumerHandler{codec: c.codec}

	return c.consumer.Consume(ctx, []string{c.topic}, &handler)
}

type consumerHandler struct {
	codec *goavro.Codec
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
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

		fmt.Printf("Received event: %+v\n", data)
		session.MarkMessage(msg, "")
	}
	return nil
}
