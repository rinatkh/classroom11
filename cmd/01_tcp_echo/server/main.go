package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rinatkh/classroom11/internal/logging"
	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	if err := logging.Configure("network-lab-01_tcp_echo-server"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-01_tcp_echo-server stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "01_tcp_echo/server")
	address := flag.String("addr", "127.0.0.1:8081", "адрес TCP-сервера")
	readyFile := flag.String("ready-file", "", "файл с фактическим адресом для demo script")
	flag.Parse()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		return fmt.Errorf("не удалось открыть TCP listener: %v", err)
	}

	defer listener.Close()
	// Readiness — отдельный канал, не парсинг INFO/JSON логов.
	if *readyFile != "" {
		if err := os.WriteFile(*readyFile, []byte(listener.Addr().String()), 0o600); err != nil {
			return fmt.Errorf("write lab readiness: %w", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("TCP listener ready", "addr", listener.Addr().String())
	fmt.Println("Accept ждёт подключение; Ctrl+C завершает сервер")
	if err := networklab.ServeTCPEcho(ctx, listener); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	return nil
}
