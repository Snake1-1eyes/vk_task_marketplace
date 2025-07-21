# Маркетплейс API

REST API сервис для маркетплейса с поддержкой авторизации, регистрации пользователей, управления объявлениями и просмотра ленты.

## Возможности

- Авторизация и регистрация пользователей
- Создание объявлений (только авторизованные пользователи)
- Просмотр ленты объявлений с сортировкой, фильтрацией и пагинацией
- REST API с поддержкой протокола gRPC
- Swagger UI для тестирования API

## Стек технологий

- **Язык**: Go
- **Фреймворк**: gRPC + gRPC Gateway
- **База данных**: PostgreSQL
- **Документация**: Swagger / OpenAPI
- **Контейнеризация**: Docker, Docker Compose
- **Миграции**: Goose

## Сборка и запуск

### Настройка переменных окружения

Перед запуском проекта можно настроить его, отредактировав файл `.env` (или создав его из `.env.example`)

Альтернативно, можно настроить приложение через файл `config/config.yml`

### Запуск через Docker Compose

1. Клонировать репозиторий:
   ```
   git clone [url-репозитория]
   cd vk_marketplace_task
   ```

2. Настроить параметры в файле `.env` или `config/config.yml` (при необходимости)

3. Запустить проект:
   ```
   docker-compose up -d
   ```

4. Сервис станет доступен по адресам:
   - gRPC: `localhost:50051`
   - HTTP REST API: `localhost:8080`
   - Swagger UI: `http://localhost:8080/swagger/`


## Тестирование API

### Swagger UI

Swagger UI доступен по следующим URL:
- Auth API: `http://localhost:8080/swagger/auth/`
- Listings API: `http://localhost:8080/swagger/listings/`

### API Endpoints

#### Авторизация и регистрация

**Регистрация нового пользователя**:
```
POST /v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

Ответ:
```json
{
  "user": {
    "id": "1",
    "username": "testuser",
    "created_at": "2025-07-21T10:30:15.123Z"
  }
}
```

**Авторизация пользователя**:
```
POST /v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "1",
    "username": "testuser",
    "created_at": "2025-07-21T10:30:15.123Z"
  }
}
```

#### Объявления

**Создание объявления** (требует авторизации):
```
POST /v1/listings
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "title": "Продам ноутбук",
  "description": "Новый ноутбук в отличном состоянии. Процессор Intel i7, 16GB RAM, 512GB SSD.",
  "image_url": "https://example.com/images/laptop.jpg",
  "price": 75000.50
}
```

Ответ:
```json
{
  "id": "1",
  "title": "Продам ноутбук",
  "description": "Новый ноутбук в отличном состоянии. Процессор Intel i7, 16GB RAM, 512GB SSD.",
  "image_url": "https://example.com/images/laptop.jpg",
  "price": 75000.5,
  "author_username": "testuser",
  "created_at": "2025-07-21T11:15:30.456Z",
  "is_owner": true
}
```

**Значения параметров сортировки и фильтрации**:

`sort_by` (тип сортировки):
- 0: SORT_FIELD_UNSPECIFIED (не указано)
- 1: SORT_FIELD_CREATED_AT (по дате создания)
- 2: SORT_FIELD_PRICE (по цене)

`sort_order` (направление сортировки):
- 0: SORT_ORDER_UNSPECIFIED (не указано)
- 1: SORT_ORDER_ASC (по возрастанию)
- 2: SORT_ORDER_DESC (по убыванию)

**Получение ленты объявлений**:
```
GET /v1/listings?page=1&per_page=10&sort_by=1&sort_order=2&min_price=10000&max_price=100000
```

Ответ:
```json
{
  "listings": [
    {
      "id": "2",
      "title": "Продам велосипед",
      "description": "Горный велосипед, 21 скорость, дисковые тормоза.",
      "image_url": "https://example.com/images/bike.jpg",
      "price": 15000,
      "author_username": "user2",
      "created_at": "2025-07-21T13:20:45.789Z",
      "is_owner": false
    },
    {
      "id": "1",
      "title": "Продам ноутбук",
      "description": "Новый ноутбук в отличном состоянии. Процессор Intel i7, 16GB RAM, 512GB SSD.",
      "image_url": "https://example.com/images/laptop.jpg",
      "price": 75000.5,
      "author_username": "testuser",
      "created_at": "2025-07-21T11:15:30.456Z",
      "is_owner": true
    }
  ],
  "total": 2,
  "page": 1,
  "per_page": 10,
  "total_pages": 1
}
```

## Реализация требований задачи

1. **Авторизация пользователя**:
   - Реализована через JWT-токены
   - Токен передается в заголовке Authorization
   - Реализована проверка токена на серверной стороне

2. **Регистрация пользователей**:
   - Проверка уникальности имени пользователя
   - Валидация формата имени пользователя и пароля
   - Безопасное хранение паролей с использованием bcrypt

3. **Размещение объявлений**:
   - Доступно только для авторизованных пользователей
   - Валидация заголовка, описания, URL изображения и цены
   - Привязка объявления к автору

4. **Лента объявлений**:
   - Постраничная навигация
   - Сортировка по дате и цене
   - Фильтрация по цене (мин/макс значения)
   - Для авторизованных пользователей отображается признак владения объявлением (is_owner)

## Возможные улучшения

1. **Технические улучшения**:
    - Unit и интеграционное тестирвоние(увы времени не хватило)
    - Кеширование (Redis) для популярных запросов
    - Добавление метрик и мониторинга (Prometheus/Grafana)
    - Расширенная система логирования и трейсинга (OpenTelemetry)
