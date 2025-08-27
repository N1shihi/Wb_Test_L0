package kafka

import (
	"Wb_Test_L0/internal/config"
	"Wb_Test_L0/internal/models"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
)

type Consumer struct {
	consumer *kafka.Consumer
	handler  Handler
	stop     bool
}

type Handler interface {
	SaveOrder(order models.Order) error
}

func NewConsumer(handler Handler, cfg *config.Config) (*Consumer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(cfg.Kafka.Brokers, ","),
		"group.id":                 cfg.Kafka.GroupID,
		"session.timeout.ms":       7000,
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest",
	}
	c, err := kafka.NewConsumer(conf)
	if err != nil {
		return nil, err
	}
	if err := c.Subscribe(cfg.Kafka.Topic, nil); err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: c,
		handler:  handler,
	}, nil
}

func (c *Consumer) Start() {
	for {
		if c.stop {
			break
		}
		kafkaMsg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			log.Println("kafka read error:", err)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(kafkaMsg.Value, &order); err != nil {
			log.Printf("invalid order json: %s", err)
			continue
		}

		if err := c.handler.SaveOrder(order); err != nil {
			log.Printf("error save order: %s", err)
			continue
		}

		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Printf("error store message: %s", err)
		} else {
			if _, err := c.consumer.CommitMessage(kafkaMsg); err != nil {
				log.Printf("commit failed: %s", err)
			} else {
				log.Printf("committed offset for message (topic=%s partition=%d offset=%v)", *kafkaMsg.TopicPartition.Topic, kafkaMsg.TopicPartition.Partition, kafkaMsg.TopicPartition.Offset)
			}
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}
	log.Printf("Committed offset")
	return c.consumer.Close()
}
