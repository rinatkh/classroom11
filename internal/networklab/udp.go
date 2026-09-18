package networklab

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// ServeUDPEcho читает одну датаграмму целиком и отправляет её обратно тому же адресу.
// В отличие от TCP, UDP сохраняет границы сообщений: один Write соответствует одной датаграмме.
func ServeUDPEcho(ctx context.Context, conn net.PacketConn) error {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	defer close(done)

	buffer := make([]byte, 64*1024)
	for {
		n, address, err := conn.ReadFrom(buffer)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("read UDP datagram: %w", err)
		}
		if _, err := conn.WriteTo(buffer[:n], address); err != nil {
			return fmt.Errorf("write UDP datagram: %w", err)
		}
	}
}

// UDPExchange отправляет сообщения отдельными датаграммами и читает ответы.
// Здесь нет Accept: UDP-сокет получает пакеты от разных адресов через один PacketConn.
func UDPExchange(ctx context.Context, address string, messages []string) ([]string, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "udp", address)
	if err != nil {
		return nil, fmt.Errorf("dial UDP server: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("set UDP deadline: %w", err)
		}
	}

	buffer := make([]byte, 64*1024)
	responses := make([]string, 0, len(messages))
	for _, message := range messages {
		if _, err := conn.Write([]byte(message)); err != nil {
			return nil, fmt.Errorf("send UDP datagram: %w", err)
		}
		n, err := conn.Read(buffer)
		if err != nil {
			return nil, fmt.Errorf("read UDP datagram: %w", err)
		}
		responses = append(responses, string(buffer[:n]))
	}
	return responses, nil
}
