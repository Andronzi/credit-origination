package messaging

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/Andronzi/credit-origination/internal/client"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/linkedin/goavro/v2"
)

type AgreementDetails struct {
	ApplicationID      string `avro:"application_id"`
	ClientID           string `avro:"client_id"`
	DisbursementAmount int64  `avro:"disbursement_amount"`
	OriginationAmount  int64  `avro:"origination_amount"`
	ToBankAccountID    string `avro:"to_bank_account_id"`
	Term               int32  `avro:"term"`
	Interest           int64  `avro:"interest"`
	ProductCode        string `avro:"product_code"`
	ProductVersion     string `avro:"product_version"`
	PaymentDate        *int64 `avro:"payment_date"` // TODO: Использую пока что указатель для поддержки nil :hmm:
}

type ApplicationStatusEvent struct {
	MessageID        string           `avro:"message_id"`
	EventType        string           `avro:"event_type"`
	ApplicationID    string           `avro:"application_id"`
	Timestamp        int64            `avro:"timestamp"`
	AgreementDetails AgreementDetails `avro:"agreement_details"`
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

func (p *KafkaProducer) SendStatusEvent(event ApplicationStatusEvent) error {
	header := createConfluentHeader(p.schemaID)
	avroData, err := p.codec.BinaryFromNative(nil, map[string]interface{}{
		"message_id":     uuid.New().String(),
		"event_type":     event.EventType,
		"timestamp":      time.Now().UnixMilli(),
		"application_id": event.ApplicationID,
		"agreement_details": map[string]interface{}{
			"application_id":      event.ApplicationID,
			"client_id":           event.AgreementDetails.ClientID,
			"disbursement_amount": event.AgreementDetails.DisbursementAmount,
			"origination_amount":  event.AgreementDetails.OriginationAmount,
			"to_bank_account_id":  event.AgreementDetails.ToBankAccountID,
			"term":                event.AgreementDetails.Term,
			"interest":            event.AgreementDetails.Interest,
			"product_code":        event.AgreementDetails.ProductCode,
			"product_version":     event.AgreementDetails.ProductVersion,
			"payment_date":        nil,
		},
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
