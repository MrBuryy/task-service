package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type mockUsecase struct {
	createFn   func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error)
	getByIDFn  func(ctx context.Context, id int64) (*taskdomain.Task, error)
	updateFn   func(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error)
	deleteFn   func(ctx context.Context, id int64) error
	listFn     func(ctx context.Context) ([]taskdomain.Task, error)
	completeFn func(ctx context.Context, id int64) (*taskdomain.Task, error)
}

func (m *mockUsecase) Create(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockUsecase) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUsecase) Update(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, input)
	}
	return nil, nil
}

func (m *mockUsecase) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockUsecase) List(ctx context.Context) ([]taskdomain.Task, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockUsecase) Complete(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if m.completeFn != nil {
		return m.completeFn(ctx, id)
	}
	return nil, nil
}

func TestTaskHandler_Create(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	t.Run("success recurring task", func(t *testing.T) {
		var gotInput taskusecase.CreateInput

		h := NewTaskHandler(&mockUsecase{
			createFn: func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
				gotInput = input
				return &taskdomain.Task{
					ID:               1,
					Title:            input.Title,
					Description:      input.Description,
					Status:           input.Status,
					ScheduledAt:      input.ScheduledAt,
					RecurrenceType:   input.RecurrenceType,
					RecurrenceConfig: input.RecurrenceConfig,
					CreatedAt:        now,
					UpdatedAt:        now,
				}, nil
			},
		})

		body := `{
			"title":"Pay rent",
			"description":"monthly payment",
			"status":"new",
			"scheduled_at":"2026-05-01T10:00:00Z",
			"recurrence":{
				"type":"monthly_day",
				"config":{"day":1}
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		if gotInput.Title != "Pay rent" {
			t.Fatalf("title = %q, want %q", gotInput.Title, "Pay rent")
		}

		if gotInput.RecurrenceType != taskdomain.RecurrenceMonthlyDay {
			t.Fatalf("recurrence type = %q, want %q", gotInput.RecurrenceType, taskdomain.RecurrenceMonthlyDay)
		}

		if string(gotInput.RecurrenceConfig) != `{"day":1}` {
			t.Fatalf("recurrence config = %s, want %s", string(gotInput.RecurrenceConfig), `{"day":1}`)
		}

		var resp taskDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp.ID != 1 {
			t.Fatalf("id = %d, want %d", resp.ID, 1)
		}
		if resp.Recurrence.Type != taskdomain.RecurrenceMonthlyDay {
			t.Fatalf("response recurrence type = %q, want %q", resp.Recurrence.Type, taskdomain.RecurrenceMonthlyDay)
		}
	})

	t.Run("success without recurrence defaults to none", func(t *testing.T) {
		var gotInput taskusecase.CreateInput

		h := NewTaskHandler(&mockUsecase{
			createFn: func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
				gotInput = input
				return &taskdomain.Task{
					ID:               2,
					Title:            input.Title,
					Description:      input.Description,
					Status:           input.Status,
					ScheduledAt:      input.ScheduledAt,
					RecurrenceType:   input.RecurrenceType,
					RecurrenceConfig: input.RecurrenceConfig,
					CreatedAt:        now,
					UpdatedAt:        now,
				}, nil
			},
		})

		body := `{
			"title":"One time task",
			"description":"do it once",
			"status":"new",
			"scheduled_at":"2026-05-10T08:00:00Z"
		}`

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}

		if gotInput.RecurrenceType != taskdomain.RecurrenceNone {
			t.Fatalf("recurrence type = %q, want %q", gotInput.RecurrenceType, taskdomain.RecurrenceNone)
		}
		if gotInput.RecurrenceConfig != nil {
			t.Fatalf("recurrence config = %v, want nil", gotInput.RecurrenceConfig)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"title":`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		body := `{
			"title":"Task",
			"description":"Desc",
			"status":"new",
			"scheduled_at":"2026-05-10T08:00:00Z",
			"unexpected":"value"
		}`

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid input from usecase", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			createFn: func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
				return nil, taskusecase.ErrInvalidInput
			},
		})

		body := `{
			"title":"",
			"description":"Desc",
			"status":"new",
			"scheduled_at":"2026-05-10T08:00:00Z"
		}`

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			createFn: func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
				return nil, errors.New("db down")
			},
		})

		body := `{
			"title":"Task",
			"description":"Desc",
			"status":"new",
			"scheduled_at":"2026-05-10T08:00:00Z"
		}`

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestTaskHandler_GetByID(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			getByIDFn: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
				if id != 42 {
					t.Fatalf("id = %d, want %d", id, 42)
				}
				return &taskdomain.Task{
					ID:             42,
					Title:          "Task",
					Description:    "Desc",
					Status:         taskdomain.StatusNew,
					ScheduledAt:    now,
					RecurrenceType: taskdomain.RecurrenceNone,
					CreatedAt:      now,
					UpdatedAt:      now,
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/tasks/42", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "42"})
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			getByIDFn: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
				return nil, taskdomain.ErrNotFound
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/tasks/42", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "42"})
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestTaskHandler_Update(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		var gotID int64
		var gotInput taskusecase.UpdateInput

		h := NewTaskHandler(&mockUsecase{
			updateFn: func(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
				gotID = id
				gotInput = input

				return &taskdomain.Task{
					ID:               id,
					Title:            input.Title,
					Description:      input.Description,
					Status:           input.Status,
					ScheduledAt:      input.ScheduledAt,
					RecurrenceType:   input.RecurrenceType,
					RecurrenceConfig: input.RecurrenceConfig,
					CreatedAt:        now,
					UpdatedAt:        now,
				}, nil
			},
		})

		body := `{
			"title":"Updated title",
			"description":"Updated desc",
			"status":"in_progress",
			"scheduled_at":"2026-06-01T09:00:00Z",
			"recurrence":{
				"type":"daily_every_n",
				"config":{"every":2}
			}
		}`

		req := httptest.NewRequest(http.MethodPut, "/tasks/7", bytes.NewBufferString(body))
		req = mux.SetURLVars(req, map[string]string{"id": "7"})
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		if gotID != 7 {
			t.Fatalf("id = %d, want %d", gotID, 7)
		}
		if gotInput.RecurrenceType != taskdomain.RecurrenceDailyEveryN {
			t.Fatalf("recurrence type = %q, want %q", gotInput.RecurrenceType, taskdomain.RecurrenceDailyEveryN)
		}
		if string(gotInput.RecurrenceConfig) != `{"every":2}` {
			t.Fatalf("recurrence config = %s, want %s", string(gotInput.RecurrenceConfig), `{"every":2}`)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		req := httptest.NewRequest(http.MethodPut, "/tasks/0", bytes.NewBufferString(`{}`))
		req = mux.SetURLVars(req, map[string]string{"id": "0"})
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		req := httptest.NewRequest(http.MethodPut, "/tasks/7", bytes.NewBufferString(`{"title":`))
		req = mux.SetURLVars(req, map[string]string{"id": "7"})
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			updateFn: func(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
				return nil, taskdomain.ErrNotFound
			},
		})

		body := `{
			"title":"Updated title",
			"description":"Updated desc",
			"status":"new",
			"scheduled_at":"2026-06-01T09:00:00Z"
		}`

		req := httptest.NewRequest(http.MethodPut, "/tasks/7", bytes.NewBufferString(body))
		req = mux.SetURLVars(req, map[string]string{"id": "7"})
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestTaskHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotID int64

		h := NewTaskHandler(&mockUsecase{
			deleteFn: func(ctx context.Context, id int64) error {
				gotID = id
				return nil
			},
		})

		req := httptest.NewRequest(http.MethodDelete, "/tasks/9", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "9"})
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if gotID != 9 {
			t.Fatalf("id = %d, want %d", gotID, 9)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		req := httptest.NewRequest(http.MethodDelete, "/tasks/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			deleteFn: func(ctx context.Context, id int64) error {
				return taskdomain.ErrNotFound
			},
		})

		req := httptest.NewRequest(http.MethodDelete, "/tasks/9", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "9"})
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestTaskHandler_List(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			listFn: func(ctx context.Context) ([]taskdomain.Task, error) {
				return []taskdomain.Task{
					{
						ID:             1,
						Title:          "Task 1",
						Description:    "Desc 1",
						Status:         taskdomain.StatusNew,
						ScheduledAt:    now,
						RecurrenceType: taskdomain.RecurrenceNone,
						CreatedAt:      now,
						UpdatedAt:      now,
					},
					{
						ID:               2,
						Title:            "Task 2",
						Description:      "Desc 2",
						Status:           taskdomain.StatusInProgress,
						ScheduledAt:      now.Add(24 * time.Hour),
						RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
						RecurrenceConfig: json.RawMessage(`{"every":2}`),
						CreatedAt:        now,
						UpdatedAt:        now,
					},
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp []taskDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if len(resp) != 2 {
			t.Fatalf("len = %d, want %d", len(resp), 2)
		}
		if resp[1].Recurrence.Type != taskdomain.RecurrenceDailyEveryN {
			t.Fatalf("recurrence type = %q, want %q", resp[1].Recurrence.Type, taskdomain.RecurrenceDailyEveryN)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			listFn: func(ctx context.Context) ([]taskdomain.Task, error) {
				return nil, errors.New("db down")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestTaskHandler_Complete(t *testing.T) {
	now := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	t.Run("success one-time task", func(t *testing.T) {
		var gotID int64

		h := NewTaskHandler(&mockUsecase{
			completeFn: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
				gotID = id
				return &taskdomain.Task{
					ID:             id,
					Title:          "One time",
					Description:    "done",
					Status:         taskdomain.StatusDone,
					ScheduledAt:    now,
					RecurrenceType: taskdomain.RecurrenceNone,
					CreatedAt:      now,
					UpdatedAt:      now,
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/tasks/5/complete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "5"})
		rec := httptest.NewRecorder()

		h.Complete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if gotID != 5 {
			t.Fatalf("id = %d, want %d", gotID, 5)
		}
	})

	t.Run("success recurring task", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			completeFn: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
				return &taskdomain.Task{
					ID:               id,
					Title:            "Recurring",
					Description:      "next occurrence calculated",
					Status:           taskdomain.StatusNew,
					ScheduledAt:      now.Add(48 * time.Hour),
					RecurrenceType:   taskdomain.RecurrenceDailyEveryN,
					RecurrenceConfig: json.RawMessage(`{"every":2}`),
					CreatedAt:        now,
					UpdatedAt:        now,
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/tasks/6/complete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "6"})
		rec := httptest.NewRecorder()

		h.Complete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp taskDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp.Recurrence.Type != taskdomain.RecurrenceDailyEveryN {
			t.Fatalf("recurrence type = %q, want %q", resp.Recurrence.Type, taskdomain.RecurrenceDailyEveryN)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{})

		req := httptest.NewRequest(http.MethodPost, "/tasks/abc/complete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rec := httptest.NewRecorder()

		h.Complete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewTaskHandler(&mockUsecase{
			completeFn: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
				return nil, taskdomain.ErrNotFound
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/tasks/5/complete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "5"})
		rec := httptest.NewRecorder()

		h.Complete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}
