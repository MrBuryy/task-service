package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnknownRule           = errors.New("unknown recurrence rule")
	ErrInvalidRecurrenceConfig = errors.New("invalid recurrence config")
	ErrRuleTypeMismatch      = errors.New("rule type mismatch")
)

type DailyEveryNConfig struct {
	Every int `json:"every"`
}

type MonthlyDayConfig struct {
	Day int `json:"day"`
}

type SpecificDatesConfig struct {
	Dates []time.Time `json:"dates"`
}

type EmptyConfig struct{}

type RuleCodec interface {
	Type() RecurrenceType
	Matches(rule Rule) bool
	Decode(config json.RawMessage) (Rule, error)
	Encode(rule Rule) (json.RawMessage, error)
}

type RuleRegistry struct {
	byType []RuleCodec
}

func NewRuleRegistry(codecs ...RuleCodec) *RuleRegistry {
	r := &RuleRegistry{
		byType: make([]RuleCodec, 0, len(codecs)),
	}

	for _, codec := range codecs {
		r.Register(codec)
	}

	return r
}

func (r *RuleRegistry) Register(codec RuleCodec) {
	r.byType = append(r.byType, codec)
}

func (r *RuleRegistry) findByType(t RecurrenceType) (RuleCodec, bool) {
	for _, codec := range r.byType {
		if codec.Type() == t {
			return codec, true
		}
	}
	return nil, false
}

func (r *RuleRegistry) findByRule(rule Rule) (RuleCodec, bool) {
	for _, codec := range r.byType {
		if codec.Matches(rule) {
			return codec, true
		}
	}
	return nil, false
}

func (r *RuleRegistry) Decode(t RecurrenceType, config json.RawMessage) (Rule, error) {
	if t == RecurrenceNone {
		return nil, nil
	}

	codec, ok := r.findByType(t)
	if !ok {
		return nil, ErrUnknownRecurrenceType
	}

	rule, err := codec.Decode(config)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, nil
	}

	if err := rule.Validate(); err != nil {
		return nil, err
	}

	return rule, nil
}

func (r *RuleRegistry) Encode(rule Rule) (RecurrenceType, json.RawMessage, error) {
	if rule == nil {
		return RecurrenceNone, nil, nil
	}

	if err := rule.Validate(); err != nil {
		return "", nil, err
	}

	codec, ok := r.findByRule(rule)
	if !ok {
		return "", nil, ErrUnknownRule
	}

	cfg, err := codec.Encode(rule)
	if err != nil {
		return "", nil, err
	}

	return codec.Type(), cfg, nil
}

var defaultRuleRegistry = NewRuleRegistry(
	dailyEveryNCodec{},
	monthlyDayCodec{},
	specificDatesCodec{},
	evenDaysCodec{},
	oddDaysCodec{},
)

func DecodeRule(t RecurrenceType, config json.RawMessage) (Rule, error) {
	return defaultRuleRegistry.Decode(t, config)
}

func EncodeRule(rule Rule) (RecurrenceType, json.RawMessage, error) {
	return defaultRuleRegistry.Encode(rule)
}

type dailyEveryNCodec struct{}

func (c dailyEveryNCodec) Type() RecurrenceType {
	return RecurrenceDailyEveryN
}

func (c dailyEveryNCodec) Matches(rule Rule) bool {
	switch rule.(type) {
	case DailyEveryNRule, *DailyEveryNRule:
		return true
	default:
		return false
	}
}

func (c dailyEveryNCodec) Decode(config json.RawMessage) (Rule, error) {
	var cfg DailyEveryNConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return DailyEveryNRule{Every: cfg.Every}, nil
}

