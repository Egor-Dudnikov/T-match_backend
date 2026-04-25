## Базовый URL

```
Development: http://localhost:8080
```

---

## 🔐 Аутентификация

### 1. Регистрация студента / соискателя

**Отправка email, пароля, device_id и даты рождения → получение Session ID**

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
  "password": "SecurePass123",
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

**Headers:**
```
Token: 550e8400-e29b-41d4-a716-446655440000
```
*Session ID (UUID)*

**❌ Ошибка валидации (400 Bad Request)**

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

**Отправка ИНН, email, пароля и device_id → получение Session ID**

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
  "password": "StrongPass456",
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

**Headers:**
```
Token: 550e8400-e29b-41d4-a716-446655440000
```

**❌ Компания не найдена (404 Not Found)**
```
Company with this TIN not exists
```
*(ИНН не найден в DADATA или компания не активна)*

**❌ Пользователь уже существует (409 Conflict)**
```
User with this email already exists
```

**❌ Ошибка внешнего сервиса (502 Bad Gateway)**
```
External service temporarily unavailable. Please try again later.
```

---

### 3. Верификация email (общая)

**Подтверждение email кодом → получение Access и Refresh токенов**

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

**❌ Неверный формат кода (400 Bad Request)**
```
Invalid verification code format
```

**❌ Код истек или Session ID не найден (400 Bad Request)**
```
Verification code expired
```

**❌ Слишком много попыток (429 Too Many Requests)**
```
Too many invalid attempts
```

---

### 4. Повторная отправка кода

**Запрос нового кода верификации (до истечения Session ID)**

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

**❌ Session ID не найден или лимит превышен (400/429)**

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
  "password": "Password123",
  "device_id": "web_chrome_123",
  "birth_date": "2008-05-15T00:00:00Z"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| email | string | ✅ | Email пользователя |
| password | string | ✅ | Пароль |
| device_id | string | ✅ | ID устройства |
| birth_date | string | ❌ | Требуется только для студентов при регистрации |

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
PUT /my/profile
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
```

**Body:**
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
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| first_name | string | ❌ | Имя |
| last_name | string | ❌ | Фамилия |
| birth_date | string | ❌ | Дата рождения (RFC3339) |
| location | string | ❌ | Город/регион |
| university | string | ❌ | Учебное заведение |
| degree | string | ❌ | Степень/курс |
| bio | string | ❌ | О себе |
| experience | string | ❌ | Опыт работы/проекты |

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
*(Только студенты могут обновлять профиль)*

---

## 🔄 Система токенов

| Тип | Где хранится | Время жизни | Формат | Использование |
|-----|-------------|-------------|--------|---------------|
| **Session ID** | Redis | 7 минут | UUID | Верификация email (заголовок `Token`) |
| **Access Token** | Клиент (Memory) | 15 минут | JWT | Авторизация API (заголовок `Token`) |
| **Refresh Token** | HttpOnly Cookie | 7 дней | JWT | Обновление Access Token |

### Структура JWT Claims

```json
{
  "UserID": 101,
  "DeviceID": "web_chrome_123",
  "Email": "user@example.com",
  "Role": "intern",
  "exp": 1705314600,
  "iat": 1705312800,
  "iss": "t-match_backend"
}
```

### Автоматическое обновление Access Token

При запросе с истекшим Access Token, но валидным Refresh Token:
- Сервер автоматически сгенерирует новый Access Token
- Новый токен вернется в заголовке `Token`
- Запрос будет обработан как обычно

---

## 📊 Коды ответов

| HTTP Status | Тело ответа | Причина |
|-------------|-------------|---------|
| **400** | `Bad request` | Ошибка валидации JSON или полей |
| **400** | `Invalid verification code format` | Код не из 6 цифр |
| **400** | `Verification code expired` | Session ID истек (7 минут) |
| **401** | `Invalid password` | Неверный пароль |
| **401** | `User Unauthorized` | Отсутствует или неверный токен |
| **403** | `Access denied: insufficient permissions` | Роль не соответствует эндпоинту |
| **404** | `User with this email not exists` | Email не зарегистрирован |
| **404** | `Company with this TIN not exists` | ИНН не найден в DADATA |
| **409** | `User with this email already exists` | Email уже используется |
| **422** | `User must be at least 16 years old` | Возраст менее 16 лет |
| **429** | `Too many invalid attempts` | Превышен лимит попыток |
| **502** | `External service temporarily unavailable...` | Ошибка DADATA API |
| **503** | `Failed to send email, please try again` | Ошибка SMTP |
| **503** | `Cache service temporarily unavailable` | Ошибка Redis |
| **500** | `Internal server error` | БД, JWT или другая внутренняя ошибка |

---

## 🌐 CORS

Сервер поддерживает CORS из коробки:
- **Allow-Origin:** Настраивается через конфиг (`control_allow_origin`)
- **Allow-Methods:** `GET, PUT, POST, OPTIONS, PATCH`
- **Allow-Headers:** Настраивается через конфиг
- **Allow-Credentials:** `true`

**Важно:** При запросах с клиента необходимо указывать `credentials: 'include'` для корректной работы с Refresh Token в cookies.

---

## 📋 Rate Limiting

| Эндпоинт | Лимит (запросов/минуту) |
|----------|------------------------|
| `/auth/students` | 20 |
| `/auth/students/verify` | 60 |
| `/auth/newverify` | 7 |
| `/auth/students/login` | 30 |
| `/auth/company` | 20 |
| `/auth/company/verify` | 60 |
| `/auth/company/login` | 30 |
| `/my/profile` | 300 |