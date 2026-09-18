// Package networklab содержит маленькие сетевые эксперименты занятия.
// Каждый эксперимент отделяет одну идею от остальных, чтобы сеть не выглядела магией.
package networklab

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// ServeTCPEcho принимает TCP-соединения и запускает отдельную goroutine для каждого клиента.
// TCP передаёт поток байтов: границы отдельных вызовов Write на другой стороне не сохраняются.
func ServeTCPEcho(ctx context.Context, listener net.Listener) error {
	var clients sync.WaitGroup
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = listener.Close()
		case <-done:
		}
	}()
	defer close(done)
	defer clients.Wait()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept TCP connection: %w", err)
		}

		clients.Add(1)
		go func() {
			defer clients.Done()
			_ = HandleTCPLines(conn)
		}()
	}
}

// HandleTCPLines добавляет к каждой строке префикс echo и отправляет её обратно.
// Символ \n здесь — наше прикладное правило framing: он отмечает конец сообщения в TCP-потоке.
func HandleTCPLines(conn net.Conn) error {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)
	for scanner.Scan() {
		if _, err := fmt.Fprintf(writer, "echo: %s\n", scanner.Text()); err != nil {
			return fmt.Errorf("write TCP response: %w", err)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("flush TCP response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read TCP stream: %w", err)
	}
	return nil
}

// TCPExchange отправляет несколько строк через одно соединение и читает столько же ответов.
// Переиспользование соединения позволяет увидеть: TCP-соединение не обязано завершаться после сообщения.
func TCPExchange(ctx context.Context, address string, messages []string) ([]string, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial TCP server: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("set TCP deadline: %w", err)
		}
	}

	reader := bufio.NewReader(conn)
	responses := make([]string, 0, len(messages))
	for _, message := range messages {
		if _, err := io.WriteString(conn, strings.TrimSuffix(message, "\n")+"\n"); err != nil {
			return nil, fmt.Errorf("send TCP message: %w", err)
		}
		response, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read TCP response: %w", err)
		}
		responses = append(responses, strings.TrimSuffix(response, "\n"))
	}
	return responses, nil
}
