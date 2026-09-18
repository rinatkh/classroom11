package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/rinatkh/classroom11/internal/networklab"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	transport := &http.Transport{MaxIdleConnsPerHost: 1}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	events, err := networklab.DoTracedRequests(context.Background(), client, server.URL, 2)
	if err != nil {
		log.Fatal(err)
	}
	for _, event := range events {
		fmt.Printf("request=%d %-12s %s\n", event.Request, event.Name, event.Detail)
	}
	fmt.Println("Во втором GotConn ожидаем reused=true: соединение вернулось в pool.")
}
