package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/rinatkh/classroom11/internal/logging"
	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	if err := logging.Configure("network-lab-04_http_trace"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-04_http_trace stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "04_http_trace")
	url := flag.String("url", "http://127.0.0.1:8080/health", "URL знакомого Task API; сначала запустите make run")
	flag.Parse()

	transport := &http.Transport{MaxIdleConnsPerHost: 1}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	fmt.Printf("Два GET %s: Handler тот же, наблюдаем connect, response и reuse.\n", *url)
	events, err := networklab.DoTracedRequests(context.Background(), client, *url, 2)
	if err != nil {
		return err
	}
	for _, event := range events {
		fmt.Printf("request=%d %-12s %s\n", event.Request, event.Name, event.Detail)
	}
	fmt.Println("Во втором GotConn ожидаем reused=true: соединение вернулось в pool.")
	return nil
}
