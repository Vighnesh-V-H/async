package service

import (
	"context"

	"github.com/Vighnesh-V-H/async/internal/models"
	"github.com/Vighnesh-V-H/async/internal/repositories"
)

type WorkflowService struct {
    repo *repositories.WorkflowRepository
}

func NewWorkflowService(repo *repositories.WorkflowRepository) *WorkflowService {
    return &WorkflowService{repo: repo}
}

type CreateWorkflowRequest struct {
    Name      string `json:"name"`
    Event     string `json:"event"`
    Message   string `json:"message"`
    HandlerURL string `json:"handler_url"`
    Steps     uint8  `json:"steps"`
}

func (s *WorkflowService) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*models.Workflow, error) {
    wf := &models.Workflow{
        Name:       req.Name,
        Event:      req.Event,
        Message:    req.Message,
        Steps:      req.Steps,
        HandlerURL: req.HandlerURL,
        Status:     "active",
    }

    if err := s.repo.Create(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

// GetWorkflowByEvent retrieves a workflow by event name
func (s *WorkflowService) GetWorkflowByEvent(ctx context.Context, event string) (*models.Workflow, error) {
    return s.repo.GetByEvent(ctx, event)
}
