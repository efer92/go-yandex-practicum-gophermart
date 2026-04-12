# API Examples

---

## 1. Регистрация пользователя

```bash
curl -i -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login":"andrey","password":"secret123"}'
```

**Успешный ответ — 200:**
```
HTTP/1.1 200 OK
Authorization: Bearer <jwt-token>
Set-Cookie: token=<jwt-token>; HttpOnly
```

**Логин уже занят — 409:**
```
HTTP/1.1 409 Conflict
login already taken
```

---

## 2. Аутентификация пользователя

```bash
curl -i -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login":"andrey","password":"secret123"}'
```

**Успешный ответ — 200:**
```
HTTP/1.1 200 OK
Authorization: Bearer <jwt-token>
Set-Cookie: token=<jwt-token>; HttpOnly
```

Сохраните токен в переменную для последующих запросов:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login":"andrey","password":"secret123"}' \
  -D - | grep -i authorization | awk '{print $3}' | tr -d '\r')
```

---

## 3. Загрузка номера заказа

Номер заказа должен быть валидным по алгоритму Луна.

```bash
curl -i -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d '12345678903'
```

| Код | Описание |
|-----|----------|
| 202 | Заказ принят в обработку |
| 200 | Заказ уже был загружен этим пользователем |
| 409 | Заказ уже загружен другим пользователем |
| 422 | Неверный формат номера заказа (не проходит проверку Луна) |
| 401 | Пользователь не аутентифицирован |

---

## 4. Список загруженных заказов

```bash
curl -s http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN" | jq .
```

**Успешный ответ — 200:**
```json
[
  {
    "number": "12345678903",
    "status": "PROCESSED",
    "accrual": 100.5,
    "uploaded_at": "2024-01-15T10:30:00Z"
  },
  {
    "number": "4561261212345467",
    "status": "NEW",
    "uploaded_at": "2024-01-15T11:00:00Z"
  }
]
```

Возможные статусы: `NEW`, `PROCESSING`, `PROCESSED`, `INVALID`.

**Нет заказов — 204** (пустое тело).

---

## 5. Текущий баланс

```bash
curl -s http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer $TOKEN" | jq .
```

**Успешный ответ — 200:**
```json
{
  "current": 85.5,
  "withdrawn": 15.0
}
```

- `current` — доступный остаток баллов
- `withdrawn` — итого списано за всё время

---

## 6. Списание баллов в счёт оплаты заказа

```bash
curl -i -X POST http://localhost:8080/api/user/balance/withdraw \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"order":"2377225624","sum":15.0}'
```

| Код | Описание |
|-----|----------|
| 200 | Средства успешно списаны |
| 402 | Недостаточно средств на счёте |
| 422 | Неверный формат номера заказа |
| 401 | Пользователь не аутентифицирован |

---

## 7. История списаний

```bash
curl -s http://localhost:8080/api/user/withdrawals \
  -H "Authorization: Bearer $TOKEN" | jq .
```

**Успешный ответ — 200:**
```json
[
  {
    "order": "2377225624",
    "sum": 15.0,
    "processed_at": "2024-01-15T12:00:00Z"
  }
]
```

**Нет списаний — 204** (пустое тело).

---

## Полный сценарий

```bash
# 1. Регистрация
curl -s -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login":"andrey","password":"secret123"}'

# 2. Логин и сохранение токена
TOKEN=$(curl -s -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login":"andrey","password":"secret123"}' \
  -D - | grep -i authorization | awk '{print $3}' | tr -d '\r')

# 3. Загрузить заказ
curl -s -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d '12345678903'

# 4. Подождать обработки, проверить статус и начисления
curl -s http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN" | jq .

# 5. Проверить баланс
curl -s http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer $TOKEN" | jq .

# 6. Списать баллы в счёт нового заказа
curl -s -X POST http://localhost:8080/api/user/balance/withdraw \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"order":"2377225624","sum":15.0}'

# 7. Посмотреть историю списаний
curl -s http://localhost:8080/api/user/withdrawals \
  -H "Authorization: Bearer $TOKEN" | jq .
```
