package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rinatkh/classroom11/internal/httpapi"
	"github.com/rinatkh/classroom11/internal/logging"
	"github.com/rinatkh/classroom11/internal/task"
)

func main() {
	if err := logging.Configure("task-api"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("task-api stopped", "error", err)
	}
}

func run() error {
	store := task.NewStore()
	router := httpapi.NewRouter(store)

	const address = ":8080"
	slog.Info("HTTP server starting", "addr", address)

	// ListenAndServe внутри открывает сетевой listener, принимает соединения
	// и передаёт HTTP-запросы router. Пакет net/http скрывает от нас ручной
	// цикл Accept и разбор байтов HTTP, поэтому в первом backend-уроке
	// сосредотачиваемся на маршрутах и handler-функциях.
	if err := http.ListenAndServe(address, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
