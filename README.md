# 📚 T-match Backend API — Документация

## Оглавление

1. [Общая информация](#1-общая-информация)
2. [Аутентификация и регистрация](#2-аутентификация-и-регистрация)
3. [Профиль пользователя](#3-профиль-пользователя)
4. [Публичные профили](#4-публичные-профили)
5. [Стажировки](#5-стажировки)
6. [Поиск и фильтрация](#6-поиск-и-фильтрация)
7. [Навыки (Skills)](#7-навыки-skills)
8. [Отклики на стажировки](#8-отклики-на-стажировки)
9. [Система токенов](#9-система-токенов)
10. [Rate Limiting](#10-rate-limiting)
11. [Обработка ошибок](#11-обработка-ошибок)
12. [Запуск приложения](#12-запуск-приложения)

---

## 1. Общая информация

| Параметр | Значение |
| :--- | :--- |
| **Базовый URL** | `http://localhost:8080` |
| **Формат данных** | `application/json` (кроме загрузки аватара — `multipart/form-data`) |
| **Аутентификация** | Заголовок `Authorization: Bearer <access_token>` |
| **Cookies** | `refresh_token` (HttpOnly, SameSite=Strict, 7 дней) |
| **CORS** | Настраивается через `CORS_ALLOW_ORIGIN`, по умолчанию `http://localhost:8000` |
| **Даты** | ISO 8601 (`2026-05-15T00:00:00Z`) |

### Особенности работы с токенами
- Access Token живёт **15 минут**
- При истечении Access Token сервер проверяет Refresh Token из cookie
- Если Refresh Token валиден — сервер возвращает новый Access Token в заголовке `X-New-Access-Token`
- Фронтенд должен проверять этот заголовок в каждом ответе и обновлять токен

---

## 2. Аутентификация и регистрация

### 2.1 Регистрация студента

```
POST /auth/students
```

**Тело запроса:**
```json
{
  "email": "ivan@example.com",
  "password": "StrongPass1",
  "device_id": "web_chrome_abc123",
  "birth_date": "2008-05-15T00:00:00Z"
}
```

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `email` | `string` | ✅ | Email (макс. 255 символов) |
| `password` | `string` | ✅ | Пароль (8-72 символа, минимум 1 заглавная, 1 строчная, 1 цифра) |
| `device_id` | `string` | ✅ | Уникальный ID устройства (5-100 символов) |
| `birth_date` | `string` | ✅ | Дата рождения в ISO 8601, возраст ≥ 16 лет |

**Успешный ответ (201 Created):**
```
Заголовок: X-Verify-Session: 550e8400-e29b-41d4-a716-446655440000
Тело: пустое
```

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `400` | Ошибка валидации (неверный формат данных) |
| `409` | Email уже зарегистрирован |
| `422` | Возраст меньше 16 лет |

---

### 2.2 Регистрация компании

```
POST /auth/company
```

**Тело запроса:**
```json
{
  "inn": "7707083893",
  "email": "hr@company.ru",
  "password": "SecurePass456!",
  "device_id": "iphone_app_xyz"
}
```

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `inn` | `string` | ✅ | ИНН организации (10 или 12 цифр) |
| `email` | `string` | ✅ | Email |
| `password` | `string` | ✅ | Пароль (те же требования) |
| `device_id` | `string` | ✅ | ID устройства |

**Успешный ответ (201 Created):**
```
Заголовок: X-Verify-Session: 660e8400-e29b-41d4-a716-446655440001
Тело: пустое
```

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `400` | Ошибка валидации |
| `404` | Компания с таким ИНН не найдена или неактивна |
| `409` | Email уже зарегистрирован |
| `502` | Ошибка сервиса DaData |

---

### 2.3 Подтверждение email

```
POST /auth/students/verify   # для студента
POST /auth/company/verify    # для компании
```

**Заголовки:**
```
X-Verify-Session: <session_id из ответа регистрации>
```

**Тело запроса:**
```json
{
  "code": "482915"
}
```

**Успешный ответ (200 OK):**
```
Заголовок: Set-Cookie: refresh_token=<token>; HttpOnly; SameSite=Strict; Path=/; Max-Age=604800
```
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `400` | Неверный код или код истёк |
| `429` | Слишком много попыток |

---

### 2.4 Запрос нового кода подтверждения

```
POST /auth/newverify
```

**Заголовки:**
```
X-Verify-Session: <session_id>
```

**Успешный ответ (200 OK):** пустое тело

---

### 2.5 Вход в систему

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

**Успешный ответ (200 OK):**
```
Заголовок: Set-Cookie: refresh_token=<token>; HttpOnly; SameSite=Strict; Path=/; Max-Age=604800
```
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `401` | Неверный пароль |
| `404` | Email не найден |
| `429` | Превышен лимит попыток входа |

---

### 2.6 Выход из системы
```
POST /auth/logout
```

*Требуется авторизация*

**Успешный ответ (200 OK):** пустое тело

**Что происходит при выходе:**
- Refresh Token удаляется из Redis
- Кука `refresh_token` очищается (Max-Age=-1)
- Клиент должен удалить Access Token из локального хранилища

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `401` | Неавторизован |

---

## 3. Профиль пользователя

*Все эндпоинты требуют `Authorization: Bearer <access_token>`*

### 3.1 Профиль студента

#### Получение своего профиля

```
GET /my/profile
```
*Роль: `intern`*

**Пример ответа (200 OK):**
```json
{
  "email": "ivan@example.com",
  "profile": {
    "id": 17,
    "first_name": "Иван",
    "last_name": "Петров",
    "birth_date": "2008-05-15T00:00:00Z",
    "location": "Москва",
    "university": "МГУ",
    "degree": "Бакалавр",
    "bio": "Студент 4-го курса. Увлекаюсь backend-разработкой.",
    "experience": "Писал курсовые на Go, участвовал в хакатонах.",
    "image": "http://localhost:9000/t-match-storage/user_17_avatar"
  },
  "skills": [
    {"id": 1, "name": "Go"},
    {"id": 2, "name": "PostgreSQL"}
  ]
}
```

#### Обновление профиля

```
PUT /my/profile
```
*Роль: `intern`*

**Тело запроса** (все поля опциональны):
```json
{
  "first_name": "Иван",
  "last_name": "Петров",
  "birth_date": "2008-05-15T00:00:00Z",
  "location": "Москва",
  "university": "МГУ",
  "degree": "Бакалавр",
  "bio": "Студент 4-го курса.",
  "experience": "Опыт работы с Go и PostgreSQL."
}
```

**Успешный ответ (200 OK):** пустое тело

---

### 3.2 Профиль компании

#### Получение своего профиля

```
GET /my/company/profile
```
*Роль: `company`*

**Пример ответа (200 OK):**
```json
{
  "email": "hr@company.ru",
  "profile": {
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
}
```

#### Обновление профиля

```
PUT /my/company/profile
```
*Роль: `company`*

**Тело запроса** (все поля опциональны):
```json
{
  "company_name": "ООО Ромашка",
  "description": "IT-компания, разрабатываем крутые продукты",
  "website": "https://romashka.ru"
}
```

**Успешный ответ (200 OK):** пустое тело

---

### 3.3 Аватар

```
PUT /my/avatar
```
*Content-Type: `multipart/form-data`*

| Поле | Тип | Описание |
| :--- | :--- | :--- |
| `avatar` | `File` | Изображение (≤10 МБ, только JPEG или PNG) |

**Успешный ответ (201 Created):**
```json
"http://localhost:9000/t-match-storage/user_17_avatar"
```

---

## 4. Публичные профили

*Без авторизации (кроме `/my/avatar`)*

### 4.1 Публичный профиль студента

```
GET /profile/:id
```

**Пример:** `GET /profile/17`

**Ответ (200 OK):**
```json
{
  "profile": {
    "id": 17,
    "first_name": "Иван",
    "last_name": "Петров",
    "birth_date": "2008-05-15T00:00:00Z",
    "location": "Москва",
    "university": "МГУ",
    "degree": "Бакалавр",
    "bio": "Студент 4-го курса. Увлекаюсь backend-разработкой.",
    "experience": "Писал курсовые на Go, участвовал в хакатонах.",
    "image": null
  },
  "skills": [
    {"id": 1, "name": "Go"},
    {"id": 2, "name": "PostgreSQL"}
  ]
}
```

**Особенности:**
- Поле `email` обычно отсутствует
- **Исключение:** если запрос делает авторизованная компания, и этот студент имеет статус `reviewing` или `accepted` хотя бы по одной стажировке этой компании — поле `email` будет заполнено

---

### 4.2 Публичный профиль компании

```
GET /profile/company/:id
```

**Пример:** `GET /company/profile/5`

**Ответ (200 OK):**
```json
{
  "profile": {
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
  },
}
```

**Особенности:**
- Поле `email` обычно отсутствует
- **Исключение:** если запрос делает авторизованный студент, который был **принят** (статус `accepted`) хотя бы на одну стажировку этой компании — поле `email` будет заполнено


---

## 5. Стажировки

### 5.1 Создание стажировки

```
POST /internships
```
*Роль: `company`*

**Тело запроса:**
```json
{
  "title": "Стажер Go",
  "description": "Ищем мотивированного стажёра для работы над микросервисами на Go. Опыт работы с PostgreSQL приветствуется.",
  "salary": 50000,
  "location": "Москва",
  "duration_months": 3
}
```

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `title` | `string` | ✅ | Название (макс. 200 символов) |
| `description` | `string` | ✅ | Описание (макс. 5000 символов) |
| `salary` | `int` | | Зарплата (≥ 0) |
| `location` | `string` | ✅ | Город (макс. 200 символов) |
| `duration_months` | `int` | ✅ | Длительность в месяцах (≥ 1) |

**Успешный ответ (201 Created):**
```json
23
```

---

### 5.2 Получение стажировки по ID

```
GET /internships/:id
```
*Без авторизации*

**Пример:** `GET /internships/15`

**Ответ (200 OK):**
```json
{
  "internship": {
    "id": 15,
    "company_id": 5,
    "title": "Стажер Go",
    "description": "Ищем мотивированного стажёра для работы над микросервисами на Go. Опыт работы с PostgreSQL приветствуется.",
    "salary": 50000,
    "location": "Москва",
    "is_archived": false,
    "duration_months": 3,
    "created_at": "2025-05-09T10:30:00Z"
  },
  "skills": [
    {"id": 1, "name": "Go"},
    {"id": 2, "name": "Docker"}
  ]
}
```

---

### 5.3 Обновление стажировки

```
PUT /internships/:id
```
*Роль: `company` (только владелец)*

**Тело запроса** (все поля опциональны):
```json
{
  "title": "Новое название",
  "description": "Обновлённое описание",
  "salary": 60000,
  "location": "Санкт-Петербург",
  "duration_months": 6
}
```

**Успешный ответ (200 OK):** пустое тело

---

### 5.4 Архивирование стажировки

```
DELETE /internships/:id
```
*Роль: `company` (только владелец)*

**Успешный ответ (200 OK):** пустое тело

---

## 6. Поиск и фильтрация

### 6.1 Поиск стажировок (превью)

```
GET /internships
```
*Без авторизации*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | - | Текстовый поиск (слова через пробел, логика OR) |
| `location` | `string` | - | Город (частичное совпадение, регистронезависимый) |
| `salary_min` | `int` | - | Минимальная зарплата |
| `salary_max` | `int` | - | Максимальная зарплата |
| `duration_min` | `int` | - | Минимальная длительность (месяцы) |
| `duration_max` | `int` | - | Максимальная длительность (месяцы) |
| `skills` | `[]int` | - | ID навыков (AND-логика). Передавать несколько раз: `?skills=1&skills=2` |
| `sort` | `string` | `created_at` | Поле сортировки: `salary`, `duration_months`, `created_at` |
| `order` | `string` | `desc` | Направление: `asc` или `desc` |
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
```

**Пример ответа (200 OK):**
```json
[
  {
    "id": 15,
    "company_id": 5,
    "title": "Go разработчик",
    "salary": 80000,
    "location": "Москва",
    "is_archived": false,
    "duration_months": 6,
    "created_at": "2026-05-15T10:30:00Z"
  },
  {
    "id": 16,
    "company_id": 8,
    "title": "Python Backend Developer",
    "salary": 70000,
    "location": "Санкт-Петербург",
    "is_archived": false,
    "duration_months": 3,
    "created_at": "2026-05-20T14:00:00Z"
  }
]
```

**Важно:** Это **превью** для карточек в списке. Полное описание и навыки — через `GET /internships/:id`.

---

### 6.2 Поиск студентов (превью)

```
GET /students
```
*Без авторизации*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | - | Текстовый поиск по имени, фамилии, университету, степени |
| `university` | `string` | - | Университет (частичное совпадение) |
| `skills` | `[]int` | - | ID навыков (AND-логика). `?skills=1&skills=2` |
| `limit` | `int` | 20 | Записей на странице |
| `offset` | `int` | 0 | Сдвиг для пагинации |

**Пример:** `GET /students?query=Иван&university=МГУ&skills=1&skills=2`

**Пример ответа (200 OK):**
```json
[
  {
    "id": 17,
    "first_name": "Иван",
    "last_name": "Петров",
    "location": "Москва",
    "university": "МГУ",
    "degree": "Computer Science",
    "image": null
  },
  {
    "id": 18,
    "first_name": "Мария",
    "last_name": "Сидорова",
    "location": "Санкт-Петербург",
    "university": "МГУ",
    "degree": "Бакалавр",
    "image": "http://localhost:9000/t-match-storage/user_18_avatar"
  }
]
```

**Важно:** Это **превью** без чувствительных данных (нет `birth_date`, `bio`, `experience`). Полный профиль — через `GET /profile/:id`.

---

### 6.3 Поиск компаний (превью)

```
GET /companies
```
*Без авторизации*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | - | Текстовый поиск по названию, описанию, адресу |
| `location` | `string` | - | Юридический адрес (частичное совпадение) |
| `limit` | `int` | 20 | Записей на странице |
| `offset` | `int` | 0 | Сдвиг для пагинации |

**Пример:** `GET /companies?query=IT&location=Москва&limit=10`

**Пример ответа (200 OK):**
```json
[
  {
    "id": 5,
    "company_name": "Яндекс",
    "description": "Крупная IT-компания, разработка поисковых технологий",
    "website": "https://ya.ru",
    "inn": "7707083893",
    "ogrn": "1027700229193",
    "legal_address": "Москва, ул. Льва Толстого, 16",
    "image": null
  },
  {
    "id": 8,
    "company_name": "ООО Ромашка",
    "description": "IT-компания, разрабатываем крутые продукты",
    "website": "https://romashka.ru",
    "inn": "7707083894",
    "ogrn": "1027700229194",
    "legal_address": "Москва, ул. Тверская, 1",
    "image": null
  }
]
```

**Важно:** Это **превью**. Полный профиль компании (с `kpp`, `director_name`) — через `GET /profile/company/:id`.

---

## 7. Навыки (Skills)

### 7.1 Список всех навыков

```
GET /skills
```
*Без авторизации*

**Пример ответа (200 OK):**
```json
[
  {"id": 1, "name": "Go"},
  {"id": 2, "name": "PostgreSQL"},
  {"id": 3, "name": "Docker"},
  {"id": 4, "name": "Python"},
  {"id": 5, "name": "Redis"}
]
```

---

### 7.2 Управление навыками стажёра

#### Добавить навыки

```
POST /my/profile/skills
```
*Роль: `intern`*

**Тело запроса:**
```json
{
  "skill_id": [1, 2, 5]
}
```

**Успешный ответ (201 Created):** пустое тело

#### Удалить навыки

```
DELETE /my/profile/skills
```
*Роль: `intern`*

**Тело запроса:**
```json
{
  "skill_id": [2, 5]
}
```

**Успешный ответ (200 OK):** пустое тело

---

### 7.3 Управление навыками стажировки

#### Добавить навыки

```
POST /internships/:id/skill
```
*Роль: `company` (владелец)*

**Тело запроса:**
```json
{
  "skill_id": [1, 2, 3]
}
```

**Успешный ответ (201 Created):** пустое тело

#### Удалить навыки

```
DELETE /internships/:id/skill
```
*Роль: `company` (владелец)*

**Тело запроса:**
```json
{
  "skill_id": [2]
}
```

**Успешный ответ (200 OK):** пустое тело

---

## 8. Отклики на стажировки

### 8.1 Откликнуться на стажировку

```
POST /internships/:id/respond
```
*Роль: `intern`*

**Успешный ответ (200 OK):** пустое тело

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `409` | Уже откликался на эту стажировку |
| `410` | Стажировка архивирована |

---

### 8.2 Мои отклики (для стажёра)

```
GET /my/responses
```
*Роль: `intern`*

**Пример ответа (200 OK):**
```json
[
  {
    "id": 101,
    "intern_id": 17,
    "internship_id": 42,
    "status": "pending",
    "created_at": "2025-05-09T10:30:00Z"
  },
  {
    "id": 102,
    "intern_id": 17,
    "internship_id": 15,
    "status": "accepted",
    "created_at": "2025-05-10T14:00:00Z"
  }
]
```

**Возможные статусы:**

| Статус | Описание |
| :--- | :--- |
| `pending` | Ожидает рассмотрения |
| `reviewing` | На рассмотрении |
| `accepted` | Принят |
| `rejected` | Отклонён |

---

### 8.3 Отклики на стажировку (для компании)

```
GET /internships/:id/responses
```
*Роль: `company` (владелец стажировки)*

**Пример ответа (200 OK):**
```json
[
  {
    "id": 101,
    "intern_id": 17,
    "internship_id": 42,
    "status": "reviewing",
    "created_at": "2025-05-09T10:30:00Z"
  }
]
```

**Особенность:** При первом запросе все отклики со статусом `pending` автоматически переводятся в `reviewing`.

---

### 8.4 Изменение статуса отклика

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

Допустимые значения: `pending`, `reviewing`, `accepted`, `rejected`.

**Успешный ответ (200 OK):** пустое тело

---

## 9. Система токенов

| Токен | Передача | Время жизни | Назначение |
| :--- | :--- | :--- | :--- |
| **Session ID** | Заголовок `X-Verify-Session` | 7 минут | Верификация email |
| **Access Token** | Заголовок `Authorization: Bearer <token>` | 15 минут | Авторизация API |
| **Refresh Token** | Cookie `refresh_token` (HttpOnly) | 7 дней | Обновление Access Token |

**Алгоритм обновления Access Token:**
1. Фронтенд отправляет запрос с истёкшим Access Token
2. Сервер проверяет Refresh Token из cookie
3. Если Refresh Token валиден — возвращает новый Access Token в заголовке `X-New-Access-Token`
4. Фронтенд должен перехватывать этот заголовок и сохранять новый токен

### Выход из системы и инвалидация токенов

При вызове `POST /auth/logout`:
1. Сервер удаляет Refresh Token из Redis
2. Кука `refresh_token` очищается (устанавливается `Max-Age=-1`)
3. Клиент **должен удалить** Access Token из локального хранилища

### Вход с нового устройства

При входе с новым `device_id` создается новая пара токенов. Старый Refresh Token (для предыдущего `device_id`) остается активным до истечения срока (7 дней) или до ручного выхода.

---

## 10. Rate Limiting

Лимиты в запросах в минуту на IP/пользователя:

### Аутентификация
| Эндпоинт | Лимит |
| :--- | :--- |
| `POST /auth/students` | 20 |
| `POST /auth/students/verify` | 60 |
| `POST /auth/newverify` | 7 |
| `POST /auth/students/login` | 30 |
| `POST /auth/company` | 20 |
| `POST /auth/company/verify` | 60 |
| `POST /auth/company/login` | 30 |

### Профиль
| Эндпоинт | Лимит |
| :--- | :--- |
| `PUT /my/profile` | 100 |
| `GET /my/profile` | 120 |
| `POST /my/profile/skills` | 5 |
| `DELETE /my/profile/skills` | 5 |
| `PUT /my/company/profile` | 100 |
| `GET /my/company/profile` | 120 |
| `PUT /my/avatar` | 100 |

### Стажировки
| Эндпоинт | Лимит |
| :--- | :--- |
| `POST /internships` | 12 |
| `PUT /internships/:id` | 12 |
| `DELETE /internships/:id` | 5 |
| `POST /internships/:id/skill` | 5 |
| `DELETE /internships/:id/skill` | 5 |

### Отклики
| Эндпоинт | Лимит |
| :--- | :--- |
| `POST /internships/:id/respond` | 10 |
| `GET /my/responses` | 20 |
| `GET /internships/:id/responses` | 20 |
| `PUT /responses/:id/status` | 20 |

### Поиск
| Эндпоинт | Лимит |
| :--- | :--- |
| `GET /internships` | 60 |
| `GET /students` | 60 |
| `GET /companies` | 60 |

### Без лимита
| Эндпоинт |
| :--- |
| `GET /internships/:id` |
| `GET /profile/:id` |
| `GET /profile/company/:id` |
| `GET /skills` |
| `OPTIONS /*path` |

---

## 11. Обработка ошибок

Все ошибки возвращаются в формате plain text с соответствующим HTTP-статусом.

| Код | Сообщение | Когда возникает |
| :--- | :--- | :--- |
| `400` | `Bad request` | Невалидный JSON, неверный формат данных |
| `401` | `User Unauthorized` | Токен отсутствует, истёк или невалиден |
| `403` | `Access denied: insufficient permissions` | Нет прав на действие |
| `404` | `Not found` | Ресурс не найден |
| `409` | `User with this email already exists` | Email занят |
| `409` | `You have already responded to this internship` | Повторный отклик |
| `410` | `Internship is archived` | Стажировка в архиве |
| `422` | `User must be at least 16 years old` | Возраст < 16 лет |
| `429` | `Too many invalid attempts` | Превышен rate limit |
| `500` | `Internal server error` | Внутренняя ошибка сервера |
| `502` | `External service temporarily unavailable` | Ошибка DaData |
| `503` | `Cache service temporarily unavailable` | Redis недоступен |
| `503` | `Failed to send email, please try again` | Ошибка отправки email |

---

## Примечания по поиску

**Полнотекстовый поиск:**
- Поддерживает русскую морфологию (словоформы)
- Слова в запросе соединяются через OR
- Результаты сортируются по релевантности

**Фильтрация по навыкам:**
- AND-логика: должны быть ВСЕ указанные навыки
- ID навыков получать через `GET /skills`

**Пагинация:**
- По умолчанию 20 записей
- Для следующей страницы: `offset = предыдущий_offset + limit`

**Разделение preview/полные данные:**
- Поиск (`GET /internships`, `GET /students`, `GET /companies`) — только превью
- Получение по ID (`GET /internships/:id`, `GET /profile/:id`, `GET /profile/company/:id`) — все данные

## 12. Запуск приложения

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
CORS_ALLOW_HEADERS=Content-Type,Authorization,X-Verify-Session,X-New-Access-Token

# S3/MinIO
S3_ENDPOINT=minio:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_USE_SSL=false
S3_PORT=:9000
S3_HOST=0.0.0.0

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

Email: если EMAIL_PASSWORD не задан, автоматически используется MailHog — все письма перехватываются и отображаются в веб-интерфейсе http://localhost:8025, реальным пользователям ничего не отправляется.

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