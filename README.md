# 💬 Система чата с друзьями

Современная система обмена сообщениями с регистрацией пользователей, системой друзей и веб-интерфейсом.

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

## 👤 Тестовые данные

Для тестирования создается пользователь:
- **Email**: `test@example.com`
- **Пароль**: `password`
- **Имя**: Тест Юзер

Вы можете зарегистрировать свой аккаунт и добавить тестового пользователя в друзья для переписки.

## 📡 API Endpoints

### Аутентификация
| Метод | URL | Описание |
|-------|-----|----------|
| POST | `/api/auth/register` | Регистрация пользователя |
| POST | `/api/auth/login` | Вход в систему |
| POST | `/api/auth/logout` | Выход из системы |
| GET | `/api/auth/user` | Получить текущего пользователя |

### Система друзей
| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/friends/search?q=query` | Поиск пользователей |
| POST | `/api/friends/request` | Отправить заявку в друзья |
| POST | `/api/friends/accept` | Принять заявку в друзья |
| POST | `/api/friends/reject` | Отклонить заявку в друзья |
| GET | `/api/friends` | Список друзей |
| GET | `/api/friends/pending` | Входящие заявки |
| GET | `/api/friends/status?friend_id=N` | Статус дружбы |

### Сообщения
| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/messages` | Все сообщения (с пагинацией) |
| GET | `/api/messages/between?recipient_id=N` | Сообщения с конкретным пользователем |
| POST | `/api/messages/create` | Создать сообщение |
| GET | `/api/messages/get?id=N` | Получить сообщение |
| PUT | `/api/messages/update?id=N` | Обновить сообщение |
| DELETE | `/api/messages/delete?id=N` | Удалить сообщение |

### Общие
| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/` | Проверка API |
| GET | `/api/users` | Список друзей (контактов) |
| GET | `/api/testdb` | Тест подключения к БД |

### Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"first_name": "Иван", "last_name": "Петров", "email": "ivan@example.com", "password": "123456"}'

# Вход
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "ivan@example.com", "password": "123456"}' \
  -c cookies.txt

# Поиск пользователей
curl -H "Cookie: $(cat cookies.txt)" \
  "http://localhost:8080/api/friends/search?q=Тест"

# Отправить заявку в друзья
curl -X POST http://localhost:8080/api/friends/request \
  -H "Content-Type: application/json" \
  -H "Cookie: $(cat cookies.txt)" \
  -d '{"friend_id": 1}'

# Отправить сообщение
curl -X POST http://localhost:8080/api/messages/create \
  -H "Content-Type: application/json" \
  -H "Cookie: $(cat cookies.txt)" \
  -d '{"recipient_id": 1, "text": "Привет!"}'

# Получить переписку
curl -H "Cookie: $(cat cookies.txt)" \
  "http://localhost:8080/api/messages/between?recipient_id=1"
```

## 🏗 Архитектура

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────┐
│   Nginx Proxy   │    │   Go API     │    │ PostgreSQL  │
│   (Port 8080)   │◄──►│  (Port 8080) │◄──►│ users       │
│                 │    │ Auth + Chat  │    │ friends     │
│                 │    │ + Friends    │    │ messages    │
└─────────────────┘    └──────────────┘    └─────────────┘
         │
         ▼
┌─────────────────┐
│ Frontend SPA    │
│ Registration    │
│ Login + Chat    │
│ Friends Search  │
└─────────────────┘
```

### Компоненты:
- **Nginx**: Прокси-сервер для API и статических файлов
- **Go API**: Backend с аутентификацией, системой друзей и чатом
- **PostgreSQL**: База данных (users, friends, messages)
- **Frontend**: SPA с регистрацией, поиском друзей и чатом

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
│   │   ├── auth.go     # Аутентификация
│   │   ├── friends.go  # Система друзей
│   │   ├── messages.go # Сообщения
│   │   └── utils.go    # Общие утилиты
│   ├── models/          # Модели данных
│   ├── storage/         # Работа с БД
│   │   ├── users.go    # Пользователи
│   │   ├── friends.go  # Друзья
│   │   ├── messages.go # Сообщения
│   │   └── config.go   # Конфигурация БД
│   └── main.go         # Точка входа
├── front/              # Frontend
│   ├── index.html      # Чат приложение
│   ├── login.html      # Страница входа
│   ├── registration.html # Регистрация
│   └── source/         # Ресурсы (CSS, изображения)
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

### 👤 Пользователи
- ✅ Регистрация и аутентификация
- ✅ Cookie-based сессии
- ✅ Хеширование паролей (bcrypt)
- ✅ Валидация данных

### 👥 Система друзей
- ✅ Поиск пользователей по имени/email
- ✅ Отправка заявок в друзья
- ✅ Принятие/отклонение заявок
- ✅ Список друзей и входящих заявок

### 💬 Чат
- ✅ Приватные сообщения между друзьями
- ✅ История переписки
- ✅ Отправка сообщений в реальном времени
- ✅ Автообновление (каждые 3 сек)
- ✅ Красивые пузырьки сообщений

### 🖥️ Интерфейс
- ✅ Современный адаптивный дизайн
- ✅ Bootstrap + кастомные стили
- ✅ Модальные окна для поиска
- ✅ Уведомления о новых заявках
- ✅ Индикаторы загрузки

### 🔧 Техническое
- ✅ CRUD операции с сообщениями
- ✅ Пагинация списков
- ✅ CORS поддержка  
- ✅ Автоматическая инициализация БД
- ✅ Проксирование через Nginx
- ✅ Health checks для PostgreSQL
- ✅ Логирование и обработка ошибок
- ✅ Тестовые данные

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

