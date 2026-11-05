package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rinatkh/classroom11/internal/task"
)

func TestHealth(t *testing.T) {
	router := NewRouter(task.NewStore())
	response := performRequest(router, http.MethodGet, "/health", "")

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)
	if strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("неожиданный body: %s", response.Body.String())
	}
}

func TestGetEmptyList(t *testing.T) {
	router := NewRouter(task.NewStore())
	response := performRequest(router, http.MethodGet, "/tasks", "")

	assertStatus(t, response, http.StatusOK)
	if strings.TrimSpace(response.Body.String()) != `[]` {
		t.Fatalf("пустой список должен быть [], получили %s", response.Body.String())
	}
}

func TestPostTask(t *testing.T) {
	router := NewRouter(task.NewStore())
	response := performRequest(router, http.MethodPost, "/tasks", `{"title":"Learn HTTP"}`)

	assertStatus(t, response, http.StatusCreated)
	assertJSONContentType(t, response)
	if location := response.Header().Get("Location"); location != "/tasks/1" {
		t.Fatalf("ожидали Location /tasks/1, получили %q", location)
	}

	var item task.Task
	if err := json.NewDecoder(response.Body).Decode(&item); err != nil {
		t.Fatalf("ответ не является Task JSON: %v", err)
	}
	if item.ID != 1 || item.Title != "Learn HTTP" || item.Done {
		t.Fatalf("неожиданная задача: %#v", item)
	}
}

func TestPostValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "сломанный JSON", body: `{"title":`},
		{name: "пустой title", body: `{"title":"   "}`},
		{name: "неизвестное поле", body: `{"titel":"опечатка"}`},
		{name: "два JSON-объекта", body: `{"title":"a"}{"title":"b"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(task.NewStore())
			response := performRequest(router, http.MethodPost, "/tasks", tt.body)
			assertStatus(t, response, http.StatusBadRequest)
			assertJSONContentType(t, response)
		})
	}
}

func TestGetTask(t *testing.T) {
	store := task.NewStore()
	created, err := store.Create("Прочитать Request")
	if err != nil {
		t.Fatalf("подготовка теста: %v", err)
	}
	router := NewRouter(store)

	response := performRequest(router, http.MethodGet, "/tasks/1", "")
	assertStatus(t, response, http.StatusOK)
	if !strings.Contains(response.Body.String(), created.Title) {
		t.Fatalf("ответ не содержит созданную задачу: %s", response.Body.String())
	}
}

func TestGetTaskErrors(t *testing.T) {
	router := NewRouter(task.NewStore())

	tests := []struct {
		path   string
		status int
	}{
		{path: "/tasks/not-a-number", status: http.StatusBadRequest},
		{path: "/tasks/0", status: http.StatusBadRequest},
		{path: "/tasks/999", status: http.StatusNotFound},
	}
	for _, tt := range tests {
		response := performRequest(router, http.MethodGet, tt.path, "")
		assertStatus(t, response, tt.status)
	}
}

func TestPatchTask(t *testing.T) {
	store := task.NewStore()
	if _, err := store.Create("Черновик"); err != nil {
		t.Fatalf("подготовка теста: %v", err)
	}
	router := NewRouter(store)

	response := performRequest(router, http.MethodPatch, "/tasks/1", `{"done":true}`)
	assertStatus(t, response, http.StatusOK)

	var item task.Task
	if err := json.NewDecoder(response.Body).Decode(&item); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !item.Done || item.Title != "Черновик" {
		t.Fatalf("PATCH должен изменить только done: %#v", item)
	}
}

func TestPatchTaskErrors(t *testing.T) {
	store := task.NewStore()
	if _, err := store.Create("Черновик"); err != nil {
		t.Fatalf("подготовка теста: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{name: "ID не число", path: "/tasks/abc", body: `{"done":true}`, status: http.StatusBadRequest},
		{name: "объект не найден", path: "/tasks/999", body: `{"done":true}`, status: http.StatusNotFound},
		{name: "сломанный JSON", path: "/tasks/1", body: `{"done":`, status: http.StatusBadRequest},
		{name: "нет изменений", path: "/tasks/1", body: `{}`, status: http.StatusBadRequest},
		{name: "пустой title", path: "/tasks/1", body: `{"title":"   "}`, status: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Как и остальные endpoint-тесты, проходим через ServeMux:
			// именно router заполняет r.PathValue("id") до вызова Handler.
			router := NewRouter(store)
			response := performRequest(router, http.MethodPatch, tt.path, tt.body)

			assertStatus(t, response, tt.status)
			assertJSONContentType(t, response)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	store := task.NewStore()
	if _, err := store.Create("Удалить"); err != nil {
		t.Fatalf("подготовка теста: %v", err)
	}
	router := NewRouter(store)

	response := performRequest(router, http.MethodDelete, "/tasks/1", "")
	assertStatus(t, response, http.StatusNoContent)
	if response.Body.Len() != 0 {
		t.Fatalf("204 не должен содержать body: %q", response.Body.String())
	}

	response = performRequest(router, http.MethodGet, "/tasks/1", "")
	assertStatus(t, response, http.StatusNotFound)
}

func TestDeleteTaskErrors(t *testing.T) {
	router := NewRouter(task.NewStore())

	tests := []struct {
		path   string
		status int
	}{
		{path: "/tasks/abc", status: http.StatusBadRequest},
		{path: "/tasks/999", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		response := performRequest(router, http.MethodDelete, tt.path, "")
		assertStatus(t, response, tt.status)
		assertJSONContentType(t, response)
	}
}

func TestWrongMethod(t *testing.T) {
	router := NewRouter(task.NewStore())
	response := performRequest(router, http.MethodPut, "/tasks", `{}`)
	assertStatus(t, response, http.StatusMethodNotAllowed)
	if allow := response.Header().Get("Allow"); allow == "" {
		t.Fatal("ServeMux должен добавить Allow для 405")
	}
}

func performRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status: получили %d, ожидали %d; body=%s", response.Code, want, response.Body.String())
	}
}

func assertJSONContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type: получили %q", got)
	}
}
