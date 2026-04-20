package task

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type mockRepository struct {
	createFn  func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	getByIDFn func(ctx context.Context, id int64) (*taskdomain.Task, error)
	updateFn  func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	deleteFn  func(ctx context.Context, id int64) error
	listFn    func(ctx context.Context) ([]taskdomain.Task, error)
}

func (m *mockRepository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	if m.createFn != nil {
		return m.createFn(ctx, task)
	}
	return nil, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, task)
	}
	return nil, nil
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      CreateInput
		wantStatus taskdomain.Status
		wantType   taskdomain.RecurrenceType
		wantErr    bool
	}{
		{
			name: "valid daily every n",
			input: CreateInput{
				Title:            "Daily task",
				Description:      "desc",
				RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
				RecurrenceConfig: json.RawMessage(`{"every":2}`),
			},
			wantStatus: taskdomain.StatusNew,
			wantType:   taskdomain.RecurrenceDailyEveryN,
			wantErr:    false,
		},
		{
			name: "invalid daily every n",
			input: CreateInput{
				Title:            "Bad daily task",
				Description:      "desc",
				RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
				RecurrenceConfig: json.RawMessage(`{"every":0}`),
			},
			wantErr: true,
		},
		{
			name: "valid monthly day",
			input: CreateInput{
				Title:            "Monthly task",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				RecurrenceType:   taskdomain.RecurrenceMonthlyDay,
				RecurrenceConfig: json.RawMessage(`{"day":15}`),
			},
			wantStatus: taskdomain.StatusInProgress,
			wantType:   taskdomain.RecurrenceMonthlyDay,
			wantErr:    false,
		},
		{
			name: "invalid monthly day",
			input: CreateInput{
				Title:            "Bad monthly task",
				Description:      "desc",
				RecurrenceType:   taskdomain.RecurrenceMonthlyDay,
				RecurrenceConfig: json.RawMessage(`{"day":31}`),
			},
			wantErr: true,
		},
		{
			name: "none recurrence",
			input: CreateInput{
				Title:          "One time task",
				Description:    "desc",
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			wantStatus: taskdomain.StatusNew,
			wantType:   taskdomain.RecurrenceNone,
			wantErr:    false,
		},
		{
			name: "invalid recurrence type",
			input: CreateInput{
				Title:            "Unknown recurrence",
				Description:      "desc",
				RecurrenceType:   taskdomain.RecurrenceType("weird"),
				RecurrenceConfig: json.RawMessage(`{}`),
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			input: CreateInput{
				Title:          "Task with bad status",
				Description:    "desc",
				Status:         taskdomain.Status("bad"),
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			wantErr: true,
		},
		{
			name: "empty title",
			input: CreateInput{
				Title:          "   ",
				Description:    "desc",
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := validateCreateInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Status != tt.wantStatus {
				t.Fatalf("expected status %q, got %q", tt.wantStatus, got.Status)
			}

			if got.RecurrenceType != tt.wantType {
				t.Fatalf("expected recurrence type %q, got %q", tt.wantType, got.RecurrenceType)
			}
		})
	}
}

func TestValidateUpdateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    UpdateInput
		wantType taskdomain.RecurrenceType
		wantErr  bool
	}{
		{
			name: "valid specific dates",
			input: UpdateInput{
				Title:            "Specific dates task",
				Description:      "desc",
				Status:           taskdomain.StatusNew,
				RecurrenceType:   taskdomain.RecurrenceSpecificDates,
				RecurrenceConfig: json.RawMessage(`{"dates":["2026-04-20T10:00:00Z","2026-05-01T10:00:00Z"]}`),
			},
			wantType: taskdomain.RecurrenceSpecificDates,
			wantErr:  false,
		},
		{
			name: "invalid specific dates config",
			input: UpdateInput{
				Title:            "Broken specific dates task",
				Description:      "desc",
				Status:           taskdomain.StatusNew,
				RecurrenceType:   taskdomain.RecurrenceSpecificDates,
				RecurrenceConfig: json.RawMessage(`{"dates":"bad"}`),
			},
			wantErr: true,
		},
		{
			name: "valid even days",
			input: UpdateInput{
				Title:          "Even days task",
				Description:    "desc",
				Status:         taskdomain.StatusNew,
				RecurrenceType: taskdomain.RecurrenceEvenDays,
			},
			wantType: taskdomain.RecurrenceEvenDays,
			wantErr:  false,
		},
		{
			name: "invalid status",
			input: UpdateInput{
				Title:          "Task",
				Description:    "desc",
				Status:         taskdomain.Status("wrong"),
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			wantErr: true,
		},
		{
			name: "empty title",
			input: UpdateInput{
				Title:          "   ",
				Description:    "desc",
				Status:         taskdomain.StatusNew,
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := validateUpdateInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.RecurrenceType != tt.wantType {
				t.Fatalf("expected recurrence type %q, got %q", tt.wantType, got.RecurrenceType)
			}
		})
	}
}

func TestService_Complete(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		id            int64
		task          *taskdomain.Task
		getErr        error
		updateErr     error
		wantStatus    taskdomain.Status
		wantScheduled time.Time
		wantErr       bool
	}{
		{
			name: "non recurring task marked done",
			id:   1,
			task: &taskdomain.Task{
				ID:             1,
				Title:          "One time task",
				Description:    "desc",
				Status:         taskdomain.StatusInProgress,
				ScheduledAt:    time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			wantStatus:    taskdomain.StatusDone,
			wantScheduled: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "daily recurrence reschedules task",
			id:   2,
			task: &taskdomain.Task{
				ID:               2,
				Title:            "Daily task",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				ScheduledAt:      time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
				RecurrenceConfig: json.RawMessage(`{"every":2}`),
			},
			wantStatus:    taskdomain.StatusNew,
			wantScheduled: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "monthly recurrence reschedules task",
			id:   3,
			task: &taskdomain.Task{
				ID:               3,
				Title:            "Monthly task",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				ScheduledAt:      time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				RecurrenceType:   taskdomain.RecurrenceMonthlyDay,
				RecurrenceConfig: json.RawMessage(`{"day":15}`),
			},
			wantStatus:    taskdomain.StatusNew,
			wantScheduled: time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "specific dates recurrence reschedules to next occurrence",
			id:   4,
			task: &taskdomain.Task{
				ID:               4,
				Title:            "Specific dates task",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				ScheduledAt:      time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				RecurrenceType:   taskdomain.RecurrenceSpecificDates,
				RecurrenceConfig: json.RawMessage(`{"dates":["2026-04-20T10:00:00Z","2026-04-25T10:00:00Z","2026-05-01T10:00:00Z"]}`),
			},
			wantStatus:    taskdomain.StatusNew,
			wantScheduled: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "specific dates without next occurrence marks done",
			id:   5,
			task: &taskdomain.Task{
				ID:               5,
				Title:            "Specific dates exhausted task",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				ScheduledAt:      time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
				RecurrenceType:   taskdomain.RecurrenceSpecificDates,
				RecurrenceConfig: json.RawMessage(`{"dates":["2026-04-20T10:00:00Z","2026-05-01T10:00:00Z"]}`),
			},
			wantStatus:    taskdomain.StatusDone,
			wantScheduled: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid id",
			id:      0,
			wantErr: true,
		},
		{
			name: "recurring task without scheduled_at returns error",
			id:   6,
			task: &taskdomain.Task{
				ID:               6,
				Title:            "Broken recurring task",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
				RecurrenceConfig: json.RawMessage(`{"every":2}`),
			},
			wantErr: true,
		},
		{
			name: "invalid recurrence config returns error",
			id:   7,
			task: &taskdomain.Task{
				ID:               7,
				Title:            "Bad recurrence config",
				Description:      "desc",
				Status:           taskdomain.StatusInProgress,
				ScheduledAt:      time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
				RecurrenceConfig: json.RawMessage(`{"every":"bad"}`),
			},
			wantErr: true,
		},
		{
			name:    "repository get by id error",
			id:      8,
			getErr:  errors.New("repo get error"),
			wantErr: true,
		},
		{
			name: "repository update error",
			id:   9,
			task: &taskdomain.Task{
				ID:             9,
				Title:          "One time task",
				Description:    "desc",
				Status:         taskdomain.StatusNew,
				ScheduledAt:    time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				RecurrenceType: taskdomain.RecurrenceNone,
			},
			updateErr: errors.New("repo update error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var updatedTask *taskdomain.Task

			repo := &mockRepository{
				getByIDFn: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return tt.task, nil
				},
				updateFn: func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					updatedTask = task
					return task, nil
				},
			}

			svc := NewService(repo)
			svc.now = func() time.Time { return now }

			got, err := svc.Complete(context.Background(), tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if updatedTask == nil {
				t.Fatal("expected update to be called")
			}

			if got.Status != tt.wantStatus {
				t.Fatalf("expected status %q, got %q", tt.wantStatus, got.Status)
			}

			if !got.ScheduledAt.Equal(tt.wantScheduled) {
				t.Fatalf("expected scheduled_at %v, got %v", tt.wantScheduled, got.ScheduledAt)
			}

			if !got.UpdatedAt.Equal(now) {
				t.Fatalf("expected updated_at %v, got %v", now, got.UpdatedAt)
			}

			if updatedTask.Status != tt.wantStatus {
				t.Fatalf("updated task status = %q, want %q", updatedTask.Status, tt.wantStatus)
			}

			if !updatedTask.ScheduledAt.Equal(tt.wantScheduled) {
				t.Fatalf("updated task scheduled_at = %v, want %v", updatedTask.ScheduledAt, tt.wantScheduled)
			}

			if !updatedTask.UpdatedAt.Equal(now) {
				t.Fatalf("updated task updated_at = %v, want %v", updatedTask.UpdatedAt, now)
			}
		})
	}
}