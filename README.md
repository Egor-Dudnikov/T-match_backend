# 📚 T-match Backend API — Полная документация

## Оглавление

1. [Общая информация](#1-общая-информация)
2. [Аутентификация и регистрация](#2-аутентификация-и-регистрация)
   - Регистрация студента
   - Регистрация компании
   - Подтверждение email
   - Запрос нового кода
   - Вход в систему
3. [Профиль пользователя](#3-профиль-пользователя)
   - Профиль студента (чтение/обновление)
   - Профиль компании (чтение/обновление)
   - Аватар
4. [Стажировки](#4-стажировки)
   - Создание стажировки
   - Получение стажировки по ID
   - Обновление стажировки
   - Архивирование стажировки
5. [Поиск и фильтрация](#5-поиск-и-фильтрация)
   - Поиск стажировок
   - Поиск интернов
   - Поиск компаний
   - Примечания по поиску
6. [Навыки (Skills)](#6-навыки-skills)
   - Список навыков
   - Управление навыками стажёра
   - Управление навыками стажировки
7. [Отклики на стажировки](#7-отклики-на-стажировки)
   - Откликнуться
   - Мои отклики
   - Отклики на стажировку (для компании)
   - Изменение статуса отклика
8. [Система токенов](#8-система-токенов)
9. [Rate Limiting](#9-rate-limiting)
10. [Обработка ошибок](#10-обработка-ошибок)
11. [Запуск приложения](#11-запуск-приложения)
    - Требования
    - Настройка окружения
    - Запуск Production
    - Запуск Development
    - Переключение между режимами

---

## 1. Общая информация

| Параметр | Значение |
| :--- | :--- |
| **Базовый URL** | `http://localhost:8080` |
| **Формат данных** | `application/json` (кроме загрузки аватара) |
| **Аутентификация** | Заголовок `Token: <access_token>` |
| **Cookies** | `refresh_token` (HttpOnly, для обновления токенов) |

---

## 2. Аутентификация и регистрация

### Регистрация студента

```
POST /auth/students
```

**Тело запроса:**

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `email` | `string` | ✅ | Email (макс. 255 символов) |
| `password` | `string` | ✅ | Пароль (8-72 символа, A-Z, a-z, 0-9) |
| `device_id` | `string` | ✅ | ID устройства (5-100 символов) |
| `birth_date` | `string` | ✅ | Дата рождения (ISO 8601, возраст ≥ 16 лет) |

**Пример запроса:**
```json
{
  "email": "ivan@example.com",
  "password": "StrongPass1",
  "device_id": "web_chrome_abc123",
  "birth_date": "2008-05-15T00:00:00Z"
}
```

**Ответы:**

| Статус | Заголовки | Описание |
| :--- | :--- | :--- |
| `200 OK` | `Token: <session_id>` | Успех. Session ID (7 минут) |
| `400 Bad Request` | - | Ошибка валидации |
| `409 Conflict` | - | Email уже существует |
| `422 Unprocessable Entity` | - | Возраст < 16 лет |

---

### Регистрация компании

```
POST /auth/company
```

**Тело запроса:**

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `inn` | `string` | ✅ | ИНН (10 или 12 цифр) |
| `email` | `string` | ✅ | Email (макс. 255 символов) |
| `password` | `string` | ✅ | Пароль (те же требования) |
| `device_id` | `string` | ✅ | ID устройства |

**Пример запроса:**
```json
{
  "inn": "7707083893",
  "email": "hr@company.ru",
  "password": "SecurePass456!",
  "device_id": "iphone_app_xyz"
}
```

**Ответы:**

| Статус | Заголовки | Описание |
| :--- | :--- | :--- |
| `200 OK` | `Token: <session_id>` | Успех |
| `400 Bad Request` | - | Ошибка валидации |
| `404 Not Found` | - | Компания не найдена или неактивна |
| `409 Conflict` | - | Email уже существует |
| `502 Bad Gateway` | - | Ошибка DaData API |

---

### Подтверждение email

```
POST /auth/students/verify   # студент
POST /auth/company/verify    # компания
```

**Заголовки:** `Token: <session_id>`

**Тело запроса:**
```json
{
  "code": "482915"
}
```

**Ответы:**

| Статус | Заголовки | Описание |
| :--- | :--- | :--- |
| `200 OK` | `Token: <access_token>`<br>`Set-Cookie: refresh_token=...` | Успех |
| `400 Bad Request` | - | Неверный код или формат |
| `400 Bad Request` | - | Код истёк |
| `429 Too Many Requests` | - | Слишком много попыток |

---

### Запрос нового кода

```
POST /auth/newverify
```

**Заголовки:** `Token: <session_id>`

**Ответы:**

| Статус | Описание |
| :--- | :--- |
| `200 OK` | Новый код отправлен |
| `500 Internal Server Error` | Ошибка сервера |

---

### Вход в систему

```
POST /auth/students/login   # студент
POST /auth/company/login    # компания
```

**Тело запроса:**
```json
{
  "email": "ivan@example.com",
  "password": "StrongPass1",
  "device_id": "web_chrome_abc123"
}
```

**Ответы:**

| Статус | Заголовки | Описание |
| :--- | :--- | :--- |
| `200 OK` | `Token: <access_token>`<br>`Set-Cookie: refresh_token=...` | Успех |
| `401 Unauthorized` | - | Неверный пароль |
| `404 Not Found` | - | Email не найден |
| `429 Too Many Requests` | - | Превышен лимит попыток |

---

## 3. Профиль пользователя

*Все эндпоинты требуют Access Token в заголовке `Token`.*

---

### Профиль студента

#### Чтение профиля

```
GET /my/profile
```
*Роль: `intern`*

**Пример ответа:**
```json
{
  "email": "ivan@example.com",
  "profile": {
    "first_name": "Иван",
    "last_name": "Петров",
    "birth_date": "2008-05-15T00:00:00Z",
    "location": "Москва",
    "university": "МГУ",
    "degree": "Бакалавр",
    "bio": "Студент 4-го курса.",
    "experience": null,
    "image": null
  },
  "skills": [
    {"id": 1, "name": "Go"},
    {"id": 2, "name": "PostgreSQL"}
  ]
}
```

#### Обновление профиля

```
PUT /my/profile/put
```
*Роль: `intern`*

**Тело запроса** (все поля опциональны):

| Параметр | Тип | Описание |
| :--- | :--- | :--- |
| `first_name` | `string` | Имя (макс. 100) |
| `last_name` | `string` | Фамилия (макс. 100) |
| `birth_date` | `string` | Дата рождения (≥16 лет) |
| `location` | `string` | Город (макс. 200) |
| `university` | `string` | Университет (макс. 200) |
| `degree` | `string` | Степень (макс. 100) |
| `bio` | `string` | О себе (макс. 2000) |
| `experience` | `string` | Опыт (макс. 5000) |

---

### Профиль компании

#### Чтение профиля

```
GET /my/company/profile
```
*Роль: `company`*

#### Обновление профиля

```
PUT /my/company/profile/put
```
*Роль: `company`*

**Тело запроса:**
```json
{
  "company_name": "ООО Ромашка",
  "description": "IT-компания",
  "website": "https://romashka.ru"
}
```

---

### Аватар

```
PUT /my/avatar/put
```
*Content-Type: `multipart/form-data`*

| Поле | Тип | Описание |
| :--- | :--- | :--- |
| `avatar` | `File` | Изображение (≤10 МБ, JPEG/PNG) |

**Ответ:** `200 OK` — возвращает URL аватара

```json
"http://localhost:9000/t-match-storge/user_17_avatar"
```

---

## 4. Стажировки

### Создание стажировки

```
POST /internships
```
*Роль: `company`*

**Тело запроса:**
```json
{
  "title": "Стажер Go",
  "description": "Ищем мотивированного стажёра...",
  "salary": 50000,
  "location": "Москва",
  "duration_month": 3
}
```

**Ответ:** `200 OK`

---

### Получение стажировки по ID

```
GET /internships/:id
```
*Публичный эндпоинт (без авторизации)*

**Пример ответа:**
```json
{
  "id": 15,
  "company_id": 5,
  "title": "Стажер Go",
  "description": "Ищем мотивированного стажёра...",
  "salary": 50000,
  "location": "Москва",
  "is_archived": false,
  "duration_month": 3,
  "created_at": "2025-05-09T10:30:00Z"
}
```

---

### Обновление стажировки

```
PUT /internships/update/:id
```
*Роль: `company` (только свои)*

**Тело запроса** (все поля опциональны):
```json
{
  "title": "Новое название",
  "salary": 60000
}
```

---

### Архивирование стажировки

```
DELETE /internships/delete/:id
```
*Роль: `company` (только свои)*

**Ответ:** `200 OK`

---

## 5. Поиск и фильтрация

### Поиск стажировок

```
GET /internships
```
*Публичный эндпоинт (без авторизации)*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | - | Текстовый поиск по title, description, company_name, location. Слова разделяются пробелом, логика OR (достаточно совпадения по одному слову) |
| `location` | `string` | - | Город (частичное совпадение, регистронезависимый) |
| `salary_min` | `int` | - | Минимальная зарплата |
| `salary_max` | `int` | - | Максимальная зарплата |
| `duration_min` | `int` | - | Минимальная длительность (месяцы) |
| `duration_max` | `int` | - | Максимальная длительность (месяцы) |
| `skills` | `[]int` | - | ID навыков (AND-логика: все должны быть у стажировки). Передавать несколько раз: `?skills=1&skills=2` |
| `sort` | `string` | `created_at` | Поле сортировки: `salary`, `duration_months`, `created_at` |
| `order` | `string` | `desc` | Направление: `asc` или `desc` (1 для asc, 0 для desc) |
| `limit` | `int` | 20 | Записей на странице |
| `offset` | `int` | 0 | Сдвиг для пагинации |

**Примеры запросов:**

```bash
# Текстовый поиск
GET /internships?query=Go+разработчик

# Поиск с фильтрами
GET /internships?query=Python&location=Москва&salary_min=50000&salary_max=150000

# Поиск по навыкам
GET /internships?skills=1&skills=3&skills=5

# Сортировка и пагинация
GET /internships?query=backend&sort=salary&order=desc&limit=20&offset=0

# Только фильтры без текстового поиска
GET /internships?location=Казань&duration_max=6
```

**Пример ответа:**
```json
[
  {
    "id": 15,
    "company_id": 5,
    "title": "Go разработчик",
    "description": "Ищем начинающего Go разработчика...",
    "salary": 80000,
    "location": "Москва",
    "is_archived": false,
    "duration_month": 6,
    "created_at": "2026-05-15T10:30:00Z"
  }
]
```

**Логика сортировки:**
1. Пользовательское поле (если указано)
2. Релевантность полнотекстового поиска (если есть `query`)
3. Дата создания (всегда, для стабильности)

---

### Поиск интернов (стажёров)

```
GET /students
```
*Публичный эндпоинт (без авторизации)*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | - | Текстовый поиск по first_name, last_name, bio, experience, university, degree |
| `university` | `string` | - | Университет (частичное совпадение) |
| `skills` | `[]int` | - | ID навыков (AND-логика). Передавать несколько раз: `?skills=1&skills=2` |
| `limit` | `int` | 20 | Записей на странице |
| `offset` | `int` | 0 | Сдвиг для пагинации |

**Примеры запросов:**

```bash
# Поиск по имени
GET /students?query=Иван

# Поиск с фильтрами
GET /students?query=Python&university=МГУ&skills=1&skills=2

# Пагинация
GET /students?limit=10&offset=20
```

**Пример ответа:**
```json
[
  {
    "id": 17,
    "first_name": "Иван",
    "last_name": "Петров",
    "birth_date": "2002-05-15T00:00:00Z",
    "location": "Москва",
    "university": "МГУ",
    "degree": "Computer Science",
    "bio": "Студент 4 курса. Увлекаюсь backend-разработкой на Go и Python.",
    "experience": "Писал курсовые на Go, участвовал в хакатонах.",
    "image": null
  }
]
```

---

### Поиск компаний

```
GET /companies
```
*Публичный эндпоинт (без авторизации)*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | - | Текстовый поиск по company_name, description, legal_address |
| `location` | `string` | - | Юридический адрес (частичное совпадение) |
| `limit` | `int` | 20 | Записей на странице |
| `offset` | `int` | 0 | Сдвиг для пагинации |

**Примеры запросов:**

```bash
# Поиск по названию
GET /companies?query=Яндекс

# Поиск по адресу
GET /companies?location=Москва

# Комбинация
GET /companies?query=IT&location=Москва&limit=10
```

**Пример ответа:**
```json
[
  {
    "id": 5,
    "company_name": "Яндекс",
    "description": "Крупная IT-компания, разработка поисковых технологий",
    "website": "https://ya.ru",
    "inn": "7707083893",
    "kpp": "770701001",
    "ogrn": "1027700229193",
    "legal_address": "Москва, ул. Льва Толстого, 16",
    "director_name": "Иванов Иван Иванович",
    "image": null
  }
]
```

---

## Примечания по поиску

**Полнотекстовый поиск (FTS):**
- Поддерживает словоформы русского языка ("разработчика" найдёт "разработчик")
- Слова в запросе соединяются через OR (достаточно совпадения по одному слову)
- Результаты сортируются по релевантности (совпадение в названии важнее, чем в описании)

**Фильтрация по навыкам:**
- Используется AND-логика: интерн или стажировка должны иметь ВСЕ указанные навыки
- ID навыков можно получить через `GET /skills`

**Пагинация:**
- Если `limit` не указан, по умолчанию возвращается 20 записей
- Для получения следующей страницы: `offset = предыдущий_offset + limit`

**Пустой запрос:**
- Если `query` не указан, возвращаются все записи с учётом остальных фильтров

---

## 6. Навыки (Skills)

### Список всех навыков

```
GET /skills
```

**Пример ответа:**
```json
[
  {"id": 1, "name": "Python"},
  {"id": 2, "name": "Go"},
  {"id": 3, "name": "PostgreSQL"},
  {"id": 4, "name": "Docker"}
]
```

---

### Управление навыками стажёра

#### Добавить навыки

```
POST /my/profile/skills/add
```
*Роль: `intern`*

**Тело запроса:**
```json
{
  "skill_id": [1, 2, 5]
}
```

#### Удалить навыки

```
DELETE /my/profile/skills/delete
```
*Роль: `intern`*

**Тело запроса:**
```json
{
  "skill_id": [2, 5]
}
```

---

### Управление навыками стажировки

#### Добавить навыки

```
POST /internship/:id/skill/add
```
*Роль: `company` (владелец)*

**Тело запроса:**
```json
{
  "skill_id": [1, 2, 3]
}
```

#### Удалить навыки

```
DELETE /internship/:id/skill/delete
```
*Роль: `company` (владелец)*

---

## 7. Отклики на стажировки

### Откликнуться

```
POST /internships/:id/respond
```
*Роль: `intern`*

**Ответы:**

| Статус | Описание |
| :--- | :--- |
| `200 OK` | Отклик создан (статус `pending`) |
| `409 Conflict` | Уже откликался |
| `410 Gone` | Стажировка архивирована |

---

### Мои отклики (стажёр)

```
GET /my/responses
```
*Роль: `intern`*

**Пример ответа:**
```json
[
  {
    "id": 101,
    "intern_id": 17,
    "internship_id": 42,
    "status": "pending",
    "created_at": "2025-05-09T10:30:00Z"
  }
]
```

**Возможные статусы:**

| Статус | Описание |
| :--- | :--- |
| `pending` | Ожидает рассмотрения |
| `reviewing` | В рассмотрении |
| `accepted` | Принят |
| `rejected` | Отклонён |

---

### Отклики на стажировку (компания)

```
GET /internships/:id/responses
```
*Роль: `company` (владелец)*

**Пример ответа:** (аналогично моим откликам)

---

### Изменение статуса отклика

```
PUT /responses/:id/status
```
*Роль: `company` (владелец стажировки)*

**Тело запроса:**
```json
{
  "status": "accepted"
}
```

---

## 8. Система токенов

| Токен | Передача | TTL | Назначение |
| :--- | :--- | :--- | :--- |
| **Session ID** | `Token` header | 7 минут | Верификация email |
| **Access Token** | `Token` header | 15 минут | Авторизация API |
| **Refresh Token** | `HttpOnly` cookie | 7 дней | Обновление Access Token |

**Механизм обновления:**
1. Запрос с истёкшим Access Token
2. Сервер проверяет Refresh Token из cookie
3. Если валиден — генерирует новый Access Token
4. Возвращает его в заголовке `Token`

---

## 9. Rate Limiting

Лимиты в **запросах в минуту**:

| Эндпоинт | Лимит |
| :--- | :--- |
| `POST /auth/students` | 20 |
| `POST /auth/students/verify` | 60 |
| `POST /auth/newverify` | 7 |
| `POST /auth/students/login` | 30 |
| `POST /auth/company` | 20 |
| `POST /auth/company/verify` | 60 |
| `POST /auth/company/login` | 30 |
| `PUT /my/profile/put` | 100 |
| `GET /my/profile` | 120 |
| `PUT /my/avatar/put` | 100 |
| `POST /internships` | 12 |
| `PUT /internships/update/:id` | 12 |
| `DELETE /internships/delete/:id` | 5 |
| `POST /internships/:id/respond` | 10 |
| `GET /my/responses` | 20 |

`GET /internships/:id` — **без лимита**

---

## 10. Обработка ошибок

| Код | Сообщение | Причина |
| :--- | :--- | :--- |
| `400` | `Bad request` | Невалидный JSON или данные |
| `401` | `User Unauthorized` | Нет токена или он невалиден |
| `403` | `Access denied` | Недостаточно прав |
| `404` | `Not found` | Ресурс не найден |
| `409` | `Already exists` | Конфликт (email, отклик) |
| `410` | `Internship is archived` | Стажировка в архиве |
| `422` | `User must be at least 16 years old` | Возраст < 16 |
| `429` | `Too many requests` | Превышен rate limit |
| `500` | `Internal server error` | Ошибка сервера |
| `502` | `External service error` | Ошибка DaData API |
| `503` | `Service unavailable` | Redis или email недоступны |

---


## 11. Запуск приложения

### Требования

| Компонент | Версия |
| :--- | :--- |
| **Docker** | ≥ 20.10 |
| **Docker Compose** | ≥ 1.29 |
| **Go** (только для dev) | ≥ 1.22 |

---

### Настройка окружения

#### 1. Клонирование репозитория
```bash
git clone https://github.com/Egor-Dudnikov/T-match_backend.git
cd T-match_backend
```

#### 2. Создание .env файла

Переименуйте `.env.example` в `.env` и отредактируйте:

```bash
mv .env.example .env
nano .env
```

**Содержимое `.env`:**
```env
# JWT (обязательно сгенерируйте свой)
JWT_SECRET=your_jwt_secret_32chars_minimum

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=t_match_database
DB_USER=postgres
DB_PASSWORD=your_db_password
DB_SSLMODE=disable

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=:8080

# Redis
REDIS_ADDR=redis:6379
REDIS_DB=0
REDIS_PASSWORD=
REDIS_MAX_RETRIES=3
REDIS_DIAL_TIMEOUT=5000000000
REDIS_TIMEOUT=3000000000

# Email 
# По умолчанию использует MailHog, для тестирования и разработки
EMAIL_ADDR=mailhog:1025
EMAIL_HOST=mailhog
EMAIL_IDENTITY=
EMAIL_USERNAME=noreply@tmatch.space
EMAIL_PASSWORD=

# CORS
CORS_ALLOW_ORIGIN=http://localhost:8000
CORS_ALLOW_HEADERS=Content-Type,Token

# S3/MinIO
S3_ENDPOINT=minio:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_USE_SSL=false

# DaData (обязательно получите свой ключ)
DA_DATA_API_KEY=your_dadata_api_key

# Config
CONFIG_PATH=./configs/configuration.json
```

#### 3. Где взять ключи

| Переменная | Где получить |
| :--- | :--- |
| `DA_DATA_API_KEY` | Зарегистрироваться на [dadata.ru](https://dadata.ru) → API → Ключи доступа |
| `JWT_SECRET` | Сгенерировать: `openssl rand -base64 32` |
| `DB_PASSWORD` | Придумать самостоятельно |

#### Примечания:

Email: если EMAIL_PASSWORD не задан, автоматически используется MailHog — все письма перехватываются и отображаются в веб-интерфейсе http://localhost:8025, реальным пользователям ничего не отправляется

---

### Запуск Production

Полная пересборка образа и запуск всех сервисов:

```bash
docker compose up -d --build
```

**Проверка статуса:**
```bash
docker compose ps
docker compose logs app
```

**Остановка:**
```bash
docker compose down
```

---

### Запуск Development

Для быстрой разработки без полной пересборки образа при каждом изменении кода.

#### 1. Собрать бинарник на хосте

```bash
CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go
```

#### 2. Запустить dev-окружение

```bash
docker compose -f docker-compose.dev.yml up -d
```

Базы данных (PostgreSQL, Redis, MinIO) запускаются в Docker, а приложение использует бинарник, собранный на хосте и проброшенный через volume.

#### 3. При изменениях в коде

```bash
# Пересобрать бинарник
CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Перезапустить только приложение (базы не трогаются)
docker compose -f docker-compose.dev.yml restart app
```

#### 4. Переключение между режимами

```bash
# Из development в production:
docker compose -f docker-compose.dev.yml stop app
docker compose up -d app

# Из production в development:
docker compose stop app
docker compose -f docker-compose.dev.yml up -d
```

Базы данных можно не перезапускать — они общие для обоих режимов.

#### 5. Отличия конфигураций

| Параметр | Production | Development |
| :--- | :--- | :--- |
| **Конфиг-файл** | `docker-compose.yml` | `docker-compose.dev.yml` |
| **Dockerfile** | `Dockerfile` (многоэтапная сборка) | Готовый бинарник с хоста |
| **Образ приложения** | `alpine:3.19` + бинарник внутри | `alpine:3.19`, бинарник через volume |
| **При изменении кода** | `docker compose up -d --build` (долго) | `go build` + `docker compose restart` (2 секунды) |
| **volumes** | Нет | `./main:/app/main` |

---

## License

MIT License — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Egor Dudnikov
