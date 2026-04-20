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

	t.Run("specific dates returns SpecificDatesRule", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceSpecificDates,
			SpecificDates: []time.Time{
				time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			},
		}

		rule, err := r.Rule()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		specific, ok := rule.(SpecificDatesRule)
		if !ok {
			t.Fatalf("expected SpecificDatesRule, got %T", rule)
		}
		if len(specific.Dates) != 1 {
			t.Fatalf("expected 1 date, got %d", len(specific.Dates))
		}
	})

	t.Run("even days returns EvenDaysRule", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceEvenDays,
		}

		rule, err := r.Rule()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if _, ok := rule.(EvenDaysRule); !ok {
			t.Fatalf("expected EvenDaysRule, got %T", rule)
		}
	})

	t.Run("odd days returns OddDaysRule", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceOddDays,
		}

		rule, err := r.Rule()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if _, ok := rule.(OddDaysRule); !ok {
			t.Fatalf("expected OddDaysRule, got %T", rule)
		}
	})

	t.Run("unknown type returns error", func(t *testing.T) {
		r := Recurrence{
			Type: RecurrenceType("unexpected_type"),
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

func TestSpecificDatesRuleValidate(t *testing.T) {
	t.Run("empty dates", func(t *testing.T) {
		rule := SpecificDatesRule{}
		err := rule.Validate()
		if !errors.Is(err, ErrNoSpecificDates) {
			t.Fatalf("expected %v, got %v", ErrNoSpecificDates, err)
		}
	})

	t.Run("zero date", func(t *testing.T) {
		rule := SpecificDatesRule{
			Dates: []time.Time{time.Time{}},
		}
		err := rule.Validate()
		if !errors.Is(err, ErrInvalidSpecificDate) {
			t.Fatalf("expected %v, got %v", ErrInvalidSpecificDate, err)
		}
	})

	t.Run("valid dates", func(t *testing.T) {
		rule := SpecificDatesRule{
			Dates: []time.Time{
				time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			},
		}
		err := rule.Validate()
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestSpecificDatesRuleNext(t *testing.T) {
	t.Run("returns nearest next date", func(t *testing.T) {
		rule := SpecificDatesRule{
			Dates: []time.Time{
				time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
			},
		}

		from := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)

		got, err := rule.Next(from)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		want := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("skips equal date and returns strictly next", func(t *testing.T) {
		rule := SpecificDatesRule{
			Dates: []time.Time{
				time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
				time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
			},
		}

		from := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)

		got, err := rule.Next(from)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		want := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("returns error when no next date", func(t *testing.T) {
		rule := SpecificDatesRule{
			Dates: []time.Time{
				time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			},
		}

		from := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)

		_, err := rule.Next(from)
		if !errors.Is(err, ErrNoNextDate) {
			t.Fatalf("expected %v, got %v", ErrNoNextDate, err)
		}
	})
}

func TestEvenDaysRuleNext(t *testing.T) {
	tests := []struct {
		name string
		from time.Time
		want time.Time
	}{
		{
			name: "from odd day",
			from: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "from even day returns next even day",
			from: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "cross month boundary",
			from: time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC),
		},
	}

	rule := EvenDaysRule{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rule.Next(tt.from)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestOddDaysRuleNext(t *testing.T) {
	tests := []struct {
		name string
		from time.Time
		want time.Time
	}{
		{
			name: "from even day",
			from: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "from odd day returns next odd day",
			from: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "cross month boundary",
			from: time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
	}

	rule := OddDaysRule{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rule.Next(tt.from)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestSpecificDatesRule_Next_FromEqualsExistingDate(t *testing.T) {
	t.Parallel()

	date1 := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	date3 := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	rule := SpecificDatesRule{
		Dates: []time.Time{date1, date2, date3},
	}

	got, err := rule.Next(date2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got.Equal(date3) {
		t.Fatalf("got = %v, want %v", got, date3)
	}
}

func TestSpecificDatesRule_Next_NoNextDate(t *testing.T) {
	t.Parallel()

	date1 := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	rule := SpecificDatesRule{
		Dates: []time.Time{date1, date2},
	}

	_, err := rule.Next(date2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMonthlyDayRule_Next_YearBoundary(t *testing.T) {
	t.Parallel()

	rule := MonthlyDayRule{Day: 5}
	from := time.Date(2026, 12, 20, 12, 0, 0, 0, time.UTC)

	got, err := rule.Next(from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := time.Date(2027, 1, 5, 12, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got = %v, want %v", got, want)
	}
}

func TestEvenDaysRule_Next_YearBoundary(t *testing.T) {
	t.Parallel()

	rule := EvenDaysRule{}
	from := time.Date(2026, 12, 31, 9, 30, 0, 0, time.UTC)

	got, err := rule.Next(from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := time.Date(2027, 1, 2, 9, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got = %v, want %v", got, want)
	}
}

func TestOddDaysRule_Next_YearBoundary(t *testing.T) {
	t.Parallel()

	rule := OddDaysRule{}
	from := time.Date(2026, 12, 31, 9, 30, 0, 0, time.UTC)

	got, err := rule.Next(from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := time.Date(2027, 1, 1, 9, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got = %v, want %v", got, want)
	}
}
func TestSpecificDatesRuleNext_SkipsDatesLessThanOrEqualToFrom(t *testing.T) {
	from := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)

	rule := SpecificDatesRule{
		Dates: []time.Time{
			time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC),  // меньше from
			time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC), // равно from
			time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC), // первая подходящая
			time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		},
	}

	got, err := rule.Next(from)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	want := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestSpecificDatesRuleNext_ReturnsErrNoNextDateWhenAllDatesAreLessThanOrEqualToFrom(t *testing.T) {
	from := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	rule := SpecificDatesRule{
		Dates: []time.Time{
			time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC), // равно from
		},
	}

	got, err := rule.Next(from)
	if !errors.Is(err, ErrNoNextDate) {
		t.Fatalf("expected ErrNoNextDate, got %v", err)
	}
	if !got.IsZero() {
		t.Fatalf("expected zero time, got %v", got)
	}
}