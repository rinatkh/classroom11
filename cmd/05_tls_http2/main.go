package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"time"

	"github.com/rinatkh/classroom11/internal/httpapi"
	"github.com/rinatkh/classroom11/internal/logging"
	"github.com/rinatkh/classroom11/internal/networklab"
	"github.com/rinatkh/classroom11/internal/task"
)

func main() {
	if err := logging.Configure("network-lab-05_tls_http2"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-05_tls_http2 stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "05_tls_http2")
	// Отдельный локальный экземпляр того же Task API. Business code не меняем:
	// тестовый listener добавляет TLS и согласование HTTP/2 до знакомого Handler.
	server := httptest.NewUnstartedServer(httpapi.NewRouter(task.NewStore()))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client := server.Client()
	client.Timeout = 2 * time.Second
	fmt.Printf("GET %s/health → тот же Task API Handler через TLS\n", server.URL)
	info, err := networklab.InspectTLS(context.Background(), client, server.URL+"/health")
	if err != nil {
		return err
	}
	fmt.Printf("HTTP protocol: %s\nTLS version: %s\nALPN: %s\n", info.HTTPProtocol, info.TLSVersion, info.ALPN)
	fmt.Println("Локальный сертификат доверен только этому учебному client — в production так не делают.")
	return nil
}
