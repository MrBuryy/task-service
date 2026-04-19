package task

import (
	"errors"
	"testing"
	"time"
)

func intPtr(v int) *int {
	return &v
}

func TestRecurrenceRule(t *testing.T) {
	t.Run("none returns nil rule and nil error", func(t *testing.T) {
		r := Recurrence{Type: RecurrenceNone}

		rule, err := r.Rule()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if rule != nil {
			t.Fatalf("expected nil rule, got %#v", rule)
		}
	})

	t.Run("daily returns DailyEveryNRule", func(t *testing.T) {
		r := Recurrence{
			Type:       RecurrenceDailyEveryN,
			EveryNDays: intPtr(3),
		}

		rule, err := r.Rule()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		daily, ok := rule.(DailyEveryNRule)
		if !ok {
			t.Fatalf("expected DailyEveryNRule, got %T", rule)
		}
		if daily.Every != 3 {
			t.Fatalf("expected Every=3, got %d", daily.Every)
		}
	})

	t.Run("daily without every returns error", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceDailyEveryN,
		}

		rule, err := r.Rule()
		if !errors.Is(err, ErrInvalidEvery) {
			t.Fatalf("expected ErrInvalidEvery, got %v", err)
		}
		if rule != nil {
			t.Fatalf("expected nil rule, got %#v", rule)
		}
	})

	t.Run("monthly returns MonthlyDayRule", func(t *testing.T) {
		r := Recurrence{
			Type:       RecurrenceMonthlyDay,
			DayOfMonth: intPtr(15),
		}

		rule, err := r.Rule()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		monthly, ok := rule.(MonthlyDayRule)
		if !ok {
			t.Fatalf("expected MonthlyDayRule, got %T", rule)
		}
		if monthly.Day != 15 {
			t.Fatalf("expected Day=15, got %d", monthly.Day)
		}
	})

	t.Run("monthly without day returns error", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceMonthlyDay,
		}

		rule, err := r.Rule()
		if !errors.Is(err, ErrInvalidDayOfMonth) {
			t.Fatalf("expected ErrInvalidDayOfMonth, got %v", err)
		}
		if rule != nil {
			t.Fatalf("expected nil rule, got %#v", rule)
		}
	})

	t.Run("unknown type returns error", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceSpecificDates,
		}

		rule, err := r.Rule()
		if !errors.Is(err, ErrUnknownRecurrenceType) {
			t.Fatalf("expected ErrUnknownRecurrenceType, got %v", err)
		}
		if rule != nil {
			t.Fatalf("expected nil rule, got %#v", rule)
		}
	})
}

func TestDailyEveryNRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    DailyEveryNRule
		wantErr error
	}{
		{
			name:    "valid every 1",
			rule:    DailyEveryNRule{Every: 1},
			wantErr: nil,
		},
		{
			name:    "valid every 3",
			rule:    DailyEveryNRule{Every: 3},
			wantErr: nil,
		},
		{
			name:    "invalid zero",
			rule:    DailyEveryNRule{Every: 0},
			wantErr: ErrInvalidEvery,
		},
		{
			name:    "invalid negative",
			rule:    DailyEveryNRule{Every: -1},
			wantErr: ErrInvalidEvery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestDailyEveryNRuleNext(t *testing.T) {
	from := time.Date(2026, 4, 19, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		rule    DailyEveryNRule
		want    time.Time
		wantErr error
	}{
		{
			name:    "every 1 day",
			rule:    DailyEveryNRule{Every: 1},
			want:    time.Date(2026, 4, 20, 10, 30, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "every 3 days",
			rule:    DailyEveryNRule{Every: 3},
			want:    time.Date(2026, 4, 22, 10, 30, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "invalid every",
			rule:    DailyEveryNRule{Every: 0},
			want:    time.Time{},
			wantErr: ErrInvalidEvery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rule.Next(from)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("expected time %v, got %v", tt.want, got)
			}
		})
	}
}

func TestMonthlyDayRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    MonthlyDayRule
		wantErr error
	}{
		{
			name:    "valid day 1",
			rule:    MonthlyDayRule{Day: 1},
			wantErr: nil,
		},
		{
			name:    "valid day 15",
			rule:    MonthlyDayRule{Day: 15},
			wantErr: nil,
		},
		{
			name:    "valid day 30",
			rule:    MonthlyDayRule{Day: 30},
			wantErr: nil,
		},
		{
			name:    "invalid day 0",
			rule:    MonthlyDayRule{Day: 0},
			wantErr: ErrInvalidDayOfMonth,
		},
		{
			name:    "invalid day 31",
			rule:    MonthlyDayRule{Day: 31},
			wantErr: ErrInvalidDayOfMonth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestMonthlyDayRuleNext(t *testing.T) {
	tests := []struct {
		name    string
		from    time.Time
		rule    MonthlyDayRule
		want    time.Time
		wantErr error
	}{
		{
			name:    "same month when target day is ahead",
			from:    time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
			rule:    MonthlyDayRule{Day: 20},
			want:    time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "next month when target day equals current day",
			from:    time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC),
			rule:    MonthlyDayRule{Day: 20},
			want:    time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "next month when target day is behind",
			from:    time.Date(2026, 4, 25, 9, 0, 0, 0, time.UTC),
			rule:    MonthlyDayRule{Day: 20},
			want:    time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "cross year boundary",
			from:    time.Date(2026, 12, 25, 8, 15, 0, 0, time.UTC),
			rule:    MonthlyDayRule{Day: 10},
			want:    time.Date(2027, 1, 10, 8, 15, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "skip invalid february 30 and go to march",
			from:    time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
			rule:    MonthlyDayRule{Day: 30},
			want:    time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC),
			wantErr: nil,
		},
		{
			name:    "invalid day",
			from:    time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
			rule:    MonthlyDayRule{Day: 31},
			want:    time.Time{},
			wantErr: ErrInvalidDayOfMonth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rule.Next(tt.from)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("expected time %v, got %v", tt.want, got)
			}
		})
	}
}