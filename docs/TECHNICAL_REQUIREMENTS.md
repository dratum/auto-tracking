# Auto-Tracking: Техническое задание

## Обзор проекта

**Название**: Auto-Tracking
**Цель**: Система отслеживания автомобиля по GPS с просмотром маршрутов и статистики пробега
**Версия документа**: 1.0
**Дата**: 2025-02-11

---

## 1. Функциональные требования

### 1.1 MVP (Минимальный продукт)

| ID | Функция | Описание |
|----|---------|----------|
| F1 | Приём GPS-данных | Сервер принимает координаты от устройства (1 раз/сек) |
| F2 | Управление поездками | Автоматическое создание/завершение поездки по сигналу зажигания |
| F3 | История маршрутов | Список всех поездок с датой, временем, дистанцией |
| F4 | Просмотр на карте | Отображение маршрута выбранной поездки на карте |
| F5 | Статистика пробега | Общий пробег за день/неделю/месяц |
| F6 | Аутентификация | Защита API и веб-интерфейса |

### 1.2 Будущие функции (после MVP)

- Live-отслеживание на карте (WebSocket)
- Геозоны и уведомления
- Экспорт данных (GPX, CSV)
- Мобильное приложение
- Несколько автомобилей

---

## 2. Технический стек

### 2.1 Backend

| Компонент | Технология | Версия |
|-----------|------------|--------|
| Язык | Go | 1.21+ |
| HTTP Framework | Fiber или Chi | latest |
| БД временных рядов | TimescaleDB | 2.x |
| БД метаданных | MongoDB | 7.x |
| Контейнеризация | Docker + Docker Compose | latest |

### 2.2 Frontend

| Компонент | Технология | Версия |
|-----------|------------|--------|
| Framework | SvelteKit | 2.x |
| Карты | Leaflet | 1.9.x |
| HTTP клиент | fetch (native) | - |
| Стили | Tailwind CSS (опционально) | 3.x |

### 2.3 Железо (Трекер)

| Компонент | Модель | Примечание |
|-----------|--------|------------|
| Микроконтроллер | ESP32-WROOM-32 | WiFi для связи |
| GPS модуль | NEO-6M | UART, NMEA протокол |
| Питание | DC-DC преобразователь | 12V → 5V |
| Определение зажигания | Делитель напряжения | 12V → 3.3V на GPIO |

---

## 3. Архитектура системы

### 3.1 Общая схема

```
┌─────────────────────────────────────────────────────────────────┐
│                        Автомобиль                               │
│  ┌─────────┐    UART    ┌─────────────────────────────────┐    │
│  │ NEO-6M  │───────────►│           ESP32                 │    │
│  │  GPS    │            │  ┌─────────────────────────┐    │    │
│  └─────────┘            │  │ - Чтение NMEA           │    │    │
│                         │  │ - Парсинг координат     │    │    │
│  12V ──► DC-DC ────────►│  │ - HTTP POST на сервер   │    │    │
│                         │  │ - Логика зажигания      │    │    │
│  Зажигание ──► делитель►│  └─────────────────────────┘    │    │
│                         │              │ WiFi              │    │
│                         └──────────────┼──────────────────┘    │
└────────────────────────────────────────┼───────────────────────┘
                                         │
                                         ▼
                              ┌──────────────────┐
                              │  Хотспот (телефон)│
                              └────────┬─────────┘
                                       │ Internet
                                       ▼
┌──────────────────────────────────────────────────────────────────┐
│                           Сервер                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                      Go Backend                             │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │  │
│  │  │ Device API   │  │  Web API     │  │  Auth Middleware │  │  │
│  │  │ /api/device  │  │  /api/v1     │  │  JWT + API Key   │  │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────────────────┘  │  │
│  │         │                 │                                 │  │
│  │         ▼                 ▼                                 │  │
│  │  ┌─────────────────────────────────────────────────────┐   │  │
│  │  │                   Service Layer                      │   │  │
│  │  │  - TrackingService (обработка GPS)                  │   │  │
│  │  │  - TripService (управление поездками)               │   │  │
│  │  │  - StatsService (расчёт статистики)                 │   │  │
│  │  └─────────────────────────────────────────────────────┘   │  │
│  │         │                 │                                 │  │
│  │         ▼                 ▼                                 │  │
│  │  ┌─────────────┐   ┌─────────────┐                         │  │
│  │  │ TimescaleDB │   │  MongoDB    │                         │  │
│  │  │ GPS points  │   │  Trips,     │                         │  │
│  │  │             │   │  Vehicles,  │                         │  │
│  │  │             │   │  Users      │                         │  │
│  │  └─────────────┘   └─────────────┘                         │  │
│  └────────────────────────────────────────────────────────────┘  │
│                              ▲                                    │
│                              │                                    │
└──────────────────────────────┼────────────────────────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │   SvelteKit SPA     │
                    │  ┌───────────────┐  │
                    │  │ - Карта       │  │
                    │  │ - Маршруты    │  │
                    │  │ - Статистика  │  │
                    │  └───────────────┘  │
                    └─────────────────────┘
```

