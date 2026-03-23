## Инструкция по запуску
### Собрать бинарники
```bash
make build
```
### Однопоточный сервер
```bash
./bin/single_thread_server server_port
```
### Многопоточный сервер
```bash
./bin/multi_thread_server server_port
```
### Сервер с ограниченным кол-вом потоков
```bash
./bin/limited_thread_server server_port concurrency_level
```
### Клиент
```bash
./bin/client server_host server_port filename
```
