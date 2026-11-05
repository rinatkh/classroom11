#!/usr/bin/env bash
set -euo pipefail

demo_dir="$(mktemp -d "${TMPDIR:-/tmp}/classroom11-udp.XXXXXX")"
server_pid=""
cleanup() {
  if [[ -n "$server_pid" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -rf "$demo_dir"
}
trap cleanup EXIT

"${GO:-go}" build -o "$demo_dir/server" ./cmd/02_udp_echo/server
"${GO:-go}" build -o "$demo_dir/client" ./cmd/02_udp_echo/client
"$demo_dir/server" -addr 127.0.0.1:0 -ready-file "$demo_dir/address" >"$demo_dir/server.log" 2>&1 &
server_pid=$!

address=""
for _ in {1..40}; do
  # Не зависим от LOG_LEVEL/LOG_FORMAT и текста служебного сообщения.
  if [[ -s "$demo_dir/address" ]]; then
    address="$(cat "$demo_dir/address")"
    break
  fi
  if ! kill -0 "$server_pid" 2>/dev/null; then break; fi
  sleep 0.05
done
if [[ -z "$address" ]]; then
  cat "$demo_dir/server.log" >&2
  echo "UDP server не запустился" >&2
  exit 1
fi
"$demo_dir/client" -addr "$address"
