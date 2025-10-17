package models

import (
	"time"
)

type Workflow struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"uniqueIndex;size:255" json:"name"`
    Event     string    `gorm:"index" json:"event"`
    Status    string    `gorm:"size:50" json:"status"`
    Message   string    `gorm:"size:500" json:"message"`
    Payload   string    `json:"payload"`
    Steps     uint8     `json:"steps"`
    HandlerURL string   `gorm:"size:500" json:"handler_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type WorkflowInstance struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    WorkflowID  uint           `gorm:"index" json:"workflow_id"`
    ExecutionID string         `gorm:"uniqueIndex;size:100" json:"execution_id"`
    Status      string         `gorm:"size:50;index" json:"status"`
    Variables   []byte         `gorm:"type:jsonb" json:"variables"`
    CurrentStep uint8          `json:"current_step"`
    History     []HistoryEntry `gorm:"foreignKey:InstanceID" json:"-"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

type Task struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    InstanceID uint      `gorm:"index" json:"instance_id"`
    StepID     uint8     `json:"step_id"`
    Type       string    `gorm:"size:50" json:"type"`
    Payload    []byte    `gorm:"type:jsonb" json:"payload"`
    Output     []byte    `gorm:"type:jsonb" json:"output"`
    Status     string    `gorm:"size:50;index" json:"status"`
    Retries    uint8     `json:"retries_left"`
    TimeoutAt  *time.Time `json:"timeout_at"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
}

type HistoryEntry struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    InstanceID uint     `gorm:"index" json:"instance_id"`
    Event     string    `gorm:"size:100" json:"event"`
    Timestamp time.Time `json:"timestamp"`
    Data      []byte    `gorm:"type:jsonb" json:"data"`
}

type WorkflowRegistry struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    WorkflowID uint      `gorm:"index" json:"workflow_id"`
    Trigger    string    `gorm:"uniqueIndex" json:"trigger"`
    HandlerURL string    `gorm:"size:500" json:"handler_url"`
    Version    string    `gorm:"size:20" json:"version"`
    LeaseExpiresAt *time.Time `json:"lease_expires_at"`
    LastHeartbeat time.Time `json:"last_heartbeat"`
    CreatedAt   time.Time  `json:"created_at"`
}