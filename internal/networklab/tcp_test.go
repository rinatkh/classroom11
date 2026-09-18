package networklab

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestTCPExchange(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() { serverDone <- ServeTCPEcho(ctx, listener) }()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), time.Second)
	got, err := TCPExchange(requestCtx, listener.Addr().String(), []string{"one", "two\n"})
	requestCancel()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	want := []string{"echo: one", "echo: two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("responses = %#v, want %#v", got, want)
	}

	cancel()
	select {
	case err := <-serverDone:
		if err != nil {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("TCP server did not stop")
	}
}

func TestTCPExchangeDialError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := TCPExchange(ctx, "127.0.0.1:1", []string{"one"}); err == nil {
		t.Fatal("expected dial error")
	}
}
