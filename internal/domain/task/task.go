package task

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID               int64
	Title            string
	Description      string
	Status           Status
	ScheduledAt      time.Time
	RecurrenceType   RecurrenceType
	RecurrenceConfig json.RawMessage
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
