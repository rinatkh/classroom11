package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	address := flag.String("addr", "", "готовый HTTP-сервер; пустое значение запускает локальный")
	path := flag.String("path", "/health", "path HTTP-запроса")
	flag.Parse()

	target, stop := startDemoServer(*address)
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("Отправляем в TCP ровно такие байты:")
	fmt.Print(networklab.BuildRawGET(target, *path))
	response, err := networklab.SendRawGET(ctx, target, target, *path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nСтатус: %s\nContent-Type: %s\nBody: %s\n", response.Status, response.Header.Get("Content-Type"), response.Body)
}

func startDemoServer(address string) (string, func()) {
	if address != "" {
		return address, func() {}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("demo server: %v", err)
		}
	}()
	return listener.Addr().String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
}
