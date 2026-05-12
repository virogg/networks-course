# Инструкция по запуску

## Сборка

```bash
make build
```
```bash
make test
```
```bash
make clean
```

## 1. Эхо-запросы через ICMP

Требует raw-сокета $\Rightarrow$ запускать через `sudo`.

```bash
sudo ./bin/ping <host> [--interval 1s] [--timeout 1s] [--count N] [--size 56]
```

## 2. Go back-N протокол
 - ```bash
   ./bin/gbn-server --addr <addr> --out <out_file> [--loss-rate 0.1]
   ```
 - ```bash
   ./bin/gbn-client --addr <addr> --file <in_file> [--chunk 1024] [--window 4] [--timeout 500ms] [--loss-rate 0.1]
   ```

`--loss-rate` имитирует потери (на отправителе — DATA/FIN, на получателе — ACK/FIN-ACK), чтобы в логах было видно повторы.

Проверка целостности после передачи: `cmp <in_file> <out_file>`.
