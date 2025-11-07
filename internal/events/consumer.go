package events

import (
	"context"
	"encoding/json"

	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

type EventConsumer struct {
	consumer *kafka.Consumer
	logger   zerolog.Logger
}

type CompletionHandler func(ctx context.Context, event *CompletionEvent) error

func NewEventConsumer(consumer *kafka.Consumer, logCfg logger.Config) *EventConsumer {
	return &EventConsumer{
		consumer: consumer,
		logger:   logger.New(logCfg),
	}
}

func (ec *EventConsumer) ConsumeCompletions(ctx context.Context, handler CompletionHandler) error {
	ec.logger.Info().Msg("Starting to consume completion events")

	for {
		select {
		case <-ctx.Done():
			ec.logger.Info().Msg("Context cancelled, stopping consumer")
			return ctx.Err()
		default:
			msg, err := ec.consumer.ReadMessage(-1)
			if err != nil {
				ec.logger.Error().Err(err).Msg("Error reading message from Kafka")
				continue
			}

			var completion CompletionEvent
			if err := json.Unmarshal(msg.Value, &completion); err != nil {
				ec.logger.Error().
					Err(err).
					Str("message", string(msg.Value)).
					Msg("Failed to unmarshal completion event")
				continue
			}

			ec.logger.Info().
				Str("execution_id", completion.ExecutionID).
				Str("task_type", completion.TaskType).
				Str("status", completion.Status).
				Uint8("step", completion.Step).
				Msg("Received completion event")

			if err := handler(ctx, &completion); err != nil {
				ec.logger.Error().
					Err(err).
					Str("execution_id", completion.ExecutionID).
					Msg("Failed to process completion event")
				continue
			}

			if _, err := ec.consumer.CommitMessage(msg); err != nil {
				ec.logger.Error().Err(err).Msg("Failed to commit offset")
			}
		}
	}
}

type TaskHandler func(ctx context.Context, event *TaskEvent) error


func (ec *EventConsumer) ConsumeTasks(ctx context.Context, handler TaskHandler) error {
	ec.logger.Info().Msg("Starting to consume task events")

	for {
		select {
		case <-ctx.Done():
			ec.logger.Info().Msg("Context cancelled, stopping task consumer")
			return ctx.Err()
		default:
			msg, err := ec.consumer.ReadMessage(-1)
			if err != nil {
				ec.logger.Error().Err(err).Msg("Error reading message from Kafka")
				continue
			}

			var task TaskEvent
			if err := json.Unmarshal(msg.Value, &task); err != nil {
				ec.logger.Error().
					Err(err).
					Str("message", string(msg.Value)).
					Msg("Failed to unmarshal task event")
				continue
			}

			ec.logger.Info().
				Str("execution_id", task.ExecutionID).
				Str("task_type", task.TaskType).
				Uint8("step", task.Step).
				Msg("Received task event")

			if err := handler(ctx, &task); err != nil {
				ec.logger.Error().
					Err(err).
					Str("execution_id", task.ExecutionID).
					Msg("Failed to process task event")
				continue
			}

			if _, err := ec.consumer.CommitMessage(msg); err != nil {
				ec.logger.Error().Err(err).Msg("Failed to commit offset")
			}
		}
	}
}

// Close closes the consumer
func (ec *EventConsumer) Close() error {
	if ec.consumer != nil {
		return ec.consumer.Close()
	}
	return nil
}
