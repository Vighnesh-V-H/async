package kafka

import (
	"os"
	"sync"

	lg "github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

var (
	producer     *kafka.Producer
	consumer     *kafka.Consumer
	initOnce     sync.Once
	initConsumer sync.Once
	logg         zerolog.Logger
)


func SetLogger(z zerolog.Logger) {
	logg = z
}

func init() {
	
	cfg := lg.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Format:      os.Getenv("LOG_FORMAT"),
		ServiceName: "async-kafka",
		Environment: os.Getenv("ENVIRONMENT"),
		IsProd:      os.Getenv("ENVIRONMENT") == "prod" || os.Getenv("ENVIRONMENT") == "production",
	}
	logg = lg.New(cfg)
}

func InitProducer() (*kafka.Producer, error) {
	var err error
	initOnce.Do(func() {
		p, e := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":    os.Getenv("KAFKA_BROKERS"),
			"acks":                 "all",
			"retries":              3,
			"linger.ms":            5,
			"compression.type":     "gzip",
			"request.timeout.ms":   30000,
			"delivery.timeout.ms":  120000,
		})
		if e != nil {
			err = e
			return
		}
		
		go func() {
			for e := range p.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						logg.Error().
							Err(ev.TopicPartition.Error).
							Str("topic", *ev.TopicPartition.Topic).
							Msg("Failed to deliver message")
					} else {
						logg.Debug().
							Str("topic", *ev.TopicPartition.Topic).
							Int32("partition", ev.TopicPartition.Partition).
							Msg("Message delivered successfully")
					}
				}
			}
		}()
		
		producer = p
		logg.Info().Msg("Kafka Producer initialized")
	})
	return producer, err
}

func InitConsumer(groupID, topic string) (*kafka.Consumer, error) {
	var err error
	initConsumer.Do(func() {
		c, e := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  os.Getenv("KAFKA_BROKERS"),
			"group.id":           groupID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": false,
		})
		if e != nil {
			err = e
			return
		}
		if err = c.SubscribeTopics([]string{topic}, nil); err != nil {
			return
		}
		consumer = c
		logg.Info().Str("topic", topic).Msg("Kafka Consumer initialized")
	})
	return consumer, err
}

func GetProducer() *kafka.Producer {
	if producer == nil {
		panic("Kafka producer not initialized. Call InitProducer first.")
	}
	return producer
}

func GetConsumer() *kafka.Consumer {
	if consumer == nil {
		panic("Kafka consumer not initialized. Call InitConsumer first.")
	}
	return consumer
}

func Close() {
	if producer != nil {
		producer.Flush(5000)
		producer.Close()
	}
	if consumer != nil {
		consumer.Close()
	}
}