### 3.2 Структура проекта

```
auto-tracking/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP handlers
│   │   │   ├── device.go        # Приём данных от ESP32
│   │   │   ├── trip.go          # CRUD поездок
│   │   │   ├── stats.go         # Статистика
│   │   │   └── auth.go          # Логин
│   │   ├── middleware/          # Auth middleware
│   │   └── router.go            # Роутинг
│   ├── config/                  # Конфигурация
│   ├── domain/
│   │   ├── model/               # Модели данных
│   │   └── service/             # Бизнес-логика
│   └── repository/
│       ├── timescale/           # Репозиторий GPS-точек
│       └── mongo/               # Репозиторий поездок
├── web/                         # SvelteKit frontend
│   ├── src/
│   │   ├── routes/
│   │   ├── lib/
│   │   └── components/
│   └── package.json
├── firmware/                    # Код ESP32
│   ├── src/
│   │   └── main.cpp
│   └── platformio.ini
├── deployments/
│   ├── docker-compose.yml
│   └── Dockerfile
├── docs/
│   └── TECHNICAL_REQUIREMENTS.md
├── scripts/
├── go.mod
└── go.sum
```

---

## 4. API Спецификация

### 4.1 Device API (для ESP32)

**Аутентификация**: API-ключ в заголовке `X-API-Key`

#### POST /api/device/location
Отправка GPS-координат.

**Request:**
```json
{
  "lat": 55.7558,
  "lon": 37.6173,
  "speed": 45.5,
  "heading": 180.0,
  "satellites": 8,
  "timestamp": "2025-02-11T12:00:00Z"
}
```

**Response:** `201 Created`

#### POST /api/device/trip/start
Сигнал о начале поездки (зажигание ON).

**Response:**
```json
{
  "trip_id": "uuid-here"
}
```

#### POST /api/device/trip/end
Сигнал о завершении поездки (зажигание OFF).

**Response:** `200 OK`

---

### 4.2 Web API (для фронтенда)

**Аутентификация**: JWT токен в заголовке `Authorization: Bearer <token>`

#### POST /api/v1/auth/login
**Request:**
```json
{
  "username": "admin",
  "password": "secret"
}
```

**Response:**
```json
{
  "token": "jwt-token-here",
  "expires_at": "2025-02-18T12:00:00Z"
}
```

#### GET /api/v1/trips
Список поездок с пагинацией.

**Query params:** `?page=1&limit=20&from=2025-02-01&to=2025-02-11`

