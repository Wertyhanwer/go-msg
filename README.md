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

---
## API Endpoints

### Сообщения

#### Получить все сообщения
- **URL:** `/api/messages`
- **Метод:** GET
- **Ответ:**
  ```json
  [
    {
      "id": 1,
      "text": "Текст сообщения",
      "created_at": "2025-07-06T16:35:01.029175Z"
    }
  ]
  ```

#### Создать новое сообщение
- **URL:** `/api/messages/create`
- **Метод:** POST
- **Тело запроса:**
  ```json
  {
    "text": "Текст сообщения"
  }
  ```
- **Ответ:**
  ```json
  {
    "id": 1,
    "text": "Текст сообщения",
    "created_at": "2025-07-06T16:35:01.029175Z"
  }
  ```

#### Получить сообщение по ID
- **URL:** `/api/messages/get?id={id}`
- **Метод:** GET
- **Параметры:**
  - id: ID сообщения (число)
- **Ответ:**
  ```json
  {
    "id": 1,
    "text": "Текст сообщения",
    "created_at": "2025-07-06T16:35:01.029175Z"
  }
  ```

#### Обновить сообщение
- **URL:** `/api/messages/update?id={id}`
- **Метод:** PUT
- **Параметры:**
  - id: ID сообщения (число)
- **Тело запроса:**
  ```json
  {
    "text": "Новый текст сообщения"
  }
  ```
- **Ответ:**
  ```json
  {
    "id": 1,
    "text": "Новый текст сообщения",
    "created_at": "2025-07-06T16:35:01.029175Z"
  }
  ```

#### Удалить сообщение
- **URL:** `/api/messages/delete?id={id}`
- **Метод:** DELETE
- **Параметры:**
  - id: ID сообщения (число)
- **Ответ:** пустой (статус 200 OK при успешном удалении)

### Тестовые endpoints

#### Проверка базы данных
- **URL:** `/api/testdb`
- **Метод:** GET
- **Ответ:**
  ```json
  {
    "messages_count": 1
  }
  ```
  Возвращает количество сообщений в базе данных.

---
## Примеры использования

### Создание сообщения через curl
```bash
curl -X POST http://localhost:8080/api/messages/create \
  -H "Content-Type: application/json" \
  -d "{\"text\":\"Hello, World!\"}"
```

### Получение всех сообщений через curl
```bash
curl http://localhost:8080/api/messages
```

### Обновление сообщения через curl
```bash
curl -X PUT http://localhost:8080/api/messages/update?id=1 \
  -H "Content-Type: application/json" \
  -d "{\"text\":\"Updated message!\"}"
```

### Удаление сообщения через curl
```bash
curl -X DELETE http://localhost:8080/api/messages/delete?id=1
```

