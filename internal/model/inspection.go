package model

import (
	"errors"
	"time"
)

var (
	ErrInspectionNotFound = errors.New("inspection task not found")
	ErrInspectionState    = errors.New("invalid inspection state transition")
)

// InspectionStatus 巡检任务状态机：due -> in_progress -> completed / cancelled。
type InspectionStatus string

const (
	InspectionDue        InspectionStatus = "due"
	InspectionInProgress InspectionStatus = "in_progress"
	InspectionCompleted  InspectionStatus = "completed"
	InspectionCancelled  InspectionStatus = "cancelled"
)

// CanTransition 校验巡检状态迁移。
func (s InspectionStatus) CanTransition(to InspectionStatus) bool {
	switch s {
	case InspectionDue:
		return to == InspectionInProgress || to == InspectionCancelled
	case InspectionInProgress:
		return to == InspectionCompleted || to == InspectionCancelled
	case InspectionCompleted, InspectionCancelled:
		return false
	}
	return false
}

// InspectionTask 是一次站点巡检任务。完成时必须写 completion notes。
type InspectionTask struct {
	ID          string           `json:"id"`
	StationID   string           `json:"station_id"`
	DueDate     time.Time        `json:"due_date"`
	Status      InspectionStatus `json:"status"`
	Assignee    string           `json:"assignee,omitempty"`
	Notes       string           `json:"notes,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

var ErrNotesRequired = errors.New("completion notes required")
