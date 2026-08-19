# 📚 T-match Backend API — Документация

## Оглавление

1. [Общая информация](#1-общая-информация)
2. [Аутентификация и регистрация](#2-аутентификация-и-регистрация)
3. [Профиль пользователя](#3-профиль-пользователя)
4. [Публичные профили](#4-публичные-профили)
5. [Стажировки](#5-стажировки)
6. [Поиск и фильтрация](#6-поиск-и-фильтрация)
7. [Города (Cities)](#7-города-cities)
8. [Навыки (Skills)](#8-навыки-skills)
9. [Отклики на стажировки](#9-отклики-на-стажировки)
10. [Уведомления](#10-уведомления)
11. [Система токенов](#11-система-токенов)
12. [Rate Limiting](#12-rate-limiting)
13. [Рекомендательная система](#13-рекомендательная-система)
14. [Обработка ошибок](#14-обработка-ошибок)
15. [Админ-панель](#15-админ-панель)
16. [Запуск приложения](#16-запуск-приложения)

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

### Архитектурные решения

**Почему монолит.** Бэкенд — один Go-процесс с чёткими слоями (`handlers → service → repository`), один деплой, один источник данных (PostgreSQL). Для команды из 1–2 человек и текущего домена монолит — это осознанный выбор: нет распределённых транзакций, дешевле тестировать (сервисный слой покрыт юнит-тестами на sqlmock без реальной БД), быстрее итерировать. Границы между слоями выделены так, чтобы при росте вынести отдельный домен (например, уведомления) в микросервис без переписывания всего.

**Почему рекомендации — отдельный сервис (recsys).** Рекомендательный движок — это другой стек (Python/FastAPI) и другая модель данных: гео-ранжирование, косинусная схожесть навыков, ALS-коллаборативная фильтрация. Вынося его в отдельный сервис, мы:
- не тащим ML-зависимости и тяжёлые вычисления в Go-процесс;
- получаем независимое масштабирование (recsys — CPU-интенсивный, у него своя БД и Redis-кэш);
- изолируем жизненный цикл: свой репозиторий, свой docker-проект, общая сеть `t-match-net`.

Связь «бэкенд → recsys» — по HTTP через тонкий клиент (`internal/recsys`).

**Сознательные компромиссы:**
- **Синхронизация — fire-and-forget.** События (создание, гео, скилы, actions) отправляются без очереди и ретраев; при недоступном recsys они теряются, рассинхрон восстановится только после следующего события. Принято ради простоты; при необходимости усиливается outbox-паттерном или очередью.
- **Чтение жёсткое, запись мягкая.** `GET /my/recommendations` при недоступном recsys отвечает `502`, а вся остальная функциональность продолжает работать. При пустом `RECSYS_URL` клиент — no-op, и сервис рекомендаций полностью опционален для разработки.
- **Recsys хранит только ID и рейтинги.** Карточки стажировок бэкенд подтягивает из своей БД, поэтому архивные/удалённые/забаненные стажировки отсекаются на лету — ранжирование может не совпадать с фактической выдачей на величину «протухших» записей.
- **Dev-безопасность:** refresh-кука без `Secure` по умолчанию (обычный HTTP в dev); в проде флаг включается через `COOKIE_SECURE=true` в `.env`. Лимит тела аватара — через `MaxBytesReader`.
- **Rate limiting** реализован на уровне приложения поверх Redis, а не на инфраструктурном шлюзе — уязвим к обходу при горизонтальном масштабировании, но проще в управлении и тестировании.

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
| `password` | `string` | ✅ | Пароль (8–72 символа, минимум 1 заглавная, 1 строчная, 1 цифра) |
| `device_id` | `string` | ✅ | Уникальный ID устройства (5–100 символов) |
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
- Кука `refresh_token` очищается (`Max-Age=-1`)
- Клиент должен удалить Access Token из локального хранилища

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `401` | Неавторизован |

---

### 2.7 Восстановление пароля

#### Запрос на восстановление

```
POST /auth/forgot-password
```

**Тело запроса:**
```json
{
  "device_id": "web_chrome_abc123",
  "email": "ivan@example.com",
  "role": "intern"
}
```

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `device_id` | `string` | ✅ | ID устройства |
| `email` | `string` | ✅ | Email |
| `role` | `string` | ✅ | Роль: `intern` или `company` |

**Успешный ответ (200 OK):**
```
Заголовок: X-Verify-Session: <session_id>
```

---

#### Подтверждение восстановления

```
POST /auth/forgot-password/verify
```

**Заголовки:**
```
X-Verify-Session: <session_id>
```

**Тело запроса:**
```json
{
  "code": "482915"
}
```

**Успешный ответ (200 OK):** возвращает новую пару токенов (как при обычной верификации)

---

#### Смена пароля

```
PUT /auth/change-password
```

*Требуется авторизация*

**Тело запроса:**
```json
{
  "password": "NewSecurePass1"
}
```

**Успешный ответ (200 OK):** пустое тело

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
    "user_id": 9,
    "first_name": "Иван",
    "last_name": "Петров",
    "birth_date": "2008-05-15T00:00:00Z",
    "city_id": 1,
    "university": "МГУ",
    "degree": "Бакалавр",
    "bio": "Студент 4-го курса. Увлекаюсь backend-разработкой.",
    "experience": "Писал курсовые на Go, участвовал в хакатонах.",
    "image": "http://localhost:9000/t-match-storage/user:17:avatar"
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
  "city_id": 1,
  "university": "МГУ",
  "degree": "Бакалавр",
  "bio": "Студент 4-го курса.",
  "experience": "Опыт работы с Go и PostgreSQL."
}
```

**Успешный ответ (200 OK):** пустое тело

> `city_id` — ID города из справочника `GET /cities` (см. [раздел 7](#7-города-cities)).

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
    "user_id": 3,
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
"http://localhost:9000/t-match-storage/user:17:avatar"
```

---

## 4. Публичные профили

*Без авторизации (кроме email — см. особенности)*

### 4.1 Публичный профиль студента

```
GET /profile/:id
```

**Пример:** `GET /profile/17`

**Ответ (200 OK):**
```json
{
  "email": "",
  "profile": {
    "id": 17,
    "user_id": 9,
    "first_name": "Иван",
    "last_name": "Петров",
    "birth_date": "2008-05-15T00:00:00Z",
    "city_id": 1,
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
- Поле `email` обычно пустое
- **Исключение:** если запрос делает авторизованная компания, и этот студент имеет статус `reviewing` или `accepted` хотя бы по одной стажировке этой компании — поле `email` будет заполнено

---

### 4.2 Публичный профиль компании

```
GET /company/profile/:id
```

**Пример:** `GET /company/profile/5`

**Ответ (200 OK):**
```json
{
  "email": "",
  "profile": {
    "id": 5,
    "user_id": 3,
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

**Особенности:**
- Поле `email` обычно пустое
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
  "city_id": 1,
  "duration_months": 3
}
```

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `title` | `string` | ✅ | Название (макс. 200 символов) |
| `description` | `string` | ✅ | Описание (макс. 5000 символов) |
| `salary` | `int` | | Зарплата (≥ 0) |
| `city_id` | `int` | ✅ | ID города из справочника `GET /cities` (≥ 1) |
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
    "city_id": 1,
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
  "city_id": 2,
  "duration_months": 6
}
```

**Успешный ответ (200 OK):** пустое тело

**Важно:** обновить можно только **неархивированную** стажировку. Попытка обновить архивированную вернёт `404`.

---

### 5.4 Архивирование стажировки

```
DELETE /internships/:id
```
*Роль: `company` (только владелец)*

**Успешный ответ (200 OK):** пустое тело

**Примечание:** архивирование — это мягкое удаление (`is_archived = TRUE`). Стажировка перестаёт отображаться в поиске и публичном доступе.

---

### 5.5 Мои стажировки (для компании)

```
GET /my/company/internships
```
*Роль: `company`*

**Ответ (200 OK):** массив стажировок (включая архивированные)

---

### 5.6 Стажировки компании (публично)

```
GET /companies/:id/internships
```
*Без авторизации*

**Ответ (200 OK):** массив **неархивированных** стажировок данной компании

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
| `query` | `string` | — | Текстовый поиск (слова через пробел, логика OR) |
| `city_id` | `int` | — | ID города из справочника `GET /cities` (точное совпадение) |
| `salary_min` | `int` | — | Минимальная зарплата |
| `salary_max` | `int` | — | Максимальная зарплата |
| `duration_min` | `int` | — | Минимальная длительность (месяцы) |
| `duration_max` | `int` | — | Максимальная длительность (месяцы) |
| `skills` | `[]int` | — | ID навыков (AND-логика). Передавать несколько раз: `?skills=1&skills=2` |
| `sort` | `string` | `created_at` | Поле сортировки: `salary`, `duration_months` |
| `order` | `int` | `desc` | Направление: `1` = `ASC`, любое другое значение = `DESC` |
| `limit` | `int` | — | Записей на странице |
| `offset` | `int` | — | Сдвиг для пагинации |

**Примеры запросов:**

```bash
# Текстовый поиск
GET /internships?query=Go+разработчик

# Поиск с фильтрами
GET /internships?query=Python&city_id=1&salary_min=50000&salary_max=150000

# Поиск по навыкам
GET /internships?skills=1&skills=3&skills=5

# Сортировка и пагинация
GET /internships?query=backend&sort=salary&order=1&limit=20&offset=0
```

**Пример ответа (200 OK):**
```json
[
  {
    "id": 15,
    "company_id": 5,
    "title": "Go разработчик",
    "salary": 80000,
    "city_id": 1,
    "is_archived": false,
    "duration_months": 6,
    "created_at": "2026-05-15T10:30:00Z"
  },
  {
    "id": 16,
    "company_id": 8,
    "title": "Python Backend Developer",
    "salary": 70000,
    "city_id": 2,
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
| `query` | `string` | — | Текстовый поиск по имени, фамилии, университету, степени |
| `university` | `string` | — | Университет (частичное совпадение) |
| `skills` | `[]int` | — | ID навыков (AND-логика). `?skills=1&skills=2` |
| `limit` | `int` | — | Записей на странице |
| `offset` | `int` | — | Сдвиг для пагинации |

**Пример:** `GET /students?query=Иван&university=МГУ&skills=1&skills=2`

**Пример ответа (200 OK):**
```json
[
  {
    "id": 17,
    "user_id": 9,
    "first_name": "Иван",
    "last_name": "Петров",
    "city_id": 1,
    "university": "МГУ",
    "degree": "Computer Science",
    "image": null
  },
  {
    "id": 18,
    "user_id": 12,
    "first_name": "Мария",
    "last_name": "Сидорова",
    "city_id": 2,
    "university": "МГУ",
    "degree": "Бакалавр",
    "image": "http://localhost:9000/t-match-storage/user:18:avatar"
  }
]
```

**Важно:** Это **превью** без чувствительных данных (нет `birth_date`, `bio`, `experience`). Полный профиль — через `GET /profile/:id`.

**Поле `user_id`:** это идентификатор пользователя в таблице `users`. Именно его нужно передавать в `POST /internships/:id/invite` для приглашения стажёра (НЕ `id` из таблицы `interns`).

---

### 6.3 Поиск компаний (превью)

```
GET /companies
```
*Без авторизации*

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `query` | `string` | — | Текстовый поиск по названию, описанию, адресу |
| `city_id` | `int` | — | ID города из справочника `GET /cities` (поиск по юр. адресу) |
| `limit` | `int` | — | Записей на странице |
| `offset` | `int` | — | Сдвиг для пагинации |

**Пример:** `GET /companies?query=IT&city_id=1&limit=10`

**Пример ответа (200 OK):**
```json
[
  {
    "id": 5,
    "user_id": 3,
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
    "user_id": 7,
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

**Важно:** Это **превью**. Полный профиль компании (с `kpp`, `director_name`) — через `GET /company/profile/:id`.

---

## 7. Города (Cities)

### 7.1 Список всех городов

```
GET /cities
```
*Без авторизации*

Справочник городов для выпадающих списков на фронтенде. Используется для полей `city_id` в профиле, стажировках и поиске.

**Пример ответа (200 OK):**
```json
[
  {"id": 1, "name": "Москва", "region": "Москва"},
  {"id": 2, "name": "Санкт-Петербург", "region": "Санкт-Петербург"},
  {"id": 3, "name": "Новосибирск", "region": "Новосибирская область"},
  {"id": 4, "name": "Белогорск", "region": "Амурская область"},
  {"id": 5, "name": "Белогорск", "region": "Республика Крым"}
]
```

| Поле | Тип | Описание |
| :--- | :--- | :--- |
| `id` | `int` | ID города (передаётся в `city_id`) |
| `name` | `string` | Название города |
| `region` | `string` | Регион (для отображения города-тёзки: «Белогорск, Республика Крым») |

**Особенности:**
- Все города уникальны по паре `(name, region)` — города-тёзки из разных регионов различаются через `region`
- Ответ отсортирован по названию города
- Фронтенд хранит только `city_id`, а имя города берёт из этого справочника

---

## 8. Навыки (Skills)

### 8.1 Список всех навыков

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

### 8.2 Управление навыками стажёра

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

### 8.3 Управление навыками стажировки

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

## 9. Отклики на стажировки

### 9.1 Откликнуться на стажировку

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

### 9.2 Отменить отклик

```
DELETE /internships/:id/respond
```
*Роль: `intern`*

**Успешный ответ (200 OK):** пустое тело

**Примечание:** удаляется только отклик со статусом `pending`.

---

### 9.3 Мои отклики (для стажёра)

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

### 9.4 Отклики на стажировку (для компании)

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

### 9.5 Изменение статуса отклика

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

**Событие:** при смене статуса студенту отправляется уведомление (через WebSocket и в список `/my/notifications`).

---

### 9.6 Приглашение стажера на стажировку

```
POST /internships/:id/invite
```
*Роль: `company` (владелец стажировки)*

**Тело запроса:**
```json
{
  "user_id": 17,
  "message": "Приглашаем вас на собеседование!"
}
```

| Параметр | Тип | Обязательный | Описание |
| :--- | :--- | :---: | :--- |
| `user_id` | `int` | ✅ | ID пользователя-стажёра |
| `message` | `string` | | Сопроводительное сообщение |

**Успешный ответ (200 OK):** пустое тело

**Событие:** студенту отправляется уведомление типа `invate`.

---

## 10. Уведомления

### 10.1 Получение уведомлений

```
GET /my/notifications
```
*Требуется авторизация*

**Пример ответа (200 OK):**
```json
[
  {
    "id": 1,
    "user_id": 17,
    "type": "change_status",
    "is_read": false,
    "created_at": "2025-05-09T10:30:00Z",
    "data": {
      "id": 1,
      "notification_id": 1,
      "internship_id": 42,
      "company_id": 5,
      "new_status": "accepted"
    }
  },
  {
    "id": 2,
    "user_id": 17,
    "type": "invate",
    "is_read": false,
    "created_at": "2025-05-10T14:00:00Z",
    "data": {
      "id": 1,
      "notification_id": 2,
      "internship_id": 42,
      "company_id": 5,
      "message": "Приглашаем вас на собеседование!"
    }
  }
]
```

---

### 10.2 Отметить все уведомления прочитанными

```
PUT /my/notifications
```
*Требуется авторизация*

**Успешный ответ (200 OK):** пустое тело

---

### 10.3 WebSocket-уведомления

```
GET /ws/notifications
```

**Авторизация:**
- Браузеры не умеют устанавливать заголовки на WebSocket-соединениях, поэтому токен передаётся query-параметром: `GET /ws/notifications?token=<access_token>`
- Сервер также принимает стандартный заголовок `Authorization: Bearer <token>` (приоритетнее query-параметра)

**Протокол:**
- Подключение устанавливает постоянное WebSocket-соединение
- Клиент должен отправлять `ping` — сервер ответит `pong` (обычное текстовое сообщение) для поддержания соединения
- При каждом событии сервер отправляет **сырое текстовое сообщение** (WebSocket `TextMessage`) с JSON уведомления — без дополнительного JSON-кодирования

**Формат входящего сообщения:** строка с JSON уведомления (см. пример выше)

---

## 11. Система токенов

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

## 12. Rate Limiting

Лимиты в запросах в минуту на IP/пользователя:

### Аутентификация
| Эндпоинт | Лимит |
| :--- | :--- |
| `POST /auth/students` | 20 |
| `POST /auth/students/verify` | 60 |
| `POST /auth/newverify` | 4 |
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
| `POST /internships/:id/invite` | 40 |

### Отклики
| Эндпоинт | Лимит |
| :--- | :--- |
| `POST /internships/:id/respond` | 10 |
| `DELETE /internships/:id/respond` | 10 |
| `GET /my/responses` | 20 |
| `GET /internships/:id/responses` | 20 |
| `PUT /responses/:id/status` | 20 |

### Поиск
| Эндпоинт | Лимит |
| :--- | :--- |
| `GET /internships` | 60 |
| `GET /students` | 60 |
| `GET /companies` | 60 |

### Админка
| Эндпоинт | Лимит |
| :--- | :---: |
| `POST /admin/login` | 10 |
| `GET /admin/stats` | 30 |
| `PATCH /admin/users/:id/ban` | 10 |
| `DELETE /admin/users/:id/ban` | 10 |
| `DELETE /admin/internships/:id` | 10 |

### Без лимита
| Эндпоинт |
| :--- |
| `GET /internships/:id` |
| `GET /profile/:id` |
| `GET /company/profile/:id` |
| `GET /skills` |
| `GET /cities` |
| `GET /my/company/internships` |
| `GET /my/notifications` |
| `PUT /my/notifications` |
| `GET /ws/notifications` |
| `OPTIONS /*path` |

---

## 13. Рекомендательная система

Отдельный микросервис на Python (FastAPI), который ранжирует стажировки для стажёра по гео-близости, схожести навыков и (при наличии данных) ALS-рекомендациям из коллаборативной фильтрации. Бэкенд автоматически синхронизирует данные в recsys и собирает действия пользователей.

### 13.1 Синхронизация данных

Бэкенд автоматически отправляет в recsys события о следующих действиях:

| Событие | Когда | Тип действия |
| :--- | :--- | :--- |
| Создание пользователя (intern) | Подтверждение email `POST /auth/students/verify` | `create_user` |
| Обновление гео стажёра | Обновление профиля `PUT /my/profile` (если указан `city_id`) | `update_user_geo` |
| Создание стажировки | `POST /internships` | `create_internship` |
| Обновление гео стажировки | `PUT /internships/:id` (если указан `city_id`) | `update_internship_geo` |
| Архивация стажировки | `DELETE /internships/:id` (компания) | `delete_internship` |
| Удаление стажировки | `DELETE /admin/internships/:id` (админ) | `delete_internship` |
| Навыки стажёра | `POST/DELETE /my/profile/skills` | `add/delete_user_skill` |
| Навыки стажировки | `POST/DELETE /internships/:id/skill` | `add/delete_internship_skill` |

### 13.2 События пользователя (actions)

| Тип действия | Когда отправляется |
| :--- | :--- |
| `click` | `POST /internships/:id/view` — просмотр карточки стажировки стажёром |
| `apply` | `POST /internships/:id/respond` — отклик на стажировку |
| `invate` | `POST /internships/:id/invite` — приглашение стажёра компанией |

### 13.3 Рекомендации для стажёра

```
GET /my/recommendations
```
*Роль: `intern`* | Rate limit: 60 запросов/мин

Возвращает ранжированный список стажировок (короткие карточки, тот же формат, что в поиске `GET /internships`), отсортированный по убыванию рекомендательного балла recsys:

```json
[
  {
    "id": 15,
    "company_id": 2,
    "title": "Go developer",
    "salary": 100000,
    "duration_months": 6,
    "city_id": 77,
    "created_at": "2026-08-19T10:00:00Z",
    "is_archived": false
  }
]
```

| Поле | Описание |
| :--- | :--- |
| `id` | ID стажировки |
| `company_id` | ID компании-работодателя |
| `title` | Название |
| `salary` | Зарплата |
| `duration_months` | Длительность в месяцах |
| `city_id` | ID города |
| `created_at` | Дата создания |
| `is_archived` | Флаг архивации (всегда `false`) |

**Примечание:** бэкенд сам запрашивает рейтинги у recsys и подставляет карточки стажировок из БД, сохраняя порядок ранжирования. Архивные стажировки и стажировки забаненных компаний в ответ не попадают. Внутренние поля recsys (`score`, `geo_similarity`, `als_score` и т.д.) фронту не отдаются. Если recsys недоступен — ответ `502`.

### 13.4 Отслеживание просмотра

```
POST /internships/:id/view
```
*Роль: `intern`*

Фиксирует клик по карточке стажировки. Используется для пересчёта ALS-рекомендаций. **Важно:** публичный `GET /internships/:id` клики не собирает — для этого нужен авторизованный запрос.

### 13.5 Конфигурация и запуск

Сервис recsys настраивается через переменную окружения `RECSYS_URL` (по умолчанию `http://recsys:8000`).

```env
# .env
RECSYS_URL=http://recsys:8000
```

**Запуск:** бэкенд и recsys — два независимых Docker-проекта. Каждый поднимается отдельно, а взаимодействуют они через общую внешнюю Docker-сеть `t-match-net`:

1. Создать общую сеть (один раз):
   ```bash
   docker network create t-match-net
   ```

2. Запустить бэкенд (из корня этого репозитория):
   ```bash
   docker compose up -d --build
   ```

3. Запустить recsys (из репозитория `recsys`, вместе со своими PostgreSQL и Redis):
   ```bash
   docker compose up -d --build
   ```

Бэкенд обращается к recsys по адресу `http://recsys:8000` — имя `recsys` резолвится через сетевой алиас в `t-match-net`. Проекты можно останавливать, перезапускать и обновлять независимо друг от друга; бэкенд работает и без recsys (см. поведение при недоступном сервисе в п. 13.3).

---

## 14. Обработка ошибок

Все ошибки возвращаются в формате plain text с соответствующим HTTP-статусом.

> **Исключение:** ответ забаненного пользователя (статус `409`) возвращается в JSON-формате `{"reason": "<причина>"}` — это единственный случай, когда тело ошибки не plain text.

| Код | Сообщение | Когда возникает |
| :--- | :--- | :--- |
| `400` | `Bad request` | Невалидный JSON, неверный формат данных |
| `400` | `Admins cannot be banned` | Попытка заблокировать администратора |
| `401` | `User Unauthorized` | Токен отсутствует, истёк или невалиден |
| `403` | `Access denied: insufficient permissions` | Нет прав на действие |
| `404` | `Not found` | Ресурс не найден |
| `404` | `User not found` | Пользователь не найден |
| `409` | `User with this email already exists` | Email занят |
| `409` | `You have already responded to this internship` | Повторный отклик |
| `409` | `User is already banned` | Попытка повторной блокировки |
| `409` | `User not baned` | Попытка разблокировать незаблокированного |
| `409` | `User is banned` (тело: `{"reason": "..."}`) | Доступ заблокированного пользователя |
| `410` | `Internship is archived` | Стажировка в архиве |
| `422` | `User must be at least 16 years old` | Возраст < 16 лет |
| `429` | `Too many invalid attempts` | Превышен rate limit |
| `500` | `Internal server error` | Внутренняя ошибка сервера |
| `502` | `External service temporarily unavailable` | Ошибка DaData или недоступен recsys |
| `503` | `Cache service temporarily unavailable` | Redis недоступен |
| `503` | `Failed to send email, please try again` | Ошибка отправки email |

---

## 15. Админ-панель

*Все эндпоинты требуют роль `admin`.*

### 15.1 Вход администратора

```
POST /admin/login
```

**Тело запроса:**
```json
{
  "email": "admin@tmatch.space",
  "password": "AdminPass1",
  "device_id": "web_admin_001"
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

*Примечание:* аккаунт администратора создаётся вручную в БД (таблица `users` с ролью `admin`, опционально запись в `admins`). API регистрации админов не существует.

---

### 15.2 Статистика

```
GET /admin/stats
```

**Успешный ответ (200 OK):**
```json
{
  "total_interns": 120,
  "total_companies": 18,
  "total_internships": 45,
  "total_responses": 310,
  "responses_pending": 40,
  "responses_reviewing": 20,
  "responses_accepted": 150,
  "responses_rejected": 100,
  "new_users_7_days": 12,
  "new_internships_7_days": 5,
  "new_responses_7_days": 30,
  "users_online": 8
}
```

| Поле | Описание |
| :--- | :--- |
| `users_online` | Пользователи с активным WebSocket-подключением к `/ws/notifications` |

---

### 15.3 Блокировка пользователя

```
PATCH /admin/users/:id/ban
```

| Параметр | Описание |
| :--- | :--- |
| `:id` | **`users.id`** — идентификатор пользователя (НЕ `interns.id` / `companies.id`) |

**Тело запроса:**
```json
{
  "reason": "Спам-рассылка"
}
```

*`reason` — обязателен (1–500 символов).*

**Успешный ответ (200 OK):** пустое тело

**Что происходит при блокировке:**
1. Запись в таблицу `user_bans`
2. Все refresh-токены пользователя удаляются из Redis (все устройства разлогиниваются)
3. WebSocket-подключение к `/ws/notifications` разрывается
4. Любой запрос забаненного пользователя возвращает `409` с причиной блокировки

**Ответ забаненного пользователя на любой запрос (409 Conflict):**
```json
{
  "reason": "Спам-рассылка"
}
```

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `400` | Админ пытается заблокировать сам себя |
| `400` | «Admins cannot be banned» — нельзя блокировать другого администратора |
| `404` | Пользователь не найден |
| `409` | Пользователь уже заблокирован |

---

### 15.4 Разблокировка пользователя

```
DELETE /admin/users/:id/ban
```

**Успешный ответ (200 OK):** пустое тело

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `404` | Пользователь не найден |
| `409` | Пользователь не был заблокирован |

---

### 15.5 Удаление стажировки

```
DELETE /admin/internships/:id
```

**Внимание:** жёсткое удаление (`DELETE FROM internships`). Каскадно удаляются отклики, навыки и уведомления, связанные со стажировкой. В отличие от архивации (`DELETE /internships/:id` для компании), восстановить стажировку нельзя.

**Успешный ответ (200 OK):** пустое тело

**Возможные ошибки:**
| Статус | Описание |
| :--- | :--- |
| `404` | Стажировка не найдена |

---

### 15.6 Rate Limiting

| Эндпоинт | Лимит (в минуту на IP) |
| :--- | :---: |
| `POST /admin/login` | 10 |
| `GET /admin/stats` | 30 |
| `PATCH /admin/users/:id/ban` | 10 |
| `DELETE /admin/users/:id/ban` | 10 |
| `DELETE /admin/internships/:id` | 10 |

---

## 16. Запуск приложения

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
DB_USER=realworld
DB_PASSWORD=your_db_password
DB_SSLMODE=disable

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=:8080
# Ставить true в проде (HTTPS): refresh-кука получит флаг Secure
COOKIE_SECURE=false

# Redis
REDIS_ADDR=redis:6379
REDIS_DB=0
REDIS_PASSWORD=
REDIS_MAX_RETRIES=3
REDIS_DIAL_TIMEOUT=5s
REDIS_TIMEOUT=3s

# Email
# Для разработки (MailHog поднимается только в docker-compose.dev.yml).
# В проде (docker-compose.yml) MailHog отсутствует — укажите реальный SMTP-сервер.
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
```

#### 3. Где взять ключи

| Переменная | Где получить |
| :--- | :--- |
| `DA_DATA_API_KEY` | Зарегистрироваться на [dadata.ru](https://dadata.ru) → API → Ключи доступа |
| `JWT_SECRET` | Сгенерировать: `openssl rand -base64 32` |
| `DB_PASSWORD` | Придумать самостоятельно |

#### Примечания:

Email: если `EMAIL_PASSWORD` не задан, автоматически используется MailHog — все письма перехватываются и отображаются в веб-интерфейсе `http://localhost:8025`, реальным пользователям ничего не отправляется.

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
