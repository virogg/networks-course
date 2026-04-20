#!/usr/bin/env bash
set -u

cd "$(dirname "$0")/.."

PORT=18083
DEAD=2s
INTERVAL=300ms
LOG=$(mktemp)

cleanup() {
    [[ -n "${SRV_PID:-}" ]] && kill "$SRV_PID" 2>/dev/null
    [[ -n "${CLI_PID:-}" ]] && kill "$CLI_PID" 2>/dev/null
    wait 2>/dev/null
    rm -f "$LOG"
}
trap cleanup EXIT

[[ -x ./bin/heartbeat_server && -x ./bin/heartbeat_client ]] || make build >/dev/null

echo "== starting server (port=$PORT dead=$DEAD) =="
./bin/heartbeat_server --port "$PORT" --dead "$DEAD" --check 300ms >"$LOG" 2>&1 &
SRV_PID=$!
sleep 0.3

echo "== starting client (interval=$INTERVAL) =="
./bin/heartbeat_client --host 127.0.0.1 --port "$PORT" --interval "$INTERVAL" >/dev/null 2>&1 &
CLI_PID=$!

echo "== running client for 1.5s =="
sleep 1.5

echo "== killing client =="
kill "$CLI_PID" 2>/dev/null
wait "$CLI_PID" 2>/dev/null

echo "== waiting 3s for DOWN detection =="
sleep 3

echo
echo "===== SERVER LOG ====="
cat "$LOG"
echo "======================"
echo

if grep -q "assumed DOWN" "$LOG"; then
    echo "PASS: клиент помечен DOWN"
    exit 0
else
    echo "FAIL: строки 'assumed DOWN' нет в логе"
    exit 1
fi
