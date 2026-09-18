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
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("secure"))
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client := server.Client()
	client.Timeout = 2 * time.Second
	info, err := networklab.InspectTLS(context.Background(), client, server.URL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("HTTP protocol: %s\nTLS version: %s\nALPN: %s\n", info.HTTPProtocol, info.TLSVersion, info.ALPN)
	fmt.Println("Локальный сертификат доверен только этому учебному client — в production так не делают.")
}
