package networklab

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// TLSInfo содержит только те поля, которые нужны для учебной демонстрации.
type TLSInfo struct {
	HTTPProtocol string
	TLSVersion   string
	ALPN         string
	ServerName   string
}

// InspectTLS выполняет HTTPS-запрос и показывает, что согласовали клиент и сервер.
// TLS защищает соединение, а ALPN выбирает HTTP/1.1 или HTTP/2 внутри этого соединения.
func InspectTLS(ctx context.Context, client *http.Client, url string) (TLSInfo, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return TLSInfo{}, fmt.Errorf("build HTTPS request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return TLSInfo{}, fmt.Errorf("perform HTTPS request: %w", err)
	}
	defer response.Body.Close()
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		return TLSInfo{}, fmt.Errorf("read HTTPS body: %w", err)
	}
	if response.TLS == nil {
		return TLSInfo{}, fmt.Errorf("response does not use TLS")
	}

	return TLSInfo{
		HTTPProtocol: response.Proto,
		TLSVersion:   tlsVersionName(response.TLS.Version),
		ALPN:         response.TLS.NegotiatedProtocol,
		ServerName:   response.TLS.ServerName,
	}, nil
}

func tlsVersionName(version uint16) string {
	switch version {
	case 0x0304:
		return "TLS 1.3"
	case 0x0303:
		return "TLS 1.2"
	default:
		return fmt.Sprintf("0x%04x", version)
	}
}
