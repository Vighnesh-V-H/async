package handler

import (
	"net/http"
	"time"

	"github.com/Vighnesh-V-H/async/internal/events"
	"github.com/Vighnesh-V-H/async/internal/models"
	"github.com/Vighnesh-V-H/async/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AudioHandler struct {
	workflowSvc  *service.WorkflowService
	instanceSvc  *service.InstanceService
	eventProducer *events.EventProducer
}

func NewAudioHandler(
	workflowSvc *service.WorkflowService,
	instanceSvc *service.InstanceService,
	eventProducer *events.EventProducer,
) *AudioHandler {
	return &AudioHandler{
		workflowSvc:   workflowSvc,
		instanceSvc:   instanceSvc,
		eventProducer: eventProducer,
	}
}

type GenerateAudioRequest struct {
	Text     string                 `json:"text" binding:"required"`
	Voice    string                 `json:"voice"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *AudioHandler) GenerateAudio(c *gin.Context) {
	var req GenerateAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	workflow, err := h.workflowSvc.GetWorkflowByEvent(ctx, "generate_audio")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workflow not found"})
		return
	}

	executionID := uuid.New().String()

	input := map[string]interface{}{
		"text":     req.Text,
		"voice":    req.Voice,
		"metadata": req.Metadata,
	}

	instance := &models.WorkflowInstance{
		WorkflowID:  workflow.ID,
		ExecutionID: executionID,
		Status:      "PENDING",
		CurrentStep: 0,
	}

	if err := h.instanceSvc.CreateInstance(ctx, instance, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow instance"})
		return
	}

	taskEvent := &events.TaskEvent{
		ExecutionID: executionID,
		WorkflowID:  workflow.ID,
		TaskType:    "extract_text", 
		Step:        1,
		Input:       input,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	topic := "task-queue"
	if err := h.eventProducer.PublishTask(topic, taskEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish task event"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": executionID,
		"workflow":     workflow.Name,
		"status":       "PENDING",
		"message":      "Audio generation workflow triggered successfully",
	})
}

func (h *AudioHandler) GetStatus(c *gin.Context) {
	executionID := c.Param("execution_id")

	ctx := c.Request.Context()
	instance, err := h.instanceSvc.GetInstanceByExecutionID(ctx, executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id": instance.ExecutionID,
		"workflow_id":  instance.WorkflowID,
		"status":       instance.Status,
		"current_step": instance.CurrentStep,
		"created_at":   instance.CreatedAt,
		"updated_at":   instance.UpdatedAt,
	})
}
