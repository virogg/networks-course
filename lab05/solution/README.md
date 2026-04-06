# Инструкция по запуску:
- Собрать бинарники:
```bash
make build
```
- Запустить нужный:
  - **A. Почта и SMTP**:  
    - 1: Почтовый клиент
      ```bash
      ./bin/mail_client -from <addr> -to <addr> -pass <pass> [-format txt|html] [-subject ...] [-body ...]
      ```  
    - 2: SMTP-клиент
      ```bash
      ./bin/smtp_client -from <addr> -to <addr> -pass <pass> [-subject ...] [-body ...]
      ```  
    - 3: SMTP-клиент: бинарные данные
      ```bash
      ./bin/smtp_client_binary -from <addr> -to <addr> -pass <pass> -image <path> [-subject ...] [-body ...]
      ```    
  - **Б. Удаленный запуск команд**
    - Клиент
      ```bash
      ./bin/remote_client -host <host> -port <port> -cmd ...
      ```
    - Сервер
      ```bash
      ./bin/remote_server -port <port>
      ```
  - **В. Широковещательная рассылка через UDP**
    - Клиент
      ```bash
      ./bin/udp_client -port <port>
      ```
    - Сервер
      ```bash
      ./bin/udp_server -port <port>
      ```