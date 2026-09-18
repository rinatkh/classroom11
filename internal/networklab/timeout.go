package networklab

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetWithTimeout ограничивает всё время запроса через context.
// Это понятный учебный вариант: при отмене клиент прекращает ожидание ответа.
func GetWithTimeout(parent context.Context, client *http.Client, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build timed request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("perform timed request: %w", err)
	}
	defer response.Body.Close()
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		return fmt.Errorf("read timed response: %w", err)
	}
	return nil
}