**Response:**
```json
{
  "trips": [
    {
      "id": "uuid",
      "start_time": "2025-02-11T08:00:00Z",
      "end_time": "2025-02-11T08:45:00Z",
      "distance_km": 23.5,
      "duration_min": 45
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

#### GET /api/v1/trips/:id
Детали поездки.

**Response:**
```json
{
  "id": "uuid",
  "start_time": "2025-02-11T08:00:00Z",
  "end_time": "2025-02-11T08:45:00Z",
  "distance_km": 23.5,
  "duration_min": 45,
  "max_speed": 85.0,
  "avg_speed": 31.3
}
```

#### GET /api/v1/trips/:id/points
GPS-точки маршрута.

**Response:**
```json
{
  "points": [
    {"lat": 55.7558, "lon": 37.6173, "speed": 0, "time": "2025-02-11T08:00:00Z"},
    {"lat": 55.7560, "lon": 37.6180, "speed": 15, "time": "2025-02-11T08:00:01Z"}
  ]
}
```

#### GET /api/v1/stats
Статистика пробега.

**Query params:** `?period=week` (day|week|month|year)

**Response:**
```json
{
  "period": "week",
  "total_distance_km": 245.8,
  "total_trips": 12,
  "total_duration_min": 480,
  "avg_trip_distance_km": 20.5
}
```

---

## 5. Модели данных

### 5.1 TimescaleDB: GPS Points

```sql
CREATE TABLE gps_points (
    time        TIMESTAMPTZ      NOT NULL,
    trip_id     UUID             NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lon         DOUBLE PRECISION NOT NULL,
    speed       REAL,            -- км/ч
    heading     REAL,            -- градусы 0-360
    satellites  SMALLINT
);

-- Преобразование в hypertable
SELECT create_hypertable('gps_points', 'time');

