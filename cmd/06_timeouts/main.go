package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(150 * time.Millisecond)
		_, _ = w.Write([]byte("too late"))
	}))
	defer server.Close()

	err := networklab.GetWithTimeout(context.Background(), server.Client(), server.URL, 40*time.Millisecond)
	fmt.Printf("Ошибка: %v\n", err)
	fmt.Printf("errors.Is(context deadline exceeded): %t\n", errors.Is(err, context.DeadlineExceeded))
	fmt.Println("Тайм-аут — часть корректности backend: бесконечно ждать нельзя.")
}
