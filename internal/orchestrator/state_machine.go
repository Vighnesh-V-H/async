package orchestrator

import (
	"context"
	"time"

	"github.com/Vighnesh-V-H/async/internal/events"
	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/Vighnesh-V-H/async/internal/service"
	"github.com/rs/zerolog"
)

type Orchestrator struct {
	workflowSvc   *service.WorkflowService
	instanceSvc   *service.InstanceService
	eventProducer *events.EventProducer
	logger        zerolog.Logger
}

func NewOrchestrator(
	workflowSvc *service.WorkflowService,
	instanceSvc *service.InstanceService,
	eventProducer *events.EventProducer,
	logCfg logger.Config,
) *Orchestrator {
	return &Orchestrator{
		workflowSvc:   workflowSvc,
		instanceSvc:   instanceSvc,
		eventProducer: eventProducer,
		logger:        logger.New(logCfg),
	}
}

// ProcessCompletion handles completion events and orchestrates the next step
func (o *Orchestrator) ProcessCompletion(ctx context.Context, completion *events.CompletionEvent) error {
	o.logger.Info().
		Str("execution_id", completion.ExecutionID).
		Str("task_type", completion.TaskType).
		Uint8("step", completion.Step).
		Str("status", completion.Status).
		Msg("Processing completion event")

	// 1. Update instance step status
	if err := o.instanceSvc.UpdateInstanceStep(ctx, completion.ExecutionID, completion.Step, "COMPLETED"); err != nil {
		o.logger.Error().Err(err).Msg("Failed to update instance step")
		return err
	}

	// 2. Check if task failed
	if completion.Status == "failed" {
		o.logger.Error().
			Str("execution_id", completion.ExecutionID).
			Str("error", completion.Error).
			Msg("Task failed, marking workflow as FAILED")
		return o.instanceSvc.UpdateInstanceStatus(ctx, completion.ExecutionID, "FAILED")
	}

	// 3. Get workflow to determine next step
	workflow, err := o.workflowSvc.GetWorkflowByEvent(ctx, "generate_audio")
	if err != nil {
		o.logger.Error().Err(err).Msg("Failed to get workflow")
		return err
	}

	// 4. Determine next task based on current step
	nextTask := o.getNextTask(completion.Step, workflow.Steps)

	if nextTask == nil {
		// Workflow completed
		o.logger.Info().
			Str("execution_id", completion.ExecutionID).
			Msg("Workflow completed successfully")
		return o.instanceSvc.UpdateInstanceStatus(ctx, completion.ExecutionID, "COMPLETED")
	}

	// 5. Publish next task event
	taskEvent := &events.TaskEvent{
		ExecutionID: completion.ExecutionID,
		WorkflowID:  completion.WorkflowID,
		TaskType:    nextTask.TaskType,
		Step:        nextTask.Step,
		Input:       completion.Output, // Pass output from previous task as input
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	topic := "task-queue"
	if err := o.eventProducer.PublishTask(topic, taskEvent); err != nil {
		o.logger.Error().Err(err).Msg("Failed to publish next task")
		return err
	}

	o.logger.Info().
		Str("execution_id", completion.ExecutionID).
		Str("next_task", nextTask.TaskType).
		Uint8("next_step", nextTask.Step).
		Msg("Published next task")

	return nil
}

// NextTask represents the next task in the workflow
type NextTask struct {
	TaskType string
	Step     uint8
}

// getNextTask determines the next task based on the workflow definition
func (o *Orchestrator) getNextTask(currentStep uint8, totalSteps uint8) *NextTask {
	// Hardcoded workflow for audio generation (3 steps)
	// Step 1: extract_text
	// Step 2: generate_audio
	// Step 3: store_audio

	nextStep := currentStep + 1
	if nextStep > totalSteps {
		return nil // No more steps
	}

	var taskType string
	switch nextStep {
	case 1:
		taskType = "extract_text"
	case 2:
		taskType = "generate_audio"
	case 3:
		taskType = "store_audio"
	default:
		return nil
	}

	return &NextTask{
		TaskType: taskType,
		Step:     nextStep,
	}
}
