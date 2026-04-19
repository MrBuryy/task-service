package handlers

import (
	"encoding/json"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	Status           taskdomain.Status       `json:"status"`
	ScheduledAt      time.Time               `json:"scheduled_at"`
	RecurrenceType   taskdomain.RecurrenceType `json:"recurrence_type"`
	RecurrenceConfig json.RawMessage         `json:"recurrence_config"`
}

type taskDTO struct {
	ID               int64                     `json:"id"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	Status           taskdomain.Status         `json:"status"`
	ScheduledAt      time.Time                 `json:"scheduled_at"`
	RecurrenceType   taskdomain.RecurrenceType `json:"recurrence_type"`
	RecurrenceConfig json.RawMessage           `json:"recurrence_config"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:               task.ID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		ScheduledAt:      task.ScheduledAt,
		RecurrenceType:   task.RecurrenceType,
		RecurrenceConfig: task.RecurrenceConfig,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
}
