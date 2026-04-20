#!/usr/bin/env bash
set -u

cd "$(dirname "$0")/.."

PORT=18084
DEAD=2s
INTERVAL=300ms
LOG=$(mktemp)

cleanup() {
    [[ -n "${SRV_PID:-}" ]] && kill "$SRV_PID" 2>/dev/null
    [[ -n "${C1_PID:-}" ]] && kill "$C1_PID" 2>/dev/null
    [[ -n "${C2_PID:-}" ]] && kill "$C2_PID" 2>/dev/null
    wait 2>/dev/null
    rm -f "$LOG"
}
trap cleanup EXIT

[[ -x ./bin/heartbeat_server && -x ./bin/heartbeat_client ]] || make build >/dev/null

echo "== starting server (port=$PORT dead=$DEAD) =="
./bin/heartbeat_server --port "$PORT" --dead "$DEAD" --check 300ms >"$LOG" 2>&1 &
SRV_PID=$!
sleep 0.3

echo "== starting client-1 =="
./bin/heartbeat_client --host 127.0.0.1 --port "$PORT" --interval "$INTERVAL" >/dev/null 2>&1 &
C1_PID=$!

echo "== starting client-2 =="
./bin/heartbeat_client --host 127.0.0.1 --port "$PORT" --interval "$INTERVAL" >/dev/null 2>&1 &
C2_PID=$!

sleep 1.5

echo "== killing client-1 =="
kill "$C1_PID" 2>/dev/null
wait "$C1_PID" 2>/dev/null

echo "== waiting 3s (client-2 продолжает слать) =="
sleep 3

echo "== killing client-2 =="
kill "$C2_PID" 2>/dev/null

echo
echo "===== SERVER LOG ====="
cat "$LOG"
echo "======================"
echo

DOWN_COUNT=$(grep -c "assumed DOWN" "$LOG" || true)
CLIENTS_REGISTERED=$(grep -c "new client" "$LOG" || true)

RC=0
if [[ "$CLIENTS_REGISTERED" -ge 2 ]]; then
    echo "PASS: зарегистрировано >=2 клиентов ($CLIENTS_REGISTERED)"
else
    echo "FAIL: ожидали >=2 клиентов, получили $CLIENTS_REGISTERED"
    RC=1
fi

if [[ "$DOWN_COUNT" -eq 1 ]]; then
    echo "PASS: ровно один клиент помечен DOWN"
else
    echo "FAIL: ожидали 1 DOWN, получили $DOWN_COUNT"
    RC=1
fi

LAST_LINE=$(tail -n 5 "$LOG" | grep "heartbeat from" | tail -n 1 || true)
if [[ -n "$LAST_LINE" ]]; then
    echo "PASS: сервер продолжал принимать heartbeat после убийства client-1"
else
    echo "FAIL: после убийства client-1 сервер не получал heartbeat"
    RC=1
fi

exit $RC