func (c dailyEveryNCodec) Encode(rule Rule) (json.RawMessage, error) {
	r, ok := asDailyEveryNRule(rule)
	if !ok {
		return nil, ErrRuleTypeMismatch
	}

	raw, err := json.Marshal(DailyEveryNConfig{
		Every: r.Every,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return raw, nil
}

type monthlyDayCodec struct{}

func (c monthlyDayCodec) Type() RecurrenceType {
	return RecurrenceMonthlyDay
}

func (c monthlyDayCodec) Matches(rule Rule) bool {
	switch rule.(type) {
	case MonthlyDayRule, *MonthlyDayRule:
		return true
	default:
		return false
	}
}

func (c monthlyDayCodec) Decode(config json.RawMessage) (Rule, error) {
	var cfg MonthlyDayConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return MonthlyDayRule{Day: cfg.Day}, nil
}

func (c monthlyDayCodec) Encode(rule Rule) (json.RawMessage, error) {
	r, ok := asMonthlyDayRule(rule)
	if !ok {
		return nil, ErrRuleTypeMismatch
	}

	raw, err := json.Marshal(MonthlyDayConfig{
		Day: r.Day,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return raw, nil
}

type specificDatesCodec struct{}

func (c specificDatesCodec) Type() RecurrenceType {
	return RecurrenceSpecificDates
}

func (c specificDatesCodec) Matches(rule Rule) bool {
	switch rule.(type) {
	case SpecificDatesRule, *SpecificDatesRule:
		return true
	default:
		return false
	}
}

func (c specificDatesCodec) Decode(config json.RawMessage) (Rule, error) {
	var cfg SpecificDatesConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return SpecificDatesRule{Dates: cfg.Dates}, nil
}

func (c specificDatesCodec) Encode(rule Rule) (json.RawMessage, error) {
	r, ok := asSpecificDatesRule(rule)
	if !ok {
		return nil, ErrRuleTypeMismatch
	}

	raw, err := json.Marshal(SpecificDatesConfig{
		Dates: r.Dates,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return raw, nil
}

type evenDaysCodec struct{}

func (c evenDaysCodec) Type() RecurrenceType {
	return RecurrenceEvenDays
}

func (c evenDaysCodec) Matches(rule Rule) bool {
	switch rule.(type) {
	case EvenDaysRule, *EvenDaysRule:
		return true
	default:
		return false
	}
}

func (c evenDaysCodec) Decode(config json.RawMessage) (Rule, error) {
	var cfg EmptyConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return EvenDaysRule{}, nil
}

func (c evenDaysCodec) Encode(rule Rule) (json.RawMessage, error) {
	if !c.Matches(rule) {
		return nil, ErrRuleTypeMismatch
	}

	raw, err := json.Marshal(EmptyConfig{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return raw, nil
}

type oddDaysCodec struct{}

func (c oddDaysCodec) Type() RecurrenceType {
	return RecurrenceOddDays
}

func (c oddDaysCodec) Matches(rule Rule) bool {
	switch rule.(type) {
	case OddDaysRule, *OddDaysRule:
		return true
	default:
		return false
	}
}

func (c oddDaysCodec) Decode(config json.RawMessage) (Rule, error) {
	var cfg EmptyConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return OddDaysRule{}, nil
}

func (c oddDaysCodec) Encode(rule Rule) (json.RawMessage, error) {
	if !c.Matches(rule) {
		return nil, ErrRuleTypeMismatch
	}

	raw, err := json.Marshal(EmptyConfig{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRecurrenceConfig, err)
	}

	return raw, nil
}

func unmarshalConfig(data json.RawMessage, dst any) error {
	if len(data) == 0 {
		data = json.RawMessage(`{}`)
	}
	return json.Unmarshal(data, dst)
}

func asDailyEveryNRule(rule Rule) (DailyEveryNRule, bool) {
	switch r := rule.(type) {
	case DailyEveryNRule:
		return r, true
	case *DailyEveryNRule:
		if r == nil {
			return DailyEveryNRule{}, false
		}
		return *r, true
	default:
		return DailyEveryNRule{}, false
	}
}

func asMonthlyDayRule(rule Rule) (MonthlyDayRule, bool) {
	switch r := rule.(type) {
	case MonthlyDayRule:
		return r, true
	case *MonthlyDayRule:
		if r == nil {
			return MonthlyDayRule{}, false
		}
		return *r, true
	default:
		return MonthlyDayRule{}, false
	}
}

func asSpecificDatesRule(rule Rule) (SpecificDatesRule, bool) {
	switch r := rule.(type) {
	case SpecificDatesRule:
		return r, true
	case *SpecificDatesRule:
		if r == nil {
			return SpecificDatesRule{}, false
		}
		return *r, true
	default:
		return SpecificDatesRule{}, false
	}
}