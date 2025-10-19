package service

import (
	"context"
	"encoding/json"

	"github.com/Vighnesh-V-H/async/internal/models"
	"github.com/Vighnesh-V-H/async/internal/repositories"
)

type InstanceService struct {
	repo *repositories.InstanceRepository
}

func NewInstanceService(repo *repositories.InstanceRepository) *InstanceService {
	return &InstanceService{repo: repo}
}

func (s *InstanceService) CreateInstance(ctx context.Context, instance *models.WorkflowInstance, variables map[string]interface{}) error {
	varsJSON, err := json.Marshal(variables)
	if err != nil {
		return err
	}
	instance.Variables = varsJSON
	return s.repo.Create(ctx, instance)
}

func (s *InstanceService) GetInstanceByExecutionID(ctx context.Context, executionID string) (*models.WorkflowInstance, error) {
	return s.repo.GetByExecutionID(ctx, executionID)
}

func (s *InstanceService) UpdateInstanceStep(ctx context.Context, executionID string, step uint8, status string) error {
	return s.repo.UpdateStep(ctx, executionID, step, status)
}

func (s *InstanceService) UpdateInstanceStatus(ctx context.Context, executionID string, status string) error {
	return s.repo.UpdateStatus(ctx, executionID, status)
}
