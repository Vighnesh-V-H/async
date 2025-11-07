package kafka

import (
	"strings"
	"sync"

	config "github.com/Vighnesh-V-H/async/configs"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

var (
	producer     *kafka.Producer
	consumer     *kafka.Consumer
	initOnce     sync.Once
	initConsumer sync.Once
	logg         zerolog.Logger
	kafkaCfg     *config.KafkaConfig
)

func SetLogger(z zerolog.Logger) {
	logg = z
}

func SetConfig(cfg *config.KafkaConfig) {
	kafkaCfg = cfg
}

func InitProducer(cfg *config.KafkaConfig, log zerolog.Logger) (*kafka.Producer, error) {
	var err error
	logg = log
	kafkaCfg = cfg
	
	initOnce.Do(func() {
		p, e := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":    strings.Join(cfg.Brokers, ","),
			"acks":                 cfg.ProducerAcks,
			"retries":              cfg.ProducerRetries,
			"linger.ms":            cfg.ProducerLingerMs,
			"compression.type":     cfg.CompressionType,
			"request.timeout.ms":   cfg.RequestTimeoutMs,
			"delivery.timeout.ms":  cfg.DeliveryTimeoutMs,
		})
		if e != nil {
			err = e
			logg.Error().Err(e).Msg("Failed to create Kafka producer")
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
		logg.Info().
			Strs("brokers", cfg.Brokers).
			Str("acks", cfg.ProducerAcks).
			Int("retries", cfg.ProducerRetries).
			Msg("Kafka Producer initialized")
	})
	return producer, err
}

func InitConsumer(cfg *config.KafkaConfig, log zerolog.Logger) (*kafka.Consumer, error) {
	var err error
	logg = log
	kafkaCfg = cfg
	
	initConsumer.Do(func() {
		c, e := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  strings.Join(cfg.Brokers, ","),
			"group.id":           cfg.ConsumerGroupID,
			"auto.offset.reset":  cfg.AutoOffsetReset,
			"enable.auto.commit": cfg.EnableAutoCommit,
			"session.timeout.ms": cfg.SessionTimeoutMs,
		})
		if e != nil {
			err = e
			logg.Error().Err(e).Msg("Failed to create Kafka consumer")
			return
		}
		if err = c.SubscribeTopics([]string{cfg.ConsumerTopic}, nil); err != nil {
			logg.Error().Err(err).Str("topic", cfg.ConsumerTopic).Msg("Failed to subscribe to topic")
			return
		}
		consumer = c
		logg.Info().
			Str("topic", cfg.ConsumerTopic).
			Str("group_id", cfg.ConsumerGroupID).
			Msg("Kafka Consumer initialized")
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
