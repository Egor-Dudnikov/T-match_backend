## T-match Backend API Documentation

### Базовый URL

```
Development: http://localhost:8080
```

---

## 🔐 Аутентификация

### 1. Регистрация студента

**Отправка email, пароля, device_id и даты рождения → получение Session ID для верификации.**

```
POST /auth/students
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
```

**Body:**
```json
{
  "email": "student@example.com",
  "password": "SecurePass123!",
  "device_id": "web_chrome_123",
  "birth_date": "2008-05-15T00:00:00Z"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| email | string | ✅ | Email пользователя (max 255 символов) |
| password | string | ✅ | Пароль: 8-72 символов, **обязательно**: A-Z, a-z, 0-9 |
| device_id | string | ✅ | Уникальный ID устройства (5-100 символов) |
| birth_date | string (RFC3339) | ✅ | Дата рождения в формате ISO 8601 |

#### 📥 Ответы

**✅ Успех (200 OK)**
```
Token: 550e8400-e29b-41d4-a716-446655440000
```
*Возвращается UUID длительностью 7 минут для подтверждения email.*

**❌ Ошибка валидации (400 Bad Request)**
```
Bad request
```

**❌ Пользователь уже существует (409 Conflict)**
```
User with this email already exists
```

**❌ Возраст менее 16 лет (422 Unprocessable Entity)**
```
User must be at least 16 years old
```

---

### 2. Регистрация компании

**Отправка ИНН, email, пароля и device_id → получение Session ID для верификации.**

```
POST /auth/company
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
```

**Body:**
```json
{
  "inn": "7707083893",
  "email": "hr@company.ru",
  "password": "StrongPass456!",
  "device_id": "iphone_app_456"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| inn | string | ✅ | ИНН организации (10 или 12 цифр) |
| email | string | ✅ | Email представителя компании (max 255 символов) |
| password | string | ✅ | Пароль: 8-72 символов, обязательно: A-Z, a-z, 0-9 |
| device_id | string | ✅ | Уникальный ID устройства (5-100 символов) |

#### 📥 Ответы

**✅ Успех (200 OK)**
```
Token: 550e8400-e29b-41d4-a716-446655440000
```

**❌ Компания не найдена или не активна (404 Not Found)**
```
Company with this TIN not exists
```

**❌ Пользователь уже существует (409 Conflict)**
```
User with this email already exists
```

**❌ Ошибка внешнего сервиса (502 Bad Gateway)**
```
External service temporarily unavailable. Please try again later.
```

---

### 3. Верификация Email

**Подтверждение email кодом → получение Access и Refresh токенов.**

*Для студентов:*
```
POST /auth/students/verify
```

*Для компаний:*
```
POST /auth/company/verify
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
Token: 550e8400-e29b-41d4-a716-446655440000
```

**Body:**
```json
{
  "code": "482915"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| code | string | ✅ | 6-значный числовой код |

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token (JWT)
Set-Cookie: refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...; Path=/; HttpOnly; SameSite=Strict; Max-Age=604800
```

*Access Token (15 минут) возвращается в заголовке. Refresh Token (7 дней) устанавливается в HttpOnly cookie.*

**❌ Неверный формат кода (400 Bad Request)**
```
Bad request
```
*Например, если код не из 6 цифр.*

**❌ Сессия не найдена или код истек (400 Bad Request)**
```
Verification code expired
```

**❌ Неверный код (400 Bad Request)**
```
Invalid verification code format
```

**❌ Слишком много попыток (429 Too Many Requests)**
```
Too many invalid attempts
```

---

### 4. Запрос нового кода верификации

**Повторная отправка кода на email, пока активен Session ID.**

```
POST /auth/newverify
```

#### 📤 Запрос

**Headers:**
```http
Token: 550e8400-e29b-41d4-a716-446655440000
```

*Body: пустой*

#### 📥 Ответы

**✅ Успех (200 OK)**

**❌ Сессия не найдена или истекла (400/500 Internal Server Error)**
*Точный код зависит от обертки ошибок, но по логике будет Internal Server Error при проблемах с Redis.*

---

### 5. Вход в систему

*Для студентов:*
```
POST /auth/students/login
```

*Для компаний:*
```
POST /auth/company/login
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
```

**Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "device_id": "web_chrome_123",
  "birth_date": "2008-05-15T00:00:00Z"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| email | string | ✅ | Email пользователя |
| password | string | ✅ | Пароль |
| device_id | string | ✅ | ID устройства |
| birth_date | string | ❌ | Присутствует в структуре, но для входа не используется. Можно не передавать. |

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
Set-Cookie: refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...; Path=/; HttpOnly; SameSite=Strict; Max-Age=604800
```

**❌ Неверный пароль (401 Unauthorized)**
```
Invalid password
```

**❌ Пользователь не найден (404 Not Found)**
```
User with this email not exists
```

---

## 👤 Профиль

### Обновление профиля студента

```
PUT /my/profile/put
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
```

**Body (все поля опциональны):**
```json
{
  "first_name": "Иван",
  "last_name": "Иванов",
  "birth_date": "2008-05-15T00:00:00Z",
  "location": "Москва",
  "university": "МГУ",
  "degree": "Бакалавр",
  "bio": "Студент 3 курса...",
  "experience": "Стажировка в Яндексе",
  "image": "base64_encoded_image_string"
}
```

#### 📥 Ответы

**✅ Успех (200 OK)**

**❌ Неавторизован (401 Unauthorized)**
```
User Unauthorized
```

**❌ Доступ запрещен (403 Forbidden)**
```
Access denied: insufficient permissions
```
*Только пользователи с ролью "intern" могут обновлять этот профиль.*

---

### Получение профиля студента

```
GET /my/profile
```

#### 📤 Запрос

**Headers:**
```http
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
```

*Body: пустой*

#### 📥 Ответ

**✅ Успех (200 OK)**
```json
{
  "email": "student@example.com",
  "profile": {
    "first_name": "Иван",
    "last_name": "Иванов",
    "birth_date": "2008-05-15T00:00:00Z",
    "location": "Москва",
    "university": "МГУ",
    "degree": "Бакалавр",
    "bio": "...",
    "experience": "...",
    "image": "..."
  }
}
```

---

### Обновление профиля компании

```
PUT /my/company/profile/put
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
```

**Body (все поля опциональны, кроме ИНН из регистрации):**
```json
{
  "company_name": "ООО Рога и Копыта",
  "descrption": "Мы лучшая компания...",
  "website": "https://example.com"
}
```

#### 📥 Ответы

**✅ Успех (200 OK)**

**❌ Доступ запрещен (403 Forbidden)**
```
Access denied: insufficient permissions
```
*Только пользователи с ролью "company" могут обновлять этот профиль.*

---

### Получение профиля компании

```
GET /my/company/profile
```

#### 📤 Запрос

**Headers:**
```http
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
```

*Body: пустой*

#### 📥 Ответ

**✅ Успех (200 OK)**
```json
{
  "email": "hr@company.ru",
  "profile": {
    "company_name": "ООО Рога и Копыта",
    "descrption": "...",
    "website": "https://example.com",
    "inn": "7707083893",
    "kpp": "770701001",
    "ogrn": "1027735403160",
    "legal_address": "г. Москва, ул. Тверская, д. 1",
    "director_name": "Иванов Иван Иванович"
  }
}
```

---

## 🔄 Система токенов

| Тип | Где хранится | Время жизни | Формат | Использование |
|-----|-------------|-------------|--------|---------------|
| **Session ID** | Ответ сервера (заголовок `Token`) | 7 минут | UUID | Подтверждение email |
| **Access Token** | Ответ сервера (заголовок `Token`) | 15 минут | JWT | Авторизация API запросов |
| **Refresh Token** | HttpOnly Cookie (`refresh_token`) | 7 дней | JWT | Обновление Access Token |

### Автоматическое обновление Access Token

При истечении Access Token сервер автоматически выпустит новый, если у клиента есть действующий Refresh Token в cookies. Новый Access Token вернется в заголовке `Token` ответа.

Клиент должен иметь возможность получать обновленный `Token` из заголовков всех ответов API.

---

## 📊 Коды ответов

| HTTP Status | Сообщение (тело ответа) | Причина |
|-------------|--------------------------|---------|
| **400** | `Bad request` | Ошибка валидации входящих данных (email, пароль и т.д.) |
| **400** | `Invalid verification code format` | Код не соответствует формату (6 цифр) |
| **400** | `Verification code expired` | Истек Session ID (7 минут с момента запроса кода) |
| **401** | `Invalid password` | Неверный пароль при входе |
| **401** | `User Unauthorized` | Отсутствует, невалидный или просроченный токен |
| **403** | `Access denied: insufficient permissions` | Роль в JWT (`intern`/`company`) не соответствует эндпоинту |
| **404** | `User with this email not exists` | Попытка входа с незарегистрированным email |
| **404** | `Company with this TIN not exists` | ИНН не найден через DADATA или компания неактивна |
| **409** | `User with this email already exists` | Попытка повторной регистрации с тем же email и ролью |
| **422** | `User must be at least 16 years old` | На момент регистрации студенту меньше 16 лет |
| **429** | `Too many invalid attempts` | Превышен лимит запросов (Rate Limiting) |
| **502** | `External service temporarily unavailable...` | Ошибка при обращении к DADATA API |
| **503** | `Failed to send email, please try again` | Ошибка отправки письма |
| **503** | `Cache service temporarily unavailable` | Ошибка соединения с Redis |
| **500** | `Internal server error` | Ошибка БД, генерации JWT или другая внутренняя ошибка |

---

## 🌐 CORS и Cookies

- **Поддержка CORS:** Настроена для разрешенных origin, методов (`GET, PUT, POST, OPTIONS, PATCH`) и заголовков, включая `Authorization` и `Token`.
- **Обработка Preflight:** Сервер автоматически отвечает на все `OPTIONS` запросы к известным эндпоинтам статусом `200 OK`.
- **Клиентские запросы:** Для получения и автоматической отправки Refresh Token (HttpOnly Cookie), все fetch/Axios запросы должны включать опцию `credentials: 'include'`.

---

## 📋 Rate Limiting (Лимиты запросов)

Лимиты указаны в запросах в минуту на один уникальный ключ (для авторизованных — `[userID].[endpoint]`, для неавторизованных — `[sessionID].[endpoint]`).

| Эндпоинт | Лимит |
|----------|-------|
| `/auth/students` | 20 |
| `/auth/students/verify` | 60 |
| `/auth/newverify` | 7 |
| `/auth/students/login` | 30 |
| `/auth/company` | 20 |
| `/auth/company/verify` | 60 |
| `/auth/company/login` | 30 |
| `/my/profile/put` | 100 |
| `/my/profile` | 120 |
| `/my/company/profile/put` | 100 |
| `/my/company/profile` | 120 |
| Корневой `/` | Без ограничений |