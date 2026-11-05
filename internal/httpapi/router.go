// Package httpapi связывает HTTP-маршруты с handler-функциями Task API.
package httpapi

import (
	"net/http"

	"github.com/rinatkh/classroom11/internal/task"
)

// Handler получает зависимости HTTP-слоя. На этом занятии зависимость одна —
// хранилище в памяти. Базу данных и сложные архитектурные слои пока не вводим.
type Handler struct {
	store *task.Store
}

// NewRouter регистрирует полный договор Task API.
//
// Router смотрит одновременно на HTTP method и path, а затем выбирает
// конкретную Go-функцию. Поэтому внутри handler не нужен большой if по Method.
func NewRouter(store *task.Store) http.Handler {
	h := &Handler{store: store}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /tasks", h.ListTasks)
	mux.HandleFunc("GET /tasks/{id}", h.GetTask)
	mux.HandleFunc("POST /tasks", h.CreateTask)
	mux.HandleFunc("PATCH /tasks/{id}", h.PatchTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)

	return mux
}

// Как не нужно: один handler с цепочкой if r.Method == ... быстро становится
// длинным и смешивает разные операции. Современный ServeMux разделяет их уже
// в момент регистрации маршрутов и сам вернёт 405 для неподходящего method.
