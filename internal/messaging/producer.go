package messaging

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/Andronzi/credit-origination/internal/client"
	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
)

type StatusEvent struct {
	ApplicationID string                   `json:"application_id"`
	NewStatus     domain.ApplicationStatus `json:"new_status"`
	Timestamp     time.Time                `json:"timestamp"`
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	codec    *goavro.Codec
	registry *client.SchemaRegistryClient
	topic    string
	schemaID int
}

func NewKafkaProducer(brokers []string, topic string, schema string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}

	registry := client.NewSchemaRegistryClient("http://schema-registry:8081")
	schemaID, err := registry.GetSchemaID(topic, schema)
	if err != nil {
		log.Printf("Ошбика получения SchemaID: %v", err)
		return nil, err
	}

	return &KafkaProducer{
		producer: producer,
		codec:    codec,
		registry: registry,
		topic:    topic,
		schemaID: schemaID,
	}, nil
}

func (p *KafkaProducer) SendStatusEvent(event StatusEvent) error {
	header := createConfluentHeader(p.schemaID)
	avroData, err := p.codec.BinaryFromNative(nil, map[string]interface{}{
		"application_id":     event.ApplicationID,
		"application_status": "NEW",
		"timestamp":          event.Timestamp.UnixMilli(),
	})
	if err != nil {
		log.Printf("Ошибка сериализации Avro: %v", err)
		return err
	}

	payload := append(header, avroData...)
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.ApplicationID),
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := p.producer.SendMessage(msg)

	if err != nil {
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
		return err
	}

	log.Printf("Отправлено сообщение: partition=%d, offset=%d", partition, offset)

	return nil
}

func createConfluentHeader(schemaID int) []byte {
	header := make([]byte, 5)
	header[0] = 0x0 // Magic byte
	binary.BigEndian.PutUint32(header[1:5], uint32(schemaID))
	return header
}
