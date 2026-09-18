#!/usr/bin/env bash
set -euo pipefail

go run ./cmd/01_tcp_echo/server >"${TMPDIR:-/tmp}/classroom11-tcp.log" 2>&1 &
server_pid=$!
cleanup() { kill "$server_pid" 2>/dev/null || true; wait "$server_pid" 2>/dev/null || true; }
trap cleanup EXIT

for _ in {1..40}; do
  if (echo > /dev/tcp/127.0.0.1/8081) 2>/dev/null; then break; fi
  sleep 0.05
done
go run ./cmd/01_tcp_echo/client
