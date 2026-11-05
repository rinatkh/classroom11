package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/rinatkh/classroom11/internal/httpapi"
	"github.com/rinatkh/classroom11/internal/logging"
	"github.com/rinatkh/classroom11/internal/networklab"
	"github.com/rinatkh/classroom11/internal/task"
)

func main() {
	if err := logging.Configure("network-lab-06_timeouts"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-06_timeouts stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "06_timeouts")
	timeout := flag.Duration("timeout", 40*time.Millisecond, "сколько клиент ждёт /debug/slow")
	delay := flag.Duration("delay", 150*time.Millisecond, "учебная задержка локального сервера")
	flag.Parse()

	// Это только инфраструктурная лаборатория. /debug/slow не входит
	// в router Task API, cmd/api или исходную точку следующего занятия.
	mux := http.NewServeMux()
	mux.Handle("/", httpapi.NewRouter(task.NewStore()))
	mux.HandleFunc("GET /debug/slow", func(w http.ResponseWriter, r *http.Request) {
		timer := time.NewTimer(*delay)
		defer timer.Stop()
		select {
		case <-timer.C:
			_, _ = w.Write([]byte("slow network lab response"))
		case <-r.Context().Done():
			// Клиент отменил ожидание; сервер прекращает учебную работу.
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := server.Client()
	client.Timeout = 2 * time.Second
	if err := networklab.GetWithTimeout(context.Background(), client, server.URL+"/health", time.Second); err != nil {
		return err
	}
	fmt.Println("Знакомый GET /health отвечает; Handler Task API не менялся.")
	fmt.Printf("Лабораторный GET /debug/slow: delay=%s, timeout=%s\n", *delay, *timeout)
	err := networklab.GetWithTimeout(context.Background(), client, server.URL+"/debug/slow", *timeout)
	fmt.Printf("Ошибка: %v\n", err)
	fmt.Printf("errors.Is(context deadline exceeded): %t\n", errors.Is(err, context.DeadlineExceeded))
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	fmt.Println("Тайм-аут — часть корректности backend: бесконечно ждать нельзя.")
	return nil
}
