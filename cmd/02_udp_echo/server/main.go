package main

import (
	"context"
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
	if err := logging.Configure("network-lab-02_udp_echo-server"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-02_udp_echo-server stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "02_udp_echo/server")
	address := flag.String("addr", "127.0.0.1:8082", "адрес UDP-сервера")
	readyFile := flag.String("ready-file", "", "файл с фактическим адресом для demo script")
	flag.Parse()

	packetConn, err := net.ListenPacket("udp", *address)
	if err != nil {
		return fmt.Errorf("не удалось открыть UDP socket: %v", err)
	}

	defer packetConn.Close()
	// Readiness — отдельный канал, не парсинг INFO/JSON логов.
	if *readyFile != "" {
		if err := os.WriteFile(*readyFile, []byte(packetConn.LocalAddr().String()), 0o600); err != nil {
			return fmt.Errorf("write lab readiness: %w", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("UDP socket ready", "addr", packetConn.LocalAddr().String())
	fmt.Println("ReadFrom ждёт датаграмму; Ctrl+C завершает сервер")
	if err := networklab.ServeUDPEcho(ctx, packetConn); err != nil {
		return err
	}
	return nil
}
