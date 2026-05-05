# Инструкция по запуску

1. Собрать бинарники
    ```bash
    make build
    ```
2. IP-адрес и маска сети
   -  ```bash
      ./bin/ipinfo
      ```
      - только IPv4, не-loopback
    - ```bash
      ./bin/ipinfo --all
      ```
      - включая IPv6 и loopback
3. Доступные порты
    ```bash
    ./bin/portscan --ip <ip> --from <n> --to <n> [--proto tcp|udp|both] [--mode auto|local|remote] [--workers 256] [--timeout 1.5s]
   ``` 
   - ``` bash
     ./bin/portscan --ip 127.0.0.1 --from 8000 --to 8100 --proto both
     ```
     - локальный IP (свободные = можно занять): 
   - ``` bash 
     ./bin/portscan --ip 142.250.181.142 --from 79 --to 81
     ```
     - удалённый хост (свободные = открытые TCP-порты):
   <hr>
   Семантика «свободного» порта зависит от режима:<br>
- `local` -- `Listen` (TCP) / `ListenPacket` (UDP) на `ip:port` успешен $\Rightarrow$ порт *свободен для биндинга*. Применим только к собственным интерфейсам.
- `remote` -- TCP-handshake (`DialTimeout`) проходит $\Rightarrow$ порт *открыт* (на нём слушает сервис). UDP в этом режиме не поддерживается.
- `auto` (по умолчанию) — `local` если IP принадлежит локальному интерфейсу, иначе `remote`.
4. Широковещательная рассылка для подсчета копий приложения
    - ```bash
      ./bin/copies [--port 9999] [--interval 2s] [--dead-multiplier 3]
      ```
      - Консоль
    - ```bash
      ./bin/copies_gui [--port 9999] [--interval 2s] [--dead-multiplier 3]
      ```
      - GUI
   
    Пир считается ушедшим, если от него не пришло сообщения за `dead-multiplier * interval`.
5. Убрать мусор
    ```bash
    make clean
    ```
