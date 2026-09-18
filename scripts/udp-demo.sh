#!/usr/bin/env bash
set -euo pipefail

go run ./cmd/02_udp_echo/server >"${TMPDIR:-/tmp}/classroom11-udp.log" 2>&1 &
server_pid=$!
cleanup() { kill "$server_pid" 2>/dev/null || true; wait "$server_pid" 2>/dev/null || true; }
trap cleanup EXIT

sleep 0.4
go run ./cmd/02_udp_echo/client