-- Индексы
CREATE INDEX idx_gps_points_trip_id ON gps_points (trip_id, time DESC);
```

### 5.2 MongoDB: Collections

**trips**
```json
{
  "_id": "ObjectId",
  "vehicle_id": "ObjectId",
  "start_time": "ISODate",
  "end_time": "ISODate",
  "distance_km": 23.5,
  "max_speed": 85.0,
  "avg_speed": 31.3,
  "status": "completed",  // "active" | "completed"
  "created_at": "ISODate"
}
```

**vehicles**
```json
{
  "_id": "ObjectId",
  "name": "Toyota Camry",
  "plate_number": "A123BC",
  "created_at": "ISODate"
}
```

**users**
```json
{
  "_id": "ObjectId",
  "username": "admin",
  "password_hash": "bcrypt-hash",
  "created_at": "ISODate"
}
```

---

## 6. Аутентификация и безопасность

### 6.1 ESP32 → Backend

- **Метод**: Статический API-ключ
- **Заголовок**: `X-API-Key: <key>`
- **Хранение ключа**: Переменная окружения на сервере, константа в прошивке ESP32
- **Валидация**: Middleware проверяет ключ на всех `/api/device/*` эндпоинтах

### 6.2 Frontend → Backend

- **Метод**: JWT токен
- **Время жизни**: 7 дней
- **Хранение**: localStorage на клиенте
- **Заголовок**: `Authorization: Bearer <token>`
- **Алгоритм**: HS256

### 6.3 Безопасность

- HTTPS обязателен в продакшене
- Пароль хешируется bcrypt
- Rate limiting для /api/device (защита от флуда)
- CORS настроен только на домен фронтенда

---

## 7. Схема железа ESP32

### 7.1 Подключение компонентов

```
                    ┌─────────────────────────────────────┐
                    │            ESP32                    │
                    │                                     │
  NEO-6M            │                                     │
  ┌──────┐          │                                     │
  │ VCC  │──────────┼── 3.3V                              │
  │ GND  │──────────┼── GND                               │
  │ TX   │──────────┼── GPIO16 (RX2)                      │
  │ RX   │──────────┼── GPIO17 (TX2)                      │
  └──────┘          │                                     │
                    │                                     │
  Питание 12V       │                                     │
  ┌──────────┐      │                                     │
  │ 12V IN   │      │                                     │
  │    │     │      │                                     │
  │ DC-DC    │      │                                     │
  │ (LM2596) │      │                                     │
  │    │     │      │                                     │
  │ 5V OUT ──┼──────┼── VIN                               │
  │ GND ─────┼──────┼── GND                               │
  └──────────┘      │                                     │
                    │                                     │
  Зажигание (ACC)   │                                     │
  ┌──────────────┐  │                                     │
  │ 12V ─── R1 ──┼──┼── GPIO34 (INPUT)                    │
  │       (10K)  │  │      ▲                              │
  │         │    │  │      │                              │
  │       R2 ────┼──┼── GND                               │
  │      (4.7K)  │  │                                     │
  └──────────────┘  │  Делитель: 12V * 4.7/(10+4.7) ≈ 3.8V│
                    │  (используем с резистором на 3.3V)  │
                    └─────────────────────────────────────┘
```

### 7.2 Список компонентов

| Компонент | Количество | Примечание |
|-----------|------------|------------|
| ESP32-WROOM-32 DevKit | 1 | Любой вариант с USB |
| GPS NEO-6M | 1 | С антенной |
| DC-DC LM2596 | 1 | Вход 12V, выход 5V |
| Резистор 10K | 1 | Для делителя напряжения |
| Резистор 4.7K | 1 | Для делителя напряжения |
| Провода | - | Для подключения |
| Корпус | 1 | Опционально |

---

## 8. План реализации

### Этап 1: Инфраструктура
- [ ] Docker-compose с TimescaleDB + MongoDB
- [ ] Базовая структура Go-проекта
- [ ] Конфигурация (env файлы)
- [ ] Миграции БД

### Этап 2: Backend Core
- [ ] Модели данных
- [ ] Репозитории (TimescaleDB, MongoDB)
- [ ] Device API: приём GPS-точек
- [ ] Device API: старт/стоп поездки
- [ ] Расчёт дистанции по точкам

### Этап 3: Аутентификация
- [ ] Middleware для API-ключа
- [ ] Login endpoint + JWT генерация
- [ ] Middleware для JWT проверки

### Этап 4: Web API
- [ ] GET /trips (список)
- [ ] GET /trips/:id (детали)
- [ ] GET /trips/:id/points (маршрут)
- [ ] GET /stats (статистика)

### Этап 5: Frontend
- [ ] Инициализация SvelteKit проекта
- [ ] Страница логина
- [ ] Страница списка поездок
- [ ] Страница с картой маршрута
- [ ] Страница статистики

### Этап 6: Firmware ESP32
- [ ] Настройка PlatformIO проекта
- [ ] Чтение NMEA с NEO-6M
- [ ] Парсинг координат
- [ ] WiFi подключение
- [ ] HTTP отправка на сервер
- [ ] Логика зажигания (старт/стоп)

### Этап 7: Интеграция и тестирование
- [ ] End-to-end тест с эмулятором GPS
- [ ] Тест с реальным железом
- [ ] Оптимизация запросов

### Этап 8: Деплой
- [ ] Dockerfile для backend
- [ ] Настройка VPS
- [ ] HTTPS (Let's Encrypt)
- [ ] CI/CD (опционально)

---

## 9. Конфигурация окружения

### .env (пример)

```env
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# TimescaleDB
TIMESCALE_HOST=localhost
TIMESCALE_PORT=5432
TIMESCALE_USER=autotrack
TIMESCALE_PASSWORD=secret
TIMESCALE_DB=autotrack

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DB=autotrack

# Auth
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=168h
API_KEY=your-device-api-key

# App
DEFAULT_USERNAME=admin
DEFAULT_PASSWORD_HASH=$2a$10$...
```

---

## 10. Критерии готовности MVP

- [ ] ESP32 отправляет координаты на сервер
- [ ] Поездки создаются/завершаются по зажиганию
- [ ] Веб-интерфейс показывает список поездок
- [ ] Маршрут отображается на карте
- [ ] Статистика пробега за период корректна
- [ ] Аутентификация работает
