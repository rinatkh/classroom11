package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rinatkh/classroom11/internal/task"
)

type healthResponse struct {
	Status string `json:"status"`
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type patchTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

// Health обрабатывает GET /health.
//
// Это самый маленький законченный endpoint: router выбирает Handler,
// Handler формирует Go-структуру, Encoder превращает её в JSON, клиент
// получает status 200 и body {"status":"ok"}.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "task.health")
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// ListTasks обрабатывает GET /tasks.
//
// Handler получает список из хранилища и возвращает JSON-массив. Пустой
// список выглядит как [], а не null: клиенту проще работать с одним типом.
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "task.listtasks")
	writeJSON(w, http.StatusOK, h.store.List())
}

// GetTask обрабатывает GET /tasks/{id}.
//
// Значение {id} извлекается через r.PathValue, затем превращается в int.
// Ошибка формата ID — 400, отсутствие корректного ID в store — 404.
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "task.gettask")
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	item, err := h.store.Get(id)
	if errors.Is(err, task.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// CreateTask обрабатывает POST /tasks — центральный путь занятия.
//
// Шаги: прочитать JSON body → проверить title → сохранить Task → поставить
// Location → вернуть status 201 и созданный объект как JSON.
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "task.createtask")
	var input createTaskRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return
	}

	item, err := h.store.Create(input.Title)
	if errors.Is(err, task.ErrEmptyTitle) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/tasks/%d", item.ID))
	slog.InfoContext(r.Context(), "task created", "task_id", item.ID)
	writeJSON(w, http.StatusCreated, item)
}

// PatchTask обрабатывает PATCH /tasks/{id}.
//
// PATCH меняет только присланные поля. Указатель на bool позволяет отличить
// отсутствующее поле done от явно присланного false.
func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "task.patchtask")
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var input patchTaskRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return
	}

	item, err := h.store.Patch(id, task.Changes{Title: input.Title, Done: input.Done})
	switch {
	case errors.Is(err, task.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, task.ErrEmptyTitle), errors.Is(err, task.ErrNoChanges):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		slog.InfoContext(r.Context(), "task updated", "task_id", item.ID)
		writeJSON(w, http.StatusOK, item)
	}
}

// DeleteTask обрабатывает DELETE /tasks/{id}.
//
// После успешного удаления возвращаем 204 No Content и намеренно не пишем
// body. Даже пустой JSON-объект здесь нарушил бы смысл выбранного статуса.
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "HTTP request", "operation", "task.deletetask")
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.store.Delete(id); errors.Is(err, task.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	slog.InfoContext(r.Context(), "task deleted", "task_id", id)
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	rawID := r.PathValue("id")
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id должен быть положительным целым числом")
		return 0, false
	}
	return id, true
}
