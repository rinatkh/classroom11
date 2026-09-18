package networklab

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestUDPExchangePreservesDatagrams(t *testing.T) {
	packetConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() { serverDone <- ServeUDPEcho(ctx, packetConn) }()

	requestCtx, requestCancel := context.WithTimeout(context.Background(), time.Second)
	want := []string{"first datagram", "second datagram"}
	got, err := UDPExchange(requestCtx, packetConn.LocalAddr().String(), want)
	requestCancel()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
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
		t.Fatal("UDP server did not stop")
	}
}

func TestUDPExchangeDialError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := UDPExchange(ctx, "bad-address", []string{"one"}); err == nil {
		t.Fatal("expected dial error")
	}
}
