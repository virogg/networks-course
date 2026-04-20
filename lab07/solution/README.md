# Инструкция по запуску
1. Собрать бинарники
    ```bash
   make build
    ```
2. Запустить ping сервер
    ```bash
   ./bin/ping_server --port <port>
    ```
3. Запустить ping клиента
    ```bash
   ./bin/ping_client --host <addr> --port <port> --timeout ... --count ... --stats
    ```
4. Запустить heartbeat сервер
    ```bash
   ./bin/heartbeat_server --port <port> --dead ... --check ...
    ```
5. Запустить heartbeat клиента
    ```bash
   ./bin/heartbeat_client --host <addr> --port <port> --interval ...
    ```
6. Убрать мусор
    ```bash
   make clean
    ```