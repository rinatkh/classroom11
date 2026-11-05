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
	if err := logging.Configure("network-lab-03_raw_http"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("network-lab-03_raw_http stopped", "error", err)
	}
}

func run() error {
	slog.Info("network experiment starting", "experiment", "03_raw_http")
	address := flag.String("addr", "127.0.0.1:8080", "адрес знакомого Task API; сначала запустите make run")
	path := flag.String("path", "/health", "path HTTP-запроса")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("В прошлом занятии: curl GET /health. Сейчас отправляем байты в тот же Task API:")
	fmt.Print(networklab.BuildRawGET(*address, *path))
	response, err := networklab.SendRawGET(ctx, *address, *address, *path)
	if err != nil {
		return err
	}
	fmt.Printf("\nСтатус: %s\nContent-Type: %s\nBody: %s\n", response.Status, response.Header.Get("Content-Type"), response.Body)
	return nil
}
