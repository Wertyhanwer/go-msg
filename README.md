# Web Messanger

---- 
## Setup
[SETUP INFO](SETUP.md)

---
## Services
- API 
- Front (View)
- nginx

---
### API
- **workdir:** `./app`
- **API:** ``http://localhost:8080/api/``

---
### View (web)
- **workdir:** `./www`
- **path:** ``http://localhost:8080/`` | ``http://localhost:8080/index.html``

---
### nginx
- **workdir:** `./nginx`

Обратный прокси, который перенаправляет запросы с пути /api/ к сервису api на порту 8080, а все остальные запросы — к сервису web на порту 80

---
### Тестирование базы данных

Чтобы проверить, что база данных работает и таблица сообщений создана:

1. Запустите проект (см. раздел Setup выше).
2. Откройте в браузере или с помощью curl:
   ```
   http://localhost:8080/api/testdb
   ```
3. Вы увидите ответ вида:
   ```json
   { "messages_count": 0 }
   ```
   Это количество сообщений в базе. Если добавите сообщения — число увеличится.

