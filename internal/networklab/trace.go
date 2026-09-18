package networklab

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// TraceEvent — одно наблюдаемое событие внутри http.Client.
type TraceEvent struct {
	Request int
	Name    string
	Detail  string
}

// DoTracedRequests выполняет несколько GET через один Client и собирает события httptrace.
// Body полностью читается и закрывается: только тогда Transport обычно может вернуть connection в pool.
func DoTracedRequests(ctx context.Context, client *http.Client, url string, count int) ([]TraceEvent, error) {
	var mu sync.Mutex
	events := make([]TraceEvent, 0, count*5)
	add := func(request int, name, detail string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, TraceEvent{Request: request, Name: name, Detail: detail})
	}

	for i := 1; i <= count; i++ {
		requestNumber := i
		trace := &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				add(requestNumber, "DNSStart", info.Host)
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				add(requestNumber, "DNSDone", fmt.Sprintf("addresses=%d err=%v", len(info.Addrs), info.Err))
			},
			ConnectStart: func(network, addr string) {
				add(requestNumber, "ConnectStart", network+" "+addr)
			},
			ConnectDone: func(network, addr string, err error) {
				add(requestNumber, "ConnectDone", fmt.Sprintf("%s %s err=%v", network, addr, err))
			},
			TLSHandshakeStart: func() {
				add(requestNumber, "TLSHandshakeStart", "")
			},
			TLSHandshakeDone: func(state tls.ConnectionState, err error) {
				add(requestNumber, "TLSHandshakeDone", fmt.Sprintf("alpn=%s err=%v", state.NegotiatedProtocol, err))
			},
			GotConn: func(info httptrace.GotConnInfo) {
				add(requestNumber, "GotConn", fmt.Sprintf("reused=%t idle=%t", info.Reused, info.WasIdle))
			},
			GotFirstResponseByte: func() {
				add(requestNumber, "FirstByte", time.Now().Format("15:04:05.000"))
			},
		}

		request, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("build request %d: %w", i, err)
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("perform request %d: %w", i, err)
		}
		if _, err := io.Copy(io.Discard, response.Body); err != nil {
			response.Body.Close()
			return nil, fmt.Errorf("read response %d: %w", i, err)
		}
		if err := response.Body.Close(); err != nil {
			return nil, fmt.Errorf("close response %d: %w", i, err)
		}
	}
	return events, nil
}
