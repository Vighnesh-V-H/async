package events

import (
	"encoding/json"
	"fmt"

	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

type EventProducer struct {
	producer *kafka.Producer
	logger   zerolog.Logger
}

type TaskEvent struct {
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  uint                   `json:"workflow_id"`
	TaskType    string                 `json:"task_type"`
	Step        uint8                  `json:"step"`
	Input       map[string]interface{} `json:"input"`
	Timestamp   string                 `json:"timestamp"`
}

type CompletionEvent struct {
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  uint                   `json:"workflow_id"`
	TaskType    string                 `json:"task_type"`
	Step        uint8                  `json:"step"`
	Status      string                 `json:"status"` // "success" or "failed"
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   string                 `json:"timestamp"`
}

func NewEventProducer(producer *kafka.Producer, logCfg logger.Config) *EventProducer {
	return &EventProducer{
		producer: producer,
		logger:   logger.New(logCfg),
	}
}

func (ep *EventProducer) PublishTask(topic string, event *TaskEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		ep.logger.Error().Err(err).Msg("Failed to marshal task event")
		return fmt.Errorf("failed to marshal task event: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.ExecutionID),
		Value: payload,
	}

	// Produce asynchronously - delivery reports will be handled by the Events() goroutine
	if err := ep.producer.Produce(msg, nil); err != nil {
		ep.logger.Error().Err(err).Str("topic", topic).Msg("Failed to produce task event")
		return fmt.Errorf("failed to produce task event: %w", err)
	}

	ep.logger.Info().
		Str("topic", topic).
		Str("execution_id", event.ExecutionID).
		Str("task_type", event.TaskType).
		Uint8("step", event.Step).
		Msg("Task event queued for publishing to Kafka")

	return nil
}

func (ep *EventProducer) PublishCompletion(topic string, event *CompletionEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		ep.logger.Error().Err(err).Msg("Failed to marshal completion event")
		return fmt.Errorf("failed to marshal completion event: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.ExecutionID),
		Value: payload,
	}

	// Produce asynchronously - delivery reports will be handled by the Events() goroutine
	if err := ep.producer.Produce(msg, nil); err != nil {
		ep.logger.Error().Err(err).Str("topic", topic).Msg("Failed to produce completion event")
		return fmt.Errorf("failed to produce completion event: %w", err)
	}

	ep.logger.Info().
		Str("topic", topic).
		Str("execution_id", event.ExecutionID).
		Str("task_type", event.TaskType).
		Str("status", event.Status).
		Uint8("step", event.Step).
		Msg("Completion event queued for publishing to Kafka")

	return nil
}
