# API Системы Поддержки

Бэкенд-сервис для системы обработки тикетов поддержки, написанный на Go с использованием Gin framework.

## Основные возможности

- 🎫 Управление тикетами (создание, просмотр, обновление, удаление)
- 👥 Управление пользователями
- 💬 Обмен сообщениями в тикетах
- 📸 Поддержка вложений (фотографий)
- 📱 REST API для интеграции с фронтендом
- 🔔 Уведомления о статусе тикетов

## Технический стек

- **Язык**: Go 1.x
- **Web Framework**: Gin
- **База данных**: PostgreSQL
- **Логирование**: Встроенный logger
- **Конфигурация**: JSON-based

## API Endpoints

### Тикеты

| Метод | Endpoint | Описание | Параметры запроса | Тело запроса |
|-------|----------|----------|------------------|--------------|
| GET | `/api/tickets` | Получение списка тикетов | `page`: номер страницы<br>`limit`: количество записей<br>`status`: фильтр по статусу | - |
| GET | `/api/tickets/:id` | Получение информации о тикете | `id`: ID тикета | - |
| POST | `/api/tickets` | Создание нового тикета | - | ```json<br>{<br>  "user_id": 123,<br>  "title": "Название",<br>  "description": "Описание",<br>  "category": "Категория"<br>}``` |
| PUT | `/api/tickets/:id` | Обновление тикета | `id`: ID тикета | ```json<br>{<br>  "status": "статус",<br>  "category": "категория"<br>}``` |
| DELETE | `/api/tickets/:id` | Удаление тикета | `id`: ID тикета | - |

### Сообщения тикетов

| Метод | Endpoint | Описание | Параметры запроса | Тело запроса |
|-------|----------|----------|------------------|--------------|
| POST | `/api/tickets/:id/messages` | Добавление сообщения | `id`: ID тикета | ```json<br>{<br>  "sender_type": "user/support",<br>  "sender_id": 123,<br>  "message": "Текст"<br>}``` |
| GET | `/api/tickets/:id/messages` | Получение сообщений | `id`: ID тикета | - |

### Фотографии тикетов

| Метод | Endpoint | Описание | Параметры запроса | Тело запроса |
|-------|----------|----------|------------------|--------------|
| POST | `/api/tickets/:id/photos` | Загрузка фотографии | `id`: ID тикета | Multipart form:<br>`photo`: файл<br>`sender_type`: тип<br>`sender_id`: ID<br>`message_id`: ID сообщения |
| GET | `/api/tickets/photos/:photo_id` | Получение фотографии | `photo_id`: ID фото | - |
| DELETE | `/api/tickets/photos/:photo_id` | Удаление фотографии | `photo_id`: ID фото | - |

### Пользователи

| Метод | Endpoint | Описание | Параметры запроса | Тело запроса |
|-------|----------|----------|------------------|--------------|
| GET | `/api/users` | Получение списка пользователей | `page`: номер страницы<br>`limit`: количество записей | - |
| GET | `/api/users/:id` | Получение информации о пользователе | `id`: ID пользователя | - |
| POST | `/api/users` | Создание пользователя | - | ```json<br>{<br>  "id": 123,<br>  "full_name": "Имя",<br>  "phone": "Телефон",<br>  "location_lat": 55.123,<br>  "location_lng": 37.123<br>}``` |
| PUT | `/api/users/:id` | Обновление пользователя | `id`: ID пользователя | ```json<br>{<br>  "full_name": "Имя",<br>  "phone": "Телефон"<br>}``` |

## Структура проекта

```
.
├── config/         # Конфигурация приложения
├── db/            # Работа с базой данных
├── handlers/      # Обработчики HTTP запросов
├── logger/        # Логирование
├── models/        # Модели данных
├── uploads/       # Директория для загруженных файлов
└── main.go        # Точка входа в приложение
```

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone [URL репозитория]
cd support_front_api
```

2. Создайте файл конфигурации `config/config.json`:
```json
{
  "port": "8080",
  "database_url": "postgres://postgres:postgres@localhost:5432/support_tickets?sslmode=disable",
  "jwt_secret": "your-secret-key",
  "log_file_path": "logs/app.log",
  "allow_origins": ["*"]
}
```

3. Установите зависимости:
```bash
go mod download
```

4. Запустите приложение:
```bash
go run main.go
```

## Особенности реализации

- **Пагинация**: Все списки (тикеты, пользователи) поддерживают пагинацию
- **Фильтрация**: Поддержка фильтрации тикетов по статусу
- **Транзакции**: Использование транзакций для обеспечения целостности данных
- **CORS**: Настраиваемая CORS политика
- **Логирование**: Детальное логирование всех операций
- **Обработка ошибок**: Централизованная обработка и логирование ошибок

## Модели данных

### Тикет
```go
type Ticket struct {
    ID           int        `json:"id"`
    UserID       int64      `json:"user_id"`
    UserFullName string     `json:"user_full_name"`
    Title        string     `json:"title"`
    Description  string     `json:"description"`
    Status       string     `json:"status"`
    Category     string     `json:"category"`
    CreatedAt    time.Time  `json:"created_at"`
    ClosedAt     *time.Time `json:"closed_at,omitempty"`
}
```

### Пользователь
```go
type User struct {
    ID           int64      `json:"id"`
    FullName     string     `json:"full_name"`
    Phone        string     `json:"phone"`
    LocationLat  float64    `json:"location_lat"`
    LocationLng  float64    `json:"location_lng"`
    BirthDate    time.Time  `json:"birth_date"`
    IsRegistered bool       `json:"is_registered"`
    RegisteredAt *time.Time `json:"registered_at"`
}
```

## Безопасность

- Валидация входных данных
- Безопасная работа с файлами
- Настраиваемые CORS политики
- Поддержка JWT для авторизации (подготовлено для реализации)

## Требования к системе

- Go 1.x
- PostgreSQL 12+
- Минимум 512MB RAM
- 1GB свободного места на диске

## Лицензия

MIT 
