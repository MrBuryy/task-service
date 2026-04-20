package task

import (
	"encoding/json"
	"testing"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       CreateInput
		wantStatus  taskdomain.Status
		wantType    taskdomain.RecurrenceType
		wantErr     bool
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