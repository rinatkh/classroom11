SHELL := /bin/bash
GO ?= go
PACKAGES := ./...
COVERAGE_FILE ?= coverage.out
COVERAGE_THRESHOLD ?= 90.0

.PHONY: help build fmt fmt-check vet test test-race coverage coverage-check ci clean tcp-server tcp-client tcp-demo udp-server udp-client udp-demo raw-http http-trace tls-http2 timeouts demos

help:
	@echo "Классная работа 11: сеть под HTTP"
	@echo "  make tcp-demo     - TCP echo: соединение и поток байтов"
	@echo "  make udp-demo     - UDP echo: отдельные датаграммы"
	@echo "  make raw-http     - HTTP/1.1, записанный вручную в TCP"
	@echo "  make http-trace   - события DNS/connect/reuse/first byte"
	@echo "  make tls-http2    - локальные TLS и HTTP/2"
	@echo "  make timeouts     - тайм-аут медленного ответа"
	@echo "  make demos        - все автономные демонстрации"
	@echo "  make ci           - форматирование, vet, тесты, race, coverage, build"

build:
	@mkdir -p bin
	$(GO) build -buildvcs=false -o bin/tcp-server ./cmd/01_tcp_echo/server
	$(GO) build -buildvcs=false -o bin/tcp-client ./cmd/01_tcp_echo/client
	$(GO) build -buildvcs=false -o bin/udp-server ./cmd/02_udp_echo/server
	$(GO) build -buildvcs=false -o bin/udp-client ./cmd/02_udp_echo/client
	$(GO) build -buildvcs=false -o bin/raw-http ./cmd/03_raw_http
	$(GO) build -buildvcs=false -o bin/http-trace ./cmd/04_http_trace
	$(GO) build -buildvcs=false -o bin/tls-http2 ./cmd/05_tls_http2
	$(GO) build -buildvcs=false -o bin/timeouts ./cmd/06_timeouts

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './bin/*')

fmt-check:
	@files="$$(gofmt -l $$(find . -name '*.go' -not -path './bin/*'))"; \
	if [ -n "$$files" ]; then echo "Найдены неотформатированные Go-файлы:"; echo "$$files"; exit 1; fi

vet:
	$(GO) vet $(PACKAGES)

test:
	$(GO) test $(PACKAGES)

test-race:
	$(GO) test -race $(PACKAGES)

coverage:
	$(GO) test ./internal/... -covermode=atomic -coverprofile=$(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)

coverage-check: coverage
	@coverage="$$($(GO) tool cover -func=$(COVERAGE_FILE) | awk '/^total:/ {gsub("%", "", $$3); print $$3}')"; \
	awk -v coverage="$$coverage" -v threshold="$(COVERAGE_THRESHOLD)" 'BEGIN { \
		if (coverage + 0 < threshold + 0) { printf "Покрытие %.1f%% ниже порога %.1f%%\n", coverage, threshold; exit 1 } \
		printf "Покрытие %.1f%% не ниже порога %.1f%%\n", coverage, threshold; \
	}'

tcp-server:
	$(GO) run ./cmd/01_tcp_echo/server

tcp-client:
	$(GO) run ./cmd/01_tcp_echo/client

tcp-demo:
	./scripts/tcp-demo.sh

udp-server:
	$(GO) run ./cmd/02_udp_echo/server

udp-client:
	$(GO) run ./cmd/02_udp_echo/client

udp-demo:
	./scripts/udp-demo.sh

raw-http:
	$(GO) run ./cmd/03_raw_http

http-trace:
	$(GO) run ./cmd/04_http_trace

tls-http2:
	$(GO) run ./cmd/05_tls_http2

timeouts:
	$(GO) run ./cmd/06_timeouts

demos: tcp-demo udp-demo raw-http http-trace tls-http2 timeouts

ci: fmt-check vet test test-race coverage-check build

clean:
	rm -f bin/tcp-server bin/tcp-client bin/udp-server bin/udp-client bin/raw-http bin/http-trace bin/tls-http2 bin/timeouts coverage.out
	@rmdir bin 2>/dev/null || true
