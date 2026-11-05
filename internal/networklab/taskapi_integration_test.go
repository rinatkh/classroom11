package networklab

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rinatkh/classroom11/internal/httpapi"
	"github.com/rinatkh/classroom11/internal/task"
)

// Обычный HTTP client создаёт Task; raw TCP client читает его из того же
// Store через неизменённые router и Handler. Сетевой опыт не подменяет API.
func TestRawHTTPReachesTaskAPI(t *testing.T) {
	server := httptest.NewServer(httpapi.NewRouter(task.NewStore()))
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second

	created, err := client.Post(server.URL+"/tasks", "application/json", strings.NewReader(`{"title":"Learn HTTP"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer created.Body.Close()
	if _, err := io.Copy(io.Discard, created.Body); err != nil {
		t.Fatal(err)
	}
	if created.StatusCode != http.StatusCreated || created.Header.Get("Location") != "/tasks/1" {
		t.Fatalf("POST: status=%s, Location=%q", created.Status, created.Header.Get("Location"))
	}

	address := strings.TrimPrefix(server.URL, "http://")
	for _, tt := range []struct {
		path   string
		status int
		body   string
	}{
		{path: "/health", status: http.StatusOK, body: `{"status":"ok"}`},
		{path: "/tasks", status: http.StatusOK, body: `[{"id":1,"title":"Learn HTTP","done":false}]`},
		{path: "/tasks/1", status: http.StatusOK, body: `{"id":1,"title":"Learn HTTP","done":false}`},
		{path: "/tasks/999", status: http.StatusNotFound, body: `{"error":"задача не найдена"}`},
	} {
		t.Run(tt.path, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			response, err := SendRawGET(ctx, address, address, tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tt.status || response.Body != tt.body {
				t.Fatalf("raw GET %s: %#v", tt.path, response)
			}
			if got := response.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type=%q", got)
			}
		})
	}
}
