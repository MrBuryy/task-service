package handlers

import (
	"encoding/json"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type   taskdomain.RecurrenceType `json:"type"`
	Config json.RawMessage           `json:"config"`
}

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	ScheduledAt time.Time         `json:"scheduled_at"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	ScheduledAt time.Time         `json:"scheduled_at"`
	Recurrence  recurrenceDTO     `json:"recurrence"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		ScheduledAt: task.ScheduledAt,
		Recurrence: recurrenceDTO{
			Type:   task.RecurrenceType,
			Config: task.RecurrenceConfig,
		},
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}
}