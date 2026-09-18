package networklab

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInspectTLSNegotiatesHTTP2(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client := server.Client()
	client.Timeout = time.Second
	info, err := InspectTLS(context.Background(), client, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if info.HTTPProtocol != "HTTP/2.0" || info.ALPN != "h2" {
		t.Fatalf("unexpected negotiation: %#v", info)
	}
	if info.TLSVersion != "TLS 1.3" {
		t.Fatalf("TLS version = %q", info.TLSVersion)
	}
}

func TestInspectTLSRejectsPlainHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("plain"))
	}))
	defer server.Close()
	if _, err := InspectTLS(context.Background(), server.Client(), server.URL); err == nil {
		t.Fatal("expected missing TLS error")
	}
}

func TestTLSVersionName(t *testing.T) {
	for version, want := range map[uint16]string{0x0303: "TLS 1.2", 0x0304: "TLS 1.3", 0x9999: "0x9999"} {
		if got := tlsVersionName(version); got != want {
			t.Fatalf("tlsVersionName(%x) = %q, want %q", version, got, want)
		}
	}
}

func TestInspectTLSReportsRequestAndTransportErrors(t *testing.T) {
	if _, err := InspectTLS(context.Background(), http.DefaultClient, "://bad"); err == nil {
		t.Fatal("expected URL error")
	}

	want := errors.New("transport failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, want
	})}
	if _, err := InspectTLS(context.Background(), client, "https://example.test"); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func TestInspectTLSReportsBodyError(t *testing.T) {
	want := errors.New("body failed")
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       errorReadCloser{err: want},
			Request:    request,
		}, nil
	})}
	if _, err := InspectTLS(context.Background(), client, "https://example.test"); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}
