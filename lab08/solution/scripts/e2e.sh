#!/usr/bin/env bash
# E2E-тесты:
# - C2S
# - S2C
# - дуплекс
# - восстановление после порчи payload (через checksum)
#
# Сервер запускается фоном, после завершения клиента — убивается
#
# Запуск: ./scripts/e2e.sh
# Перед запуском нужно собрать бинари: make build

set -u

VERBOSE=0

while getopts "v" opt; do
  case $opt in
    v) VERBOSE=1 ;;
  esac
done

shift $((OPTIND-1))

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
BIN="$ROOT_DIR/bin"
TMP=$(mktemp -d -t snw_e2e.XXXXXX)
PORT=19100
PASS=0
FAIL=0

if [[ ! -x "$BIN/snw_server" || ! -x "$BIN/snw_client" ]]; then
    echo "не найдены $BIN/snw_server / $BIN/snw_client. сначала: make build" >&2
    exit 2
fi

cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

run_case() {
    local name=$1; shift
    local srv_args=()
    while [[ $# -gt 0 && "$1" != "--" ]]; do srv_args+=("$1"); shift; done
    shift # eat --
    local cli_args=("$@")

    local srv_log="$TMP/srv.log" cli_log="$TMP/cli.log"
    : > "$srv_log"; : > "$cli_log"

    "$BIN/snw_server" --addr ":$PORT" "${srv_args[@]}" >"$srv_log" 2>&1 &
    local spid=$!
    sleep 0.2

    "$BIN/snw_client" --host 127.0.0.1 --port "$PORT" "${cli_args[@]}" >"$cli_log" 2>&1
    local crc=$?

    sleep 0.1
    kill "$spid" 2>/dev/null
    wait "$spid" 2>/dev/null

    local ok=1
    if [[ $crc -ne 0 ]]; then ok=0; fi
    if [[ -n "${CMP1_A:-}" ]]; then
        cmp -s "$CMP1_A" "$CMP1_B" || ok=0
    fi
    if [[ -n "${CMP2_A:-}" ]]; then
        cmp -s "$CMP2_A" "$CMP2_B" || ok=0
    fi

    if [[ $ok -eq 1 ]]; then
        printf "PASS  %s\n" "$name"
        PASS=$((PASS+1))
    else
        printf "FAIL  %s (client_rc=%d)\n" "$name" "$crc"
        FAIL=$((FAIL+1))
    fi

    if [[ $VERBOSE -eq 1 ]]; then
        echo "--- server log ---"; sed 's/^/  /' "$srv_log"
        echo "--- client log ---"; sed 's/^/  /' "$cli_log"
    fi

    unset CMP1_A CMP1_B CMP2_A CMP2_B
}

dd if=/dev/urandom of="$TMP/in_a.bin" bs=1024 count=8  status=none
dd if=/dev/urandom of="$TMP/in_b.bin" bs=1024 count=8  status=none

# 1. C2S
CMP1_A="$TMP/in_a.bin" CMP1_B="$TMP/out_c2s.bin" \
run_case "C2S file transfer (loss=0.2)" \
    --recv-file "$TMP/out_c2s.bin" --loss 0.2 --timeout 50ms --chunk-size 512 --seed 1 \
    -- \
    --send-file "$TMP/in_a.bin"     --loss 0.2 --timeout 50ms --chunk-size 512 --seed 2

# 2. S2C
CMP1_A="$TMP/in_a.bin" CMP1_B="$TMP/out_s2c.bin" \
run_case "S2C file transfer (loss=0.2)" \
    --send-file "$TMP/in_a.bin"     --loss 0.2 --timeout 50ms --chunk-size 512 --seed 11 \
    -- \
    --recv-file "$TMP/out_s2c.bin" --loss 0.2 --timeout 50ms --chunk-size 512 --seed 22

# 3. Дуплекс
CMP1_A="$TMP/in_a.bin" CMP1_B="$TMP/cli_recv.bin" \
CMP2_A="$TMP/in_b.bin" CMP2_B="$TMP/srv_recv.bin" \
run_case "Duplex (loss=0.2)" \
    --send-file "$TMP/in_a.bin" --recv-file "$TMP/srv_recv.bin" \
        --loss 0.2 --timeout 50ms --chunk-size 512 --seed 31 \
    -- \
    --send-file "$TMP/in_b.bin" --recv-file "$TMP/cli_recv.bin" \
        --loss 0.2 --timeout 50ms --chunk-size 512 --seed 32

# 4. Порча payload, loss=0 - checksum обязан ловить, retx - восстанавливать.
CMP1_A="$TMP/in_a.bin" CMP1_B="$TMP/out_corr.bin" \
run_case "Corruption recovery via checksum (corrupt=0.5, loss=0)" \
    --recv-file "$TMP/out_corr.bin" --loss 0.0 --timeout 50ms --chunk-size 512 --seed 41 \
    -- \
    --send-file "$TMP/in_a.bin" --loss 0.0 --corrupt-prob 0.5 \
        --timeout 50ms --chunk-size 512 --seed 42

echo
echo "Итого: PASS=$PASS FAIL=$FAIL"
[[ $FAIL -eq 0 ]] || exit 1
