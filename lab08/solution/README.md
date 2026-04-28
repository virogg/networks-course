# Инструкция по запуску

## Сборка

```bash
make build
```

Итого появятся `bin/snw_server` и `bin/snw_client`.

## Stop-and-Wait поверх UDP

Общие флаги (`--loss` действует на отправляющую сторону: с этой вероятностью
кадр молча отбрасывается; так получается потеря в обоих направлениях, потому
что обе стороны теряют свои исходящие кадры):

```
--addr / --host --port    адрес UDP
--timeout                 таймаут retransmit
--chunk-size              размер payload в одном кадре
--loss                    вероятность потери исходящего кадра
--corrupt-prob            вероятность одиночного бит-сбоя в payload
--send-file               файл, который сторона отправит peer'у (опционально)
--recv-file               куда сохранить входящий файл (опционально)
--seed                    seed для генератора потерь (0 = текущее время)
```

### A. Передача файла от клиента серверу

Терминал 1:
```bash
./bin/snw_server --addr <addr> --recv-file <recv_file> --loss ... --timeout ...
```

Терминал 2:
```bash
./bin/snw_client --host <host> --port <port> --send-file <send_file> --loss ... --timeout ...
```

После завершения проверить целостность:
```bash
cmp <send_file> <recv_file>
```

В логах видны события `LOSS: drop ...`, `sender: timeout ..., retransmit`,
`receiver: duplicate seq=..., re-ACK`, `receiver: EOF`.

### Б. Дуплекс

#### Передача в обе стороны одновременно
Оба эндпоинта задают и `--send-file`, и `--recv-file`:

Терминал 1:
```bash
./bin/snw_server --addr <addr> --send-file <srv_in_file> --recv-file <srv_rcv_file> --loss ...
```

Терминал 2:
```bash
./bin/snw_client --host <host> --port <port> --send-file <cli_in_file> --recv-file <cli_rcv_file> --loss ...
```

#### Передача только от сервера клиенту
У клиента указывается лишь `--recv-file`
(клиент шлёт HELLO, чтобы сервер узнал его адрес, повторяет до первого ответа):
```bash
./bin/snw_server --addr <addr> --send-file <in_file> --loss ...
./bin/snw_client --host <host> --port <port> --recv-file <out_file> --loss ...
```

### В. Контрольные суммы в протоколе

Для имитации шума в канале (одиночный битовый сбой в payload уже после
вычисления контрольной суммы) использовать `--corrupt-prob`. <br>
На принимающей стороне такие кадры падают на проверке checksum и приводят к retransmit:

```bash
./bin/snw_server --addr <addr> --recv-file <out_file> --loss 0.0 --timeout ...
```
```bash
./bin/snw_client --host <host> --port <port> --send-file <in_file> --loss 0.0 --corrupt-prob 0.4 --timeout ...
```

В логах сервера: `drop bad frame: checksum mismatch`. У клиента — соответствующие
`retransmit`. Файл по итогу совпадает с исходным.

## Контрольная сумма

Реализация по RFC 1071 -- `pkg/checksum`
## Тесты

- Юнит-тесты:
    ```bash
    go test ./...
    ```

- End-to-end:
    ```bash
    ./scripts/e2e.sh [-v]
    ```

- Все сразу:
    ```bash
    make tests
    ```

## Очистка

```bash
make clean
```
