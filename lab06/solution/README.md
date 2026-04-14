# Инструкция по запуску
1. Собрать бинарники
    ```bash
   make build
    ```
2. Запустить FTP сервер
    ```bash
   ./bin/ftp_server --root <dir> --port <port> --user ... --pass ...
    ```
3. Запустить FTP клиента
    ```bash
   ./bin/ftp_client --host <addr> --port <port> --user ... --pass ...
    ```
4. Запустить FTP клиента (GUI)
    ```bash
   ./bin/ftp_gui
    ```
5. Убрать мусор 
    ```bash
   make clean
    ```