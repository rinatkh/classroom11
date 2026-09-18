package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	address := flag.String("addr", "127.0.0.1:8081", "адрес TCP-сервера")
	flag.Parse()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("не удалось открыть TCP listener: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("TCP server слушает %s\n", listener.Addr())
	fmt.Println("Accept ждёт подключение; Ctrl+C завершает сервер")
	if err := networklab.ServeTCPEcho(ctx, listener); err != nil && !errors.Is(err, net.ErrClosed) {
		log.Fatal(err)
	}
}
