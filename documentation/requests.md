**Регистрация пользователя:**

## Регистрация

**Request:**
POST /auth/users
Content-Type: application/json
{
    "email": "user@example.com",
    "password": "123456"
}

**Response 201 Created**
Content-Type: application/json
Token: eyJhbGciOiJIUzI1NiIs...
{
    ...
}

## Обновление токена

API автоматически обновляет access-токен при его истечении.

**Механизм работы:**
1. Клиент отправляет запрос с истекшим или отсутствующим `access_token`
2. Сервер проверяет `refresh_token` в cookies
3. При успехе сервер возвращает новый `access_token` в заголовке `New-Access-Token`
