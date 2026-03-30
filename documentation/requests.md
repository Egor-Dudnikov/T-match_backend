## Базовый URL
```
Development: http://localhost:8080
```

---

## 🔐 Аутентификация

### 1. Регистрация пользователя
**Отправка email и пароля, получение session ID**

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
  "email": "user@example.com",
  "password": "Password123",
  "device_id": "device_1234567890"
}
```

**Поля:**
| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| email | string | ✅ | Email пользователя (валидация email) |
| password | string | ✅ | Пароль (8-72 символа) |
| device_id | string | ✅ | Уникальный ID устройства (5-100 символов) |

**Требования к паролю:**
- Минимум 8 символов
- Максимум 72 символа
- Должен содержать хотя бы:
  - Одну заглавную букву (A-Z)
  - Одну цифру (0-9)
  - Одну строчную букву (a-z)

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: a7f3e9d2c1b4a5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7
```
*В ответе возвращается Session ID (строка 64 символа)*


**❌ Ошибка валидации (400 Bad Request)**

**Body:**
```
Validation failed
```

**❌ Пользователь уже существует (409 Conflict)**

**Body:**
```
User with this email already exists
```

---

### 2. Верификация пользователя
**Подтверждение email кодом, получение access и refresh токенов**

```
POST /auth/students/verify
```

#### 📤 Запрос

**Headers:**
```http
Content-Type: application/json
Token: a7f3e9d2c1b4a5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7
```

**Body:**
```json
{
  "code": "123456"
}
```

**Поля:**
| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| code | string | ✅ | 6-значный числовой код из email |

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token
Set-Cookie: refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...; Path=/; HttpOnly; SameSite=Lax; Max-Age=604800
```

**Body:**
```json
{
  "message": "User verified successfully"
}
```

**Важно:** 
- **Access Token** передается в заголовке `Token` (время жизни: 15 минут)
- **Refresh Token** автоматически устанавливается в HttpOnly cookie (время жизни: 7 дней)
- Cookie настройки: `HttpOnly` (недоступен из JavaScript), `SameSite=Lax`, `Max-Age=604800`

**❌ Неверный или просроченный Session ID (400 Bad Request)**

**Body:**
```
Verification code expired
```

**❌ Неверный код (400 Bad Request)**

**Body:**
```
Invalid verification code format
```

**❌ Ошибка валидации кода (400 Bad Request)**

**Body:**
```
Validation failed
```

---

## 🔄 Работа с токенами

### Типы токенов и идентификаторов

| Тип | Где хранится | Время жизни | Формат | Использование |
|-----|-------------|-------------|--------|---------------|
| **Session ID** | Memory / State | 7 минут | Hex строка (64 символа) | Для верификации email, возвращается при регистрации |
| **Access Token** | Memory | 15 минут | JWT | Для авторизации API запросов |
| **Refresh Token** | HttpOnly cookie | 7 дней | JWT | Для обновления access token |

### Описание процесса

1. **Регистрация**: 
   - Пользователь отправляет email, пароль и device_id
   - Сервер генерирует 6-значный код и отправляет на email
   - Сервер создает Session ID (случайная строка) и сохраняет временные данные в Redis
   - Клиент получает Session ID в заголовке `Token`

2. **Верификация**:
   - Клиент отправляет Session ID (в заголовке) и код из email
   - Сервер проверяет код, создает пользователя в БД
   - Сервер генерирует Access Token (15 мин) и Refresh Token (7 дней)
   - Клиент получает Access Token в заголовке, Refresh Token в HttpOnly cookie

3. **Использование токенов**:
   - Все последующие API запросы используют Access Token в заголовке `Token`
   - Refresh Token автоматически отправляется браузером в cookie при запросах на обновление

### Структура JWT Claims

**Access Token и Refresh Token содержат:**
```json
{
  "user_id": "123",
  "device_id": "device_1234567890",
  "email": "user@example.com",
  "exp": 1705314600,
  "iat": 1705312800,
  "iss": "t-match_backend"
}
```

---

## 📊 Полный сценарий работы

### Обработка ошибок

| HTTP Status | Ошибка | Описание |
|-------------|--------|----------|
| 400 | Validation failed | Неверный формат email, пароля или device_id |
| 400 | Invalid verification code format | Код не соответствует формату (не 6 цифр) |
| 400 | Verification code expired | Session ID истек или не найден в Redis |
| 409 | User with this email already exists | Пользователь с таким email уже зарегистрирован |
| 500 | Internal server error | Ошибка БД, генерации JWT или другие внутренние ошибки |
| 503 | Failed to send email | Ошибка отправки email (временно недоступно) |
| 503 | Cache service temporarily unavailable | Ошибка Redis (временно недоступно) |

### Безопасность

1. **Пароли**: Хранятся в БД в хешированном виде (bcrypt)
2. **Session ID**: Криптостойкая случайная строка, не содержит информации о пользователе
3. **Access Token**: JWT с коротким временем жизни (15 минут)
4. **Refresh Token**: 
   - HttpOnly cookie (недоступен для JavaScript)
   - Долгое время жизни (7 дней)
   - Привязан к user_id и device_id
5. **Код верификации**: 
   - 6 цифр
   - Отправляется только на email
   - Хранится в Redis с TTL 7 минут
