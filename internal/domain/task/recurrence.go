package task

import (
	"errors"
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

// Rule defines behavior for calculating the next occurrence.
type Rule interface {
	Next(from time.Time) (time.Time, error)
}

var ErrUnknownRecurrenceType = errors.New("unknown recurrence type")