package repositories

import (
	"context"

	"github.com/Vighnesh-V-H/async/internal/models"
	"gorm.io/gorm"
)

type WorkflowRepository struct {
    db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
    return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) Create(ctx context.Context, wf *models.Workflow) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(wf).Error; err != nil {
            return err
        }
        return nil
    })
}

// GetByEvent retrieves a workflow by event name
func (r *WorkflowRepository) GetByEvent(ctx context.Context, event string) (*models.Workflow, error) {
    var wf models.Workflow
    err := r.db.WithContext(ctx).Where("event = ? AND status = ?", event, "active").First(&wf).Error
    if err != nil {
        return nil, err
    }
    return &wf, nil
}
