package networklab

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// RawHTTPResponse — разобранный ответ на HTTP/1.1-запрос, который мы вручную записали в TCP.
type RawHTTPResponse struct {
	Status     string
	StatusCode int
	Header     http.Header
	Body       string
}

// BuildRawGET возвращает текст корректного HTTP/1.1-запроса.
// Пустая строка после headers обязательна: последовательность \r\n\r\n завершает секцию заголовков.
func BuildRawGET(host, path string) string {
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", path, host)
}

// SendRawGET показывает, что HTTP/1.1 — это правила записи байтов поверх TCP.
// http.ReadResponse используется только для удобного разбора ответа, а запрос пишется вручную.
func SendRawGET(ctx context.Context, address, host, path string) (RawHTTPResponse, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return RawHTTPResponse{}, fmt.Errorf("dial HTTP server: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return RawHTTPResponse{}, fmt.Errorf("set deadline: %w", err)
		}
	}

	if _, err := io.WriteString(conn, BuildRawGET(host, path)); err != nil {
		return RawHTTPResponse{}, fmt.Errorf("write raw HTTP request: %w", err)
	}

	response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodGet})
	if err != nil {
		return RawHTTPResponse{}, fmt.Errorf("read HTTP response: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return RawHTTPResponse{}, fmt.Errorf("read HTTP body: %w", err)
	}

	return RawHTTPResponse{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Header:     response.Header.Clone(),
		Body:       strings.TrimSpace(string(body)),
	}, nil
}
