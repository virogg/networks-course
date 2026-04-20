#!/usr/bin/env bash
set -u

cd "$(dirname "$0")/.."

PORT=18085
LOG=$(mktemp)

cleanup() {
    [[ -n "${SRV_PID:-}" ]] && kill "$SRV_PID" 2>/dev/null
    wait 2>/dev/null
    rm -f "$LOG"
}
trap cleanup EXIT

[[ -x ./bin/heartbeat_server ]] || make build >/dev/null

echo "== starting server (port=$PORT) =="
./bin/heartbeat_server --port "$PORT" --dead 30s --check 500ms >"$LOG" 2>&1 &
SRV_PID=$!
sleep 0.3

exec 3>/dev/udp/127.0.0.1/$PORT

nanos() { printf "%d000000000" "$(date +%s)"; }

send_seq() {
    printf "%s %s" "$1" "$(nanos)" >&3
    sleep 0.1
}

echo "== sending seq=1 =="; send_seq 1
echo "== sending seq=2 =="; send_seq 2
echo "== skipping 3,4; sending seq=5 =="; send_seq 5
echo "== skipping 6; sending seq=7 =="; send_seq 7

exec 3>&-

sleep 0.5

echo
echo "===== SERVER LOG ====="
cat "$LOG"
echo "======================"
echo

RC=0

if grep -q "missed 2 packet(s) before seq=5" "$LOG"; then
    echo "PASS: обнаружен gap 3..4 (2 пропущенных перед seq=5)"
else
    echo "FAIL: ожидали 'missed 2 packet(s) before seq=5'"
    RC=1
fi

if grep -q "missed 1 packet(s) before seq=7" "$LOG"; then
    echo "PASS: обнаружен gap 6 (1 пропущенный перед seq=7)"
else
    echo "FAIL: ожидали 'missed 1 packet(s) before seq=7'"
    RC=1
fi

if grep -q "seq=7.*missed_since_start=3" "$LOG"; then
    echo "PASS: суммарно missed_since_start=3"
else
    echo "FAIL: ожидали 'seq=7 ... missed_since_start=3'"
    RC=1
fi

exit $RC
