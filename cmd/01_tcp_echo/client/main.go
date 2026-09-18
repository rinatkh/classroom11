package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	address := flag.String("addr", "127.0.0.1:8081", "адрес TCP-сервера")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	messages := []string{"первая строка", "вторая строка", "третья строка"}
	responses, err := networklab.TCPExchange(ctx, *address, messages)
	if err != nil {
		log.Fatal(err)
	}
	for i, response := range responses {
		fmt.Printf("ответ %d: %s\n", i+1, response)
	}
}
