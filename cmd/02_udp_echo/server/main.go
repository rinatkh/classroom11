package main

import (
	"context"
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
	address := flag.String("addr", "127.0.0.1:8082", "адрес UDP-сервера")
	flag.Parse()

	packetConn, err := net.ListenPacket("udp", *address)
	if err != nil {
		log.Fatalf("не удалось открыть UDP socket: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("UDP server слушает %s\n", packetConn.LocalAddr())
	fmt.Println("ReadFrom ждёт датаграмму; Ctrl+C завершает сервер")
	if err := networklab.ServeUDPEcho(ctx, packetConn); err != nil {
		log.Fatal(err)
	}
}
