package networklab

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoTracedRequestsShowsReuse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("body must be consumed"))
	}))
	defer server.Close()

	transport := &http.Transport{MaxIdleConnsPerHost: 1}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	events, err := DoTracedRequests(context.Background(), client, server.URL, 2)
	if err != nil {
		t.Fatal(err)
	}

	var firstNew, secondReused, firstByte bool
	for _, event := range events {
		if event.Name == "GotConn" && event.Request == 1 && strings.Contains(event.Detail, "reused=false") {
			firstNew = true
		}
		if event.Name == "GotConn" && event.Request == 2 && strings.Contains(event.Detail, "reused=true") {
			secondReused = true
		}
		if event.Name == "FirstByte" {
			firstByte = true
		}
	}
	if !firstNew || !secondReused || !firstByte {
		t.Fatalf("unexpected trace events: %#v", events)
	}
}

func TestDoTracedRequestsRejectsBadURL(t *testing.T) {
	_, err := DoTracedRequests(context.Background(), http.DefaultClient, "://bad", 1)
	if err == nil {
		t.Fatal("expected request error")
	}
}

func TestDoTracedRequestsIncludesTLSHandshake(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	defer server.Close()
	events, err := DoTracedRequests(context.Background(), server.Client(), server.URL, 1)
	if err != nil {
		t.Fatal(err)
	}
	var started, completed bool
	for _, event := range events {
		started = started || event.Name == "TLSHandshakeStart"
		completed = completed || event.Name == "TLSHandshakeDone"
	}
	if !started || !completed {
		t.Fatalf("TLS events missing: %#v", events)
	}
}

func TestDoTracedRequestsReportsTransportAndBodyErrors(t *testing.T) {
	transportError := errors.New("transport failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportError
	})}
	if _, err := DoTracedRequests(context.Background(), client, "http://example.test", 1); !errors.Is(err, transportError) {
		t.Fatalf("error = %v", err)
	}

	bodyError := errors.New("body failed")
	client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       errorReadCloser{err: bodyError},
			Request:    request,
		}, nil
	})
	if _, err := DoTracedRequests(context.Background(), client, "http://example.test", 1); !errors.Is(err, bodyError) {
		t.Fatalf("error = %v", err)
	}

	closeError := errors.New("close failed")
	client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       closeErrorBody{err: closeError},
			Request:    request,
		}, nil
	})
	if _, err := DoTracedRequests(context.Background(), client, "http://example.test", 1); !errors.Is(err, closeError) {
		t.Fatalf("error = %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

type errorReadCloser struct{ err error }

func (body errorReadCloser) Read([]byte) (int, error) { return 0, body.err }
func (body errorReadCloser) Close() error             { return nil }

var _ io.ReadCloser = errorReadCloser{}

type closeErrorBody struct{ err error }

func (body closeErrorBody) Read([]byte) (int, error) { return 0, io.EOF }
func (body closeErrorBody) Close() error             { return body.err }
