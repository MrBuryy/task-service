package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:            normalized.Title,
		Description:      normalized.Description,
		Status:           normalized.Status,
		ScheduledAt:      normalized.ScheduledAt,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: normalized.RecurrenceConfig,
	}

	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:               id,
		Title:            normalized.Title,
		Description:      normalized.Description,
		Status:           normalized.Status,
		ScheduledAt:      normalized.ScheduledAt,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: normalized.RecurrenceConfig,
		UpdatedAt:        s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	recType, recConfig, err := normalizeRecurrence(input.RecurrenceType, input.RecurrenceConfig)
	if err != nil {
		return CreateInput{}, err
	}

	input.RecurrenceType = recType
	input.RecurrenceConfig = recConfig

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	recType, recConfig, err := normalizeRecurrence(input.RecurrenceType, input.RecurrenceConfig)
	if err != nil {
		return UpdateInput{}, err
	}

	input.RecurrenceType = recType
	input.RecurrenceConfig = recConfig

	return input, nil
}

func normalizeRecurrence(
	recType taskdomain.RecurrenceType,
	recConfig json.RawMessage,
) (taskdomain.RecurrenceType, json.RawMessage, error) {
	rule, err := taskdomain.DecodeRule(recType, recConfig)
	if err != nil {
		return "", nil, fmt.Errorf("%w: invalid recurrence: %v", ErrInvalidInput, err)
	}

	if rule == nil {
		return taskdomain.RecurrenceNone, nil, nil
	}

	if err := rule.Validate(); err != nil {
		return "", nil, fmt.Errorf("%w: invalid recurrence: %v", ErrInvalidInput, err)
	}

	normalizedType, normalizedConfig, err := taskdomain.EncodeRule(rule)
	if err != nil {
		return "", nil, fmt.Errorf("%w: failed to encode recurrence: %v", ErrInvalidInput, err)
	}

	return normalizedType, normalizedConfig, nil
}

func (s *Service) Complete(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := s.now()

	if model.RecurrenceType == taskdomain.RecurrenceNone {
		model.Status = taskdomain.StatusDone
		model.UpdatedAt = now
		return s.repo.Update(ctx, model)
	}

	if model.ScheduledAt.IsZero() {
		return nil, fmt.Errorf("%w: scheduled_at is required for recurring task", ErrInvalidInput)
	}

	rule, err := taskdomain.DecodeRule(model.RecurrenceType, model.RecurrenceConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid recurrence: %v", ErrInvalidInput, err)
	}

	if rule == nil {
		model.Status = taskdomain.StatusDone
		model.UpdatedAt = now
		return s.repo.Update(ctx, model)
	}

	if err := rule.Validate(); err != nil {
		return nil, fmt.Errorf("%w: invalid recurrence: %v", ErrInvalidInput, err)
	}

	next, err := rule.Next(model.ScheduledAt)
	if err != nil {
		if errors.Is(err, taskdomain.ErrNoNextDate) {
			model.Status = taskdomain.StatusDone
			model.UpdatedAt = now
			return s.repo.Update(ctx, model)
		}

		return nil, fmt.Errorf("%w: failed to calculate next occurrence: %v", ErrInvalidInput, err)
	}

	model.Status = taskdomain.StatusNew
	model.ScheduledAt = next
	model.UpdatedAt = now

	return s.repo.Update(ctx, model)
}
