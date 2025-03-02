package config

type KafkaConfig struct {
	Brokers       []string
	StatusTopic   string
	ReplyTopic    string
	ConsumerGroup string
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:       []string{"localhost:9092"},
		StatusTopic:   "application-status-events",
		ReplyTopic:    "status-change-responses",
		ConsumerGroup: "credit-service-group",
	}
}
