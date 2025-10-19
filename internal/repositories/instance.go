package repositories

import (
	"context"

	"github.com/Vighnesh-V-H/async/internal/models"
	"gorm.io/gorm"
)

type InstanceRepository struct {
	db *gorm.DB
}

func NewInstanceRepository(db *gorm.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

func (r *InstanceRepository) Create(ctx context.Context, instance *models.WorkflowInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

func (r *InstanceRepository) GetByExecutionID(ctx context.Context, executionID string) (*models.WorkflowInstance, error) {
	var instance models.WorkflowInstance
	err := r.db.WithContext(ctx).Where("execution_id = ?", executionID).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *InstanceRepository) UpdateStep(ctx context.Context, executionID string, step uint8, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.WorkflowInstance{}).
		Where("execution_id = ?", executionID).
		Updates(map[string]interface{}{
			"current_step": step,
			"status":       status,
		}).Error
}

func (r *InstanceRepository) UpdateStatus(ctx context.Context, executionID string, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.WorkflowInstance{}).
		Where("execution_id = ?", executionID).
		Update("status", status).Error
}
