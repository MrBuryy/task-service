package task

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

type unknownRule struct{}

func (u unknownRule) Validate() error                  { return nil }
func (u unknownRule) Next(from time.Time) (time.Time, error) { return time.Time{}, nil }

func TestDecodeRule(t *testing.T) {
	date1 := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		rType   RecurrenceType
		config  json.RawMessage
		want    Rule
		wantErr error
	}{
		{
			name:   "none returns nil rule",
			rType:  RecurrenceNone,
			config: nil,
			want:   nil,
		},
		{
			name:   "decode daily every n",
			rType:  RecurrenceDailyEveryN,
			config: json.RawMessage(`{"every":2}`),
			want:   DailyEveryNRule{Every: 2},
		},
		{
			name:   "decode monthly day",
			rType:  RecurrenceMonthlyDay,
			config: json.RawMessage(`{"day":15}`),
			want:   MonthlyDayRule{Day: 15},
		},
		{
			name:   "decode specific dates",
			rType:  RecurrenceSpecificDates,
			config: mustMarshalJSON(t, SpecificDatesConfig{Dates: []time.Time{date1, date2}}),
			want:   SpecificDatesRule{Dates: []time.Time{date1, date2}},
		},
		{
			name:   "decode even days with empty config",
			rType:  RecurrenceEvenDays,
			config: json.RawMessage(`{}`),
			want:   EvenDaysRule{},
		},
		{
			name:   "decode odd days with nil config",
			rType:  RecurrenceOddDays,
			config: nil,
			want:   OddDaysRule{},
		},
		{
			name:    "unknown recurrence type",
			rType:   RecurrenceType("unknown"),
			config:  json.RawMessage(`{}`),
			wantErr: ErrUnknownRecurrenceType,
		},
		{
			name:    "invalid json for daily",
			rType:   RecurrenceDailyEveryN,
			config:  json.RawMessage(`{"every":`),
			wantErr: ErrInvalidRecurrenceConfig,
		},
		{
			name:    "invalid daily config value",
			rType:   RecurrenceDailyEveryN,
			config:  json.RawMessage(`{"every":0}`),
			wantErr: ErrInvalidEvery,
		},
		{
			name:    "invalid monthly day value",
			rType:   RecurrenceMonthlyDay,
			config:  json.RawMessage(`{"day":31}`),
			wantErr: ErrInvalidDayOfMonth,
		},
		{
			name:    "empty specific dates",
			rType:   RecurrenceSpecificDates,
			config:  json.RawMessage(`{"dates":[]}`),
			wantErr: ErrNoSpecificDates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeRule(tt.rType, tt.config)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected rule %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestEncodeRule(t *testing.T) {
	date1 := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		rule       Rule
		wantType   RecurrenceType
		wantConfig json.RawMessage
		wantErr    error
	}{
		{
			name:       "nil rule",
			rule:       nil,
			wantType:   RecurrenceNone,
			wantConfig: nil,
		},
		{
			name:       "encode daily every n",
			rule:       DailyEveryNRule{Every: 2},
			wantType:   RecurrenceDailyEveryN,
			wantConfig: json.RawMessage(`{"every":2}`),
		},
		{
			name:       "encode monthly day",
			rule:       MonthlyDayRule{Day: 15},
			wantType:   RecurrenceMonthlyDay,
			wantConfig: json.RawMessage(`{"day":15}`),
		},
		{
			name:       "encode specific dates",
			rule:       SpecificDatesRule{Dates: []time.Time{date1, date2}},
			wantType:   RecurrenceSpecificDates,
			wantConfig: mustMarshalJSON(t, SpecificDatesConfig{Dates: []time.Time{date1, date2}}),
		},
		{
			name:       "encode even days",
			rule:       EvenDaysRule{},
			wantType:   RecurrenceEvenDays,
			wantConfig: json.RawMessage(`{}`),
		},
		{
			name:       "encode odd days",
			rule:       OddDaysRule{},
			wantType:   RecurrenceOddDays,
			wantConfig: json.RawMessage(`{}`),
		},
		{
			name:    "invalid daily rule",
			rule:    DailyEveryNRule{Every: 0},
			wantErr: ErrInvalidEvery,
		},
		{
			name:    "unknown rule",
			rule:    unknownRule{},
			wantErr: ErrUnknownRule,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotConfig, err := EncodeRule(tt.rule)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if err != nil {
				return
			}

			if gotType != tt.wantType {
				t.Fatalf("expected recurrence type %q, got %q", tt.wantType, gotType)
			}

			if !jsonEqual(gotConfig, tt.wantConfig) {
				t.Fatalf("expected config %s, got %s", string(tt.wantConfig), string(gotConfig))
			}
		})
	}
}

func TestEncodeDecodeRule_RoundTrip(t *testing.T) {
	date1 := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		rule Rule
	}{
		{
			name: "daily every n",
			rule: DailyEveryNRule{Every: 3},
		},
		{
			name: "monthly day",
			rule: MonthlyDayRule{Day: 12},
		},
		{
			name: "specific dates",
			rule: SpecificDatesRule{Dates: []time.Time{date1, date2}},
		},
		{
			name: "even days",
			rule: EvenDaysRule{},
		},
		{
			name: "odd days",
			rule: OddDaysRule{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rType, config, err := EncodeRule(tt.rule)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}

			decoded, err := DecodeRule(rType, config)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if !reflect.DeepEqual(decoded, tt.rule) {
				t.Fatalf("expected decoded rule %#v, got %#v", tt.rule, decoded)
			}
		})
	}
}

func mustMarshalJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}

	return data
}

func jsonEqual(a, b []byte) bool {
	var va any
	var vb any

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	if err := json.Unmarshal(a, &va); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		return false
	}

	return reflect.DeepEqual(va, vb)
}