package kafka

import (
	"Wb_Test_L0/internal/models"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
)

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(address []string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","),
	}
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("new producer error: %w", err)
	}

	return &Producer{
		producer: p,
	}, nil
}

func (p *Producer) Produce(order models.Order, topic string) error {
	b, _ := json.Marshal(order)
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          b,
	}

	if err := p.producer.Produce(kafkaMsg, nil); err != nil {
		return err
	}

	log.Println("message queued, will flush on Close()")
	return nil
}

func (p *Producer) Close() {
	p.producer.Flush(10000)
	p.producer.Close()
}
