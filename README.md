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
5. [Навыки (Skills)](#5-навыки-skills)
   - Список навыков
   - Управление навыками стажёра
   - Управление навыками стажировки
6. [Отклики на стажировки](#6-отклики-на-стажировки)
   - Откликнуться
   - Мои отклики
   - Отклики на стажировку (для компании)
   - Изменение статуса отклика
7. [Система токенов](#7-система-токенов)
8. [Rate Limiting](#8-rate-limiting)
9. [Обработка ошибок](#9-обработка-ошибок)

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

## 5. Навыки (Skills)

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

## 6. Отклики на стажировки

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

## 7. Система токенов

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

## 8. Rate Limiting

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

## 9. Обработка ошибок

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