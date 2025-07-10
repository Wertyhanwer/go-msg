# 📨 Система сообщений

Современная система сообщений с Go API, PostgreSQL и веб-интерфейсом.

## 🚀 Быстрый старт

### Windows
```bash
setup.bat
```

### Linux/macOS
```bash
chmod +x setup.py
python3 setup.py
```

Или вручную:
```bash
docker compose up -d --build
```

## 🌐 Доступ

- **Веб-интерфейс**: http://localhost:8080
- **API корень**: http://localhost:8080/api/
- **Тест БД**: http://localhost:8080/api/testdb

## 📡 API Endpoints

| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/` | Проверка API |
| GET | `/api/messages` | Список сообщений (с пагинацией) |
| POST | `/api/messages/create` | Создать сообщение |
| GET | `/api/messages/get?id=N` | Получить сообщение |
| PUT | `/api/messages/update?id=N` | Обновить сообщение |
| DELETE | `/api/messages/delete?id=N` | Удалить сообщение |
| GET | `/api/testdb` | Тест подключения к БД |

### Примеры запросов

```bash
# Создать сообщение
curl -X POST http://localhost:8080/api/messages/create \
  -H "Content-Type: application/json" \
  -d '{"text": "Привет мир!"}'

# Получить все сообщения
curl http://localhost:8080/api/messages?limit=10&offset=0

# Удалить сообщение
curl -X DELETE http://localhost:8080/api/messages/delete?id=1
```

## 🏗 Архитектура

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────┐
│   Nginx Proxy   │    │   Go API     │    │ PostgreSQL  │
│   (Port 8080)   │◄──►│  (Port 8080) │◄──►│   (Port     │
│                 │    │              │    │    5432)    │
└─────────────────┘    └──────────────┘    └─────────────┘
         │
         ▼
┌─────────────────┐
│ Static Files    │
│ (HTML/CSS/JS)   │
└─────────────────┘
```

### Компоненты:
- **Nginx**: Прокси-сервер для API и статических файлов
- **Go API**: Backend на Golang с pgx драйвером
- **PostgreSQL**: База данных для хранения сообщений
- **Frontend**: SPA на ванильном JavaScript

## 🛠 Технологии

- **Backend**: Go 1.21+ с pgx/v5
- **Frontend**: HTML5, CSS3, ES6+
- **База данных**: PostgreSQL 13
- **Proxy**: Nginx Alpine
- **Контейнеризация**: Docker & Docker Compose

## 📂 Структура проекта

```
go-msg/
├── app/                 # Go API
│   ├── handlers/        # HTTP обработчики
│   ├── models/          # Модели данных
│   ├── storage/         # Работа с БД
│   ├── utils/           # Утилиты (логгер)
│   └── main.go         # Точка входа
├── www/                 # Frontend
│   ├── index.html      # Главная страница
│   ├── scripts.js      # JavaScript
│   └── styles.css      # Стили
├── nginx/              # Nginx конфигурация
└── docker-compose.yml  # Оркестрация
```

## 🔧 Управление

```bash
# Запуск
docker compose up -d

# Просмотр логов
docker compose logs -f

# Перезапуск API
docker compose restart api

# Остановка
docker compose down

# Полная очистка (с удалением данных)
docker compose down -v
```

## ✨ Возможности

- ✅ CRUD операции с сообщениями
- ✅ Пагинация списка сообщений  
- ✅ Современный адаптивный UI
- ✅ CORS поддержка
- ✅ Автоматическая инициализация БД
- ✅ Проксирование через Nginx
- ✅ Health checks для PostgreSQL
- ✅ Логирование и обработка ошибок

## 🐛 Troubleshooting

### Порт занят
```bash
# Найти процесс
netstat -tulpn | grep :8080
# Убить процесс
kill -9 <PID>
```

### База данных недоступна
```bash
# Проверить состояние контейнеров
docker compose ps
# Проверить логи БД
docker compose logs db
```

### API не отвечает
```bash
# Логи API
docker compose logs api
# Перезапуск API
docker compose restart api
```

## 📝 Лицензия

MIT License

