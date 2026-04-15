### Бекенд для проекта T-match
*сервис для подборки стажировок и стажёров*

## Базовый URL

```
Development: http://localhost:8080
```

---

## 🔐 Аутентификация

### 1. Регистрация пользователя (Студент / Соискатель)

**Отправка email и пароля, получение Session ID**

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
  "device_id": "web_chrome_123"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| email | string | ✅ | Email пользователя |
| password | string | ✅ | Пароль: мин. 8 символов, **обязательно**: A-Z, a-z, 0-9 |
| device_id | string | ✅ | Уникальный ID устройства (мин. 5 символов) |

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: 550e8400-e29b-41d4-a716-446655440000
```
*Формат: стандартный UUID (36 символов)*

**❌ Ошибка валидации (400 Bad Request)**
*Тело ответа: `Bad request` или текст ошибки валидации поля.*

**❌ Пользователь уже существует (409 Conflict)**
*Тело ответа: `User with this email already exists`*

---

### 2. Регистрация компании (Работодатель)

**Отправка email, пароля и ИНН, получение Session ID**

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
| email | string | ✅ | Email представителя компании |
| password | string | ✅ | Пароль: мин. 8 символов, **обязательно**: A-Z, a-z, 0-9 |
| device_id | string | ✅ | Уникальный ID устройства |

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: 550e8400-e29b-41d4-a716-446655440000
```

**❌ Ошибка (404 Not Found)**
*Тело ответа: `Company with this TIN not exists` (если компания не найдена в DADATA или не активна).*

---

### 3. Верификация (Общая для студентов и компаний)

**Подтверждение email кодом, получение Access и Refresh токенов**

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
Token: 550e8400-e29b-41d4-a716-446655440000  // Session ID из шага 1 или 2
```

**Body:**
```json
{
  "code": "482915"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| code | string | ✅ | 6-значный числовой код (строка из цифр) |

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token (JWT)
Set-Cookie: refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...; Path=/; HttpOnly; SameSite=Strict; Max-Age=604800
```

**Body:**
*Пустое тело (или простой текст OK, в коде не указан body).*

**Важно:**
- **Access Token** в заголовке `Token` (время жизни: 15 минут).
- **Refresh Token** в HttpOnly Cookie `refresh_token` (время жизни: 7 дней).

**❌ Код неверный (400 Bad Request)**
*Тело ответа: `Invalid verification code format`*

**❌ Session ID истек (400 Bad Request)**
*Тело ответа: `Verification code expired`*

**❌ Слишком много попыток (429 Too Many Requests)**
*Тело ответа: `Too many invalid attempts`*

---

### 4. Повторная отправка кода

**Если код не пришел или истек (до истечения срока жизни Session ID)**

```
POST /auth/newverify
```

#### 📤 Запрос

**Headers:**
```http
Token: 550e8400-e29b-41d4-a716-446655440000  // Текущий Session ID
```

*Body: Пустой*

#### 📥 Ответы

**✅ Успех (200 OK)**
*На почту придет новый код. Старый код становится невалидным. Ограничение: 3 генерации кода на 1 сессию.*

**❌ Ошибка (400 Bad Request)**
*Если Session ID не найден или лимит генераций исчерпан.*

---

### 5. Вход (Логин)

**Для студентов:**
```
POST /auth/students/login
```

**Для компаний:**
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
  "device_id": "device_1234567890"
}
```

#### 📥 Ответы

**✅ Успех (200 OK)**

**Headers:**
```
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  // Access Token (JWT)
Set-Cookie: refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...; Path=/; HttpOnly; SameSite=Strict; Max-Age=604800
```

**❌ Неверный пароль (401 Unauthorized)**
*Тело ответа: `Invalid password`*

**❌ Пользователь не найден (404 Not Found)**
*Тело ответа: `User with this email not exists`*

---

## 🔄 Работа с токенами

### Типы токенов и идентификаторов

| Тип | Где хранится | Время жизни | Формат | Использование |
|-----|-------------|-------------|--------|---------------|
| **Session ID** | Redis (Server) | 7 минут | UUID | Только для этапа верификации email. |
| **Access Token** | Memory (Client) | 15 минут | JWT | Для авторизации API запросов в заголовке `Token`. |
| **Refresh Token** | HttpOnly Cookie | 7 дней | JWT | Для обновления Access Token. |

### Структура JWT Claims (для разработчиков)

**Access Token и Refresh Token содержат:**
```json
{
  "UserID": "101",
  "DeviceID": "web_chrome_123",
  "Email": "user@example.com",
  "Role": "internal",  // или "company"
  "exp": 1705314600,
  "iat": 1705312800,
  "iss": "t-match_backend"
}
```

---

## 📊 Обработка ошибок (Сводная таблица)

| HTTP Status | Ошибка (Тело ответа) | Причина |
|-------------|----------------------|---------|
| **400** | `Bad request` | Ошибка валидации полей (пароль простой, email неверный) |
| **400** | `Invalid verification code format` | Код не из 6 цифр |
| **400** | `Verification code expired` | Session ID просрочен (7 мин) |
| **401** | `Invalid password` | Неверный пароль при логине |
| **404** | `User with this email not exists` | Email не найден в системе |
| **404** | `Company with this TIN not exists` | ИНН не найден в DADATA или компания ликвидирована |
| **409** | `User with this email already exists` | Email уже зарегистрирован |
| **429** | `Too many invalid attempts` | >3 попыток ввода кода / >3 запросов нового кода |
| **500** | `Internal server error` | Ошибка БД, генерации JWT |
| **502** | `External service temporarily unavailable...` | Ошибка связи с DADATA |
| **503** | `Failed to send email...` / `Cache service temporarily unavailable` | Ошибка SMTP или Redis |

### Дополнительно: CORS

Сервер настроен на прием запросов с любых ориджинов (`*` в коде нет, но `ControlAllowOrigin` берется из конфига) и поддерживает метод `OPTIONS` для всех основных эндпоинтов. Не забудьте включать `credentials: 'include'` в запросах на клиенте, чтобы работали Cookies.
