package postgres

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
	"reflect"

	taskdomain "example.com/taskservice/internal/domain/task"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestRepository(t *testing.T) (*Repository, *pgxpool.Pool) {
	t.Helper()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("pool.Ping() error = %v", err)
	}

	repo := New(pool)
	cleanupTasks(t, pool)

	return repo, pool
}

func cleanupTasks(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `TRUNCATE TABLE tasks RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("cleanup tasks error = %v", err)
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	return data
}

func TestRepository_Create(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	scheduledAt := now.Add(24 * time.Hour)

	input := &taskdomain.Task{
		Title:            "Pay rent",
		Description:      "Monthly rent payment",
		Status:           taskdomain.StatusNew,
		ScheduledAt:      scheduledAt,
		RecurrenceType:   taskdomain.RecurrenceMonthlyDay,
		RecurrenceConfig: mustJSON(t, map[string]any{"day": 10}),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("Create() ID = 0, want non-zero")
	}
	if created.Title != input.Title {
		t.Fatalf("Create() Title = %q, want %q", created.Title, input.Title)
	}
	if created.Description != input.Description {
		t.Fatalf("Create() Description = %q, want %q", created.Description, input.Description)
	}
	if created.Status != input.Status {
		t.Fatalf("Create() Status = %q, want %q", created.Status, input.Status)
	}
	if !created.ScheduledAt.Equal(input.ScheduledAt) {
		t.Fatalf("Create() ScheduledAt = %v, want %v", created.ScheduledAt, input.ScheduledAt)
	}
	if created.RecurrenceType != input.RecurrenceType {
		t.Fatalf("Create() RecurrenceType = %q, want %q", created.RecurrenceType, input.RecurrenceType)
	}
	assertJSONEqual(t, created.RecurrenceConfig, input.RecurrenceConfig)
	if created.CreatedAt.IsZero() {
		t.Fatalf("Create() CreatedAt is zero")
	}
	if created.UpdatedAt.IsZero() {
		t.Fatalf("Create() UpdatedAt is zero")
	}
}

func TestRepository_GetByID(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, &taskdomain.Task{
		Title:            "Doctor visit",
		Description:      "Annual checkup",
		Status:           taskdomain.StatusInProgress,
		ScheduledAt:      now.Add(48 * time.Hour),
		RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
		RecurrenceConfig: mustJSON(t, map[string]any{"every": 2}),
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("GetByID() ID = %d, want %d", got.ID, created.ID)
	}
	if got.Title != created.Title {
		t.Fatalf("GetByID() Title = %q, want %q", got.Title, created.Title)
	}
	if got.Description != created.Description {
		t.Fatalf("GetByID() Description = %q, want %q", got.Description, created.Description)
	}
	if got.Status != created.Status {
		t.Fatalf("GetByID() Status = %q, want %q", got.Status, created.Status)
	}
	if !got.ScheduledAt.Equal(created.ScheduledAt) {
		t.Fatalf("GetByID() ScheduledAt = %v, want %v", got.ScheduledAt, created.ScheduledAt)
	}
	if got.RecurrenceType != created.RecurrenceType {
		t.Fatalf("GetByID() RecurrenceType = %q, want %q", got.RecurrenceType, created.RecurrenceType)
	}
	if string(got.RecurrenceConfig) != string(created.RecurrenceConfig) {
		t.Fatalf("GetByID() RecurrenceConfig = %s, want %s", got.RecurrenceConfig, created.RecurrenceConfig)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999999)
	if err == nil {
		t.Fatal("GetByID() error = nil, want not found")
	}
	if err != taskdomain.ErrNotFound {
		t.Fatalf("GetByID() error = %v, want %v", err, taskdomain.ErrNotFound)
	}
}

func TestRepository_Update(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, &taskdomain.Task{
		Title:            "Initial title",
		Description:      "Initial description",
		Status:           taskdomain.StatusNew,
		ScheduledAt:      now.Add(24 * time.Hour),
		RecurrenceType:   taskdomain.RecurrenceNone,
		RecurrenceConfig: nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updatedAt := now.Add(2 * time.Hour)
	created.Title = "Updated title"
	created.Description = "Updated description"
	created.Status = taskdomain.StatusDone
	created.ScheduledAt = now.Add(72 * time.Hour)
	created.RecurrenceType = taskdomain.RecurrenceSpecificDates
	created.RecurrenceConfig = mustJSON(t, map[string]any{
		"dates": []string{
			now.Add(72 * time.Hour).Format(time.RFC3339),
			now.Add(96 * time.Hour).Format(time.RFC3339),
		},
	})
	created.UpdatedAt = updatedAt

	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.ID != created.ID {
		t.Fatalf("Update() ID = %d, want %d", updated.ID, created.ID)
	}
	if updated.Title != "Updated title" {
		t.Fatalf("Update() Title = %q, want %q", updated.Title, "Updated title")
	}
	if updated.Description != "Updated description" {
		t.Fatalf("Update() Description = %q, want %q", updated.Description, "Updated description")
	}
	if updated.Status != taskdomain.StatusDone {
		t.Fatalf("Update() Status = %q, want %q", updated.Status, taskdomain.StatusDone)
	}
	if !updated.ScheduledAt.Equal(created.ScheduledAt) {
		t.Fatalf("Update() ScheduledAt = %v, want %v", updated.ScheduledAt, created.ScheduledAt)
	}
	if updated.RecurrenceType != taskdomain.RecurrenceSpecificDates {
		t.Fatalf("Update() RecurrenceType = %q, want %q", updated.RecurrenceType, taskdomain.RecurrenceSpecificDates)
	}
	assertJSONEqual(t, updated.RecurrenceConfig, created.RecurrenceConfig)
	if !updated.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("Update() UpdatedAt = %v, want %v", updated.UpdatedAt, updatedAt)
	}
}

func TestRepository_Update_NotFound(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	task := &taskdomain.Task{
		ID:               999999,
		Title:            "Missing task",
		Description:      "Should not update",
		Status:           taskdomain.StatusNew,
		ScheduledAt:      now.Add(24 * time.Hour),
		RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
		RecurrenceConfig: mustJSON(t, map[string]any{"every": 1}),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	_, err := repo.Update(ctx, task)
	if err == nil {
		t.Fatal("Update() error = nil, want not found")
	}
	if err != taskdomain.ErrNotFound {
		t.Fatalf("Update() error = %v, want %v", err, taskdomain.ErrNotFound)
	}
}

func TestRepository_Delete(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, &taskdomain.Task{
		Title:            "Task to delete",
		Description:      "Delete me",
		Status:           taskdomain.StatusNew,
		ScheduledAt:      now.Add(24 * time.Hour),
		RecurrenceType:   taskdomain.RecurrenceEvenDays,
		RecurrenceConfig: mustJSON(t, map[string]any{}),
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(ctx, created.ID)
	if err == nil {
		t.Fatal("GetByID() after Delete() error = nil, want not found")
	}
	if err != taskdomain.ErrNotFound {
		t.Fatalf("GetByID() after Delete() error = %v, want %v", err, taskdomain.ErrNotFound)
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()

	err := repo.Delete(ctx, 999999)
	if err == nil {
		t.Fatal("Delete() error = nil, want not found")
	}
	if err != taskdomain.ErrNotFound {
		t.Fatalf("Delete() error = %v, want %v", err, taskdomain.ErrNotFound)
	}
}

func TestRepository_List(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	first, err := repo.Create(ctx, &taskdomain.Task{
		Title:            "First task",
		Description:      "Created first",
		Status:           taskdomain.StatusNew,
		ScheduledAt:      now.Add(24 * time.Hour),
		RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
		RecurrenceConfig: mustJSON(t, map[string]any{"every": 3}),
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}

	second, err := repo.Create(ctx, &taskdomain.Task{
		Title:            "Second task",
		Description:      "Created second",
		Status:           taskdomain.StatusInProgress,
		ScheduledAt:      now.Add(48 * time.Hour),
		RecurrenceType:   taskdomain.RecurrenceOddDays,
		RecurrenceConfig: mustJSON(t, map[string]any{}),
		CreatedAt:        now.Add(1 * time.Minute),
		UpdatedAt:        now.Add(1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("List() len = %d, want 2", len(got))
	}

	// repository uses ORDER BY id DESC
	if got[0].ID != second.ID {
		t.Fatalf("List()[0].ID = %d, want %d", got[0].ID, second.ID)
	}
	if got[1].ID != first.ID {
		t.Fatalf("List()[1].ID = %d, want %d", got[1].ID, first.ID)
	}

	if got[0].Title != second.Title {
		t.Fatalf("List()[0].Title = %q, want %q", got[0].Title, second.Title)
	}
	if got[1].Title != first.Title {
		t.Fatalf("List()[1].Title = %q, want %q", got[1].Title, first.Title)
	}

	if got[0].RecurrenceType != second.RecurrenceType {
		t.Fatalf("List()[0].RecurrenceType = %q, want %q", got[0].RecurrenceType, second.RecurrenceType)
	}
	if got[1].RecurrenceType != first.RecurrenceType {
		t.Fatalf("List()[1].RecurrenceType = %q, want %q", got[1].RecurrenceType, first.RecurrenceType)
	}

	if !got[0].ScheduledAt.Equal(second.ScheduledAt) {
		t.Fatalf("List()[0].ScheduledAt = %v, want %v", got[0].ScheduledAt, second.ScheduledAt)
	}
	if !got[1].ScheduledAt.Equal(first.ScheduledAt) {
		t.Fatalf("List()[1].ScheduledAt = %v, want %v", got[1].ScheduledAt, first.ScheduledAt)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	repo, pool := setupTestRepository(t)
	defer pool.Close()

	ctx := context.Background()

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("List() len = %d, want 0", len(got))
	}
}

func assertJSONEqual(t *testing.T, got, want json.RawMessage) {
	t.Helper()

	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("json.Unmarshal(got) error = %v", err)
	}

	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("json.Unmarshal(want) error = %v", err)
	}

	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("JSON not equal\n got:  %s\n want: %s", got, want)
	}
}
