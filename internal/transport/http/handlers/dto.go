package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title            string                       `json:"title"`
	Description      string                       `json:"description"`
	Status           taskdomain.Status            `json:"status"`
	RecurrenceType   string                       `json:"recurrence_type,omitempty"`
	RecurrenceConfig *taskdomain.RecurrenceConfig `json:"recurrence_config,omitempty"`
}

type taskDTO struct {
	ID               int64                        `json:"id"`
	Title            string                       `json:"title"`
	Description      string                       `json:"description"`
	Status           taskdomain.Status            `json:"status"`
	CreatedAt        time.Time                    `json:"created_at"`
	UpdatedAt        time.Time                    `json:"updated_at"`
	RecurrenceType   string                       `json:"recurrence_type,omitempty"`
	RecurrenceConfig *taskdomain.RecurrenceConfig `json:"recurrence_config,omitempty"`
	ParentTaskID     *int64                       `json:"parent_task_id,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	if task.RecurrenceType != "" && task.RecurrenceType != "none" {
		dto.RecurrenceType = task.RecurrenceType
		dto.RecurrenceConfig = &task.RecurrenceConfig
	}

	if task.ParentTaskID != nil {
		dto.ParentTaskID = task.ParentTaskID
	}

	return dto
}
