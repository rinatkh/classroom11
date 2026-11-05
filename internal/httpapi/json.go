package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/rinatkh/classroom11/internal/logging"
)

var errSeveralJSONValues = errors.New("body должен содержать ровно одно JSON-значение")

// decodeJSON читает JSON прямо из body запроса в Go-структуру.
//
// DisallowUnknownFields помогает заметить опечатку клиента: поле "titel"
// не будет молча проигнорировано вместо ожидаемого "title".
func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	// После первого объекта разрешён только конец потока. Так запрос
	// `{...}{...}` не будет частично принят как корректный.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errSeveralJSONValues
		}
		return err
	}
	return nil
}

// writeJSON соблюдает порядок HTTP-ответа:
// сначала headers, затем status, затем body.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		logging.Error(context.Background(), "HTTP response write failed", err, "status", status)
		// Статус и часть ответа уже могли уйти клиенту, поэтому заменить ответ
		// на 500 здесь поздно. В учебном API кодируются только простые структуры,
		// для которых Encode не должен завершаться ошибкой.
		return
	}
}

// writeError возвращает ошибку в одном и том же JSON-формате.
func writeError(w http.ResponseWriter, status int, message string) {
	slog.Debug("HTTP request rejected", "status", status)
	writeJSON(w, status, map[string]string{"error": message})
}

// Как не нужно:
//
//	w.Write([]byte(`{"error":"not found"}`)) // status останется 200
//	w.WriteHeader(http.StatusNotFound)          // менять status уже поздно
//
// Первый Write начинает body и неявно отправляет 200 OK. Поэтому правильный
// порядок всегда такой: Header().Set → WriteHeader → Encode или Write.
