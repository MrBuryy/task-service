package task

import (
	"errors"
	"slices"
	"time"
)

// RecurrenceType defines the type of recurrence rule.
type RecurrenceType string

const (
	RecurrenceNone          RecurrenceType = "none"
	RecurrenceDailyEveryN   RecurrenceType = "daily_every_n"
	RecurrenceMonthlyDay    RecurrenceType = "monthly_day"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenDays      RecurrenceType = "even_days"
	RecurrenceOddDays       RecurrenceType = "odd_days"
)

type Recurrence struct {
	Type          RecurrenceType
	EveryNDays    *int
	DayOfMonth    *int
	SpecificDates []time.Time
}

// Rule defines behavior for calculating the next occurrence.
type Rule interface {
	Validate() error
	Next(from time.Time) (time.Time, error)
}

var (
	ErrUnknownRecurrenceType = errors.New("unknown recurrence type")
	ErrInvalidEvery          = errors.New("every must be greater than 0")
	ErrInvalidDayOfMonth     = errors.New("day of month must be between 1 and 30")

	ErrNoSpecificDates     = errors.New("specific dates are required")
	ErrInvalidSpecificDate = errors.New("specific date is zero")
	ErrNoNextDate          = errors.New("no next date")
)

func (r Recurrence) Rule() (Rule, error) {
	switch r.Type {
	case RecurrenceNone:
		return nil, nil

	case RecurrenceDailyEveryN:
		if r.EveryNDays == nil {
			return nil, ErrInvalidEvery
		}
		return DailyEveryNRule{Every: *r.EveryNDays}, nil

	case RecurrenceMonthlyDay:
		if r.DayOfMonth == nil {
			return nil, ErrInvalidDayOfMonth
		}
		return MonthlyDayRule{Day: *r.DayOfMonth}, nil

	case RecurrenceSpecificDates:
		return SpecificDatesRule{Dates: r.SpecificDates}, nil

	case RecurrenceEvenDays:
		return EvenDaysRule{}, nil

	case RecurrenceOddDays:
		return OddDaysRule{}, nil

	default:
		return nil, ErrUnknownRecurrenceType
	}
}

type DailyEveryNRule struct {
	Every int
}

func (r DailyEveryNRule) Validate() error {
	if r.Every <= 0 {
		return ErrInvalidEvery
	}
	return nil
}

func (r DailyEveryNRule) Next(from time.Time) (time.Time, error) {
	if err := r.Validate(); err != nil {
		return time.Time{}, err
	}

	return from.AddDate(0, 0, r.Every), nil
}

type MonthlyDayRule struct {
	Day int
}

func (r MonthlyDayRule) Validate() error {
	if r.Day < 1 || r.Day > 30 {
		return ErrInvalidDayOfMonth
	}
	return nil
}

func (r MonthlyDayRule) Next(from time.Time) (time.Time, error) {
	if err := r.Validate(); err != nil {
		return time.Time{}, err
	}

	location := from.Location()
	year, month := from.Year(), from.Month()

	for i := 0; i < 12; i++ {
		candidate := time.Date(year, month, r.Day, from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), location)

		if candidate.Month() == month && candidate.Day() == r.Day && candidate.After(from) {
			return candidate, nil
		}

		next := time.Date(year, month, 1, from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), location).AddDate(0, 1, 0)
		year, month = next.Year(), next.Month()
	}

	return time.Time{}, ErrInvalidDayOfMonth
}

type SpecificDatesRule struct {
	Dates []time.Time
}

func (r SpecificDatesRule) Validate() error {
	if len(r.Dates) == 0 {
		return ErrNoSpecificDates
	}

	for _, d := range r.Dates {
		if d.IsZero() {
			return ErrInvalidSpecificDate
		}
	}

	return nil
}

func (r SpecificDatesRule) Next(from time.Time) (time.Time, error) {
	if err := r.Validate(); err != nil {
		return time.Time{}, err
	}

	dates := make([]time.Time, 0, len(r.Dates))
	for _, d := range r.Dates {
		dates = append(dates, d.UTC())
	}

	slices.SortFunc(dates, func(a, b time.Time) int {
		if a.Before(b) {
			return -1
		}
		if a.After(b) {
			return 1
		}
		return 0
	})

	from = from.UTC()

	for _, d := range dates {
		if d.After(from) {
			return d, nil
		}
	}

	return time.Time{}, ErrNoNextDate
}

type EvenDaysRule struct{}

func (r EvenDaysRule) Validate() error {
	return nil
}

func (r EvenDaysRule) Next(from time.Time) (time.Time, error) {
	next := from.AddDate(0, 0, 1)

	for {
		if next.Day()%2 == 0 {
			return next, nil
		}
		next = next.AddDate(0, 0, 1)
	}
}

type OddDaysRule struct{}

func (r OddDaysRule) Validate() error {
	return nil
}

func (r OddDaysRule) Next(from time.Time) (time.Time, error) {
	next := from.AddDate(0, 0, 1)

	for {
		if next.Day()%2 != 0 {
			return next, nil
		}
		next = next.AddDate(0, 0, 1)
	}
}