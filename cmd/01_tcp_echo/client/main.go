package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/rinatkh/classroom11/internal/logging"
	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	if err := logging.Configure("network-lab-01_tcp_echo-client"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-01_tcp_echo-client stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "01_tcp_echo/client")
	address := flag.String("addr", "127.0.0.1:8081", "адрес TCP-сервера")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messages := []string{"первая строка", "вторая строка", "третья строка"}
	responses, err := networklab.TCPExchange(ctx, *address, messages)
	if err != nil {
		return err
	}
	for i, response := range responses {
		fmt.Printf("ответ %d: %s\n", i+1, response)
	}
	return nil
}
