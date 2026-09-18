package networklab

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case <-time.After(100 * time.Millisecond):
			_, _ = w.Write([]byte("late"))
		case <-time.After(time.Second):
		}
	}))
	defer server.Close()

	err := GetWithTimeout(context.Background(), server.Client(), server.URL, 10*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
}

func TestGetWithTimeoutSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()
	if err := GetWithTimeout(context.Background(), server.Client(), server.URL, time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestGetWithTimeoutRejectsBadURL(t *testing.T) {
	if err := GetWithTimeout(context.Background(), http.DefaultClient, "://bad", time.Second); err == nil {
		t.Fatal("expected URL error")
	}
}

func TestGetWithTimeoutReportsBodyError(t *testing.T) {
	want := errors.New("body failed")
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       errorReadCloser{err: want},
			Request:    request,
		}, nil
	})}
	err := GetWithTimeout(context.Background(), client, "http://example.test", time.Second)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}
