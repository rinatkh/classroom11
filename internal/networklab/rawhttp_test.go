package networklab

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBuildRawGET(t *testing.T) {
	got := BuildRawGET("example.test", "/health")
	want := "GET /health HTTP/1.1\r\nHost: example.test\r\nConnection: close\r\n\r\n"
	if got != want {
		t.Fatalf("request = %q, want %q", got, want)
	}
	if !strings.HasSuffix(BuildRawGET("example.test", ""), "\r\n\r\n") {
		t.Fatal("headers must end with an empty line")
	}
}

func TestSendRawGET(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != listener.Addr().String() {
			t.Errorf("Host = %q", r.Host)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := SendRawGET(ctx, listener.Addr().String(), listener.Addr().String(), "/health")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated || response.Status != "201 Created" {
		t.Fatalf("unexpected status: %#v", response)
	}
	if response.Header.Get("Content-Type") != "application/json" || response.Body != `{"status":"ok"}` {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestSendRawGETDialError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := SendRawGET(ctx, "127.0.0.1:1", "localhost", "/"); err == nil {
		t.Fatal("expected dial error")
	}
}

func TestSendRawGETReportsInvalidResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_, _ = conn.Write([]byte("not an HTTP response\r\n"))
			_ = conn.Close()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := SendRawGET(ctx, listener.Addr().String(), "localhost", "/"); err == nil {
		t.Fatal("expected response parsing error")
	}
}
