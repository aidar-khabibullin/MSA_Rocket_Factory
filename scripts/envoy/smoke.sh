#!/usr/bin/env bash
# Smoke-тест Envoy: проверяет маршруты, ошибки, CORS и адрес клиента на поднятом стеке (task up-all).
# Запуск: task envoy:test
set -uo pipefail

ENVOY_URL="${ENVOY_URL:-http://localhost:8085}"
ORIGIN="http://localhost:5173"
# Ограничитель входа в iam: 5 попыток в минуту на адрес (backend/iam/internal/app/app.go).
LOGIN_WINDOW_SECONDS="${LOGIN_WINDOW_SECONDS:-61}"

FAILED=0
BODY="$(mktemp)"
HEADERS="$(mktemp)"
trap 'rm -f "$BODY" "$HEADERS"' EXIT

pass() { echo "  ✅ $1"; }
fail() { echo "  ❌ $1"; FAILED=1; }

# request METHOD URL [curl-аргументы] -> в $STATUS код ответа, в $BODY тело, в $HEADERS заголовки
request() {
  STATUS=$(curl -s -D "$HEADERS" -o "$BODY" -w '%{http_code}' -X "$@")
}

# expect_status ОПИСАНИЕ ОЖИДАЕМЫЙ_КОД
expect_status() {
  if [ "$STATUS" = "$2" ]; then pass "$1: HTTP $STATUS"; else fail "$1: ожидали $2, получили $STATUS ($(cat "$BODY"))"; fi
}

# expect_body ОПИСАНИЕ РЕГУЛЯРКА
expect_body() {
  if grep -Eq "$2" "$BODY"; then pass "$1"; else fail "$1: в ответе нет '$2' ($(cat "$BODY"))"; fi
}

# expect_header ОПИСАНИЕ РЕГУЛЯРКА (без учёта регистра, по заголовкам ответа)
expect_header() {
  if grep -Eiq "$2" "$HEADERS"; then pass "$1"; else fail "$1: в заголовках нет '$2' ($(tr -d '\r' <"$HEADERS" | tr '\n' ';'))"; fi
}

# uuid_of ИМЯ_ПОЛЯ -> значение поля-uuid из $BODY
uuid_of() {
  grep -Eo "\"$1\"[[:space:]]*:[[:space:]]*\"[0-9a-f-]{36}\"" "$BODY" | head -1 | grep -Eo '[0-9a-f-]{36}'
}

JSON=(-H 'Content-Type: application/json')

# login ЛОГИН ПАРОЛЬ -> $STATUS; если сработал ограничитель (повторный запуск в пределах минуты), ждём окно и повторяем
login() {
  request POST "$ENVOY_URL/api/auth/login" "${JSON[@]}" -d "{\"login\":\"$1\",\"password\":\"$2\"}"
  if [ "$STATUS" = "429" ]; then
    echo "  ⏳ Ограничитель входа ещё помнит прошлый запуск, ждём ${LOGIN_WINDOW_SECONDS}с"
    sleep "$LOGIN_WINDOW_SECONDS"
    request POST "$ENVOY_URL/api/auth/login" "${JSON[@]}" -d "{\"login\":\"$1\",\"password\":\"$2\"}"
  fi
}

echo "🧪 Smoke-тест Envoy ($ENVOY_URL)"

STAMP="$(date +%s)"
LOGIN="smoke_$STAMP"
EMAIL="smoke_$STAMP@example.com"
PASSWORD="TestP@ss123!"

echo "📝 Регистрация"
request POST "$ENVOY_URL/api/auth/register" "${JSON[@]}" -d "{\"info\":{\"user_info\":{\"login\":\"$LOGIN\",\"email\":\"$EMAIL\"},\"password\":\"$PASSWORD\"}}"
expect_status "POST /api/auth/register" 200
expect_body "ответ содержит user_uuid" '"user_uuid"[[:space:]]*:[[:space:]]*"[0-9a-f-]{36}"'
USER_UUID="$(uuid_of user_uuid)"
if [ -z "$USER_UUID" ]; then
  echo "❌ Не удалось зарегистрировать пользователя, дальше идти нет смысла"
  exit 1
fi

request POST "$ENVOY_URL/api/auth/register" "${JSON[@]}" -d "{\"info\":{\"user_info\":{\"login\":\"$LOGIN\",\"email\":\"$EMAIL\"},\"password\":\"$PASSWORD\"}}"
expect_status "повторная регистрация того же логина" 409
expect_body "в теле ошибки есть message" '"message"'

request POST "$ENVOY_URL/api/auth/register" "${JSON[@]}" -d '{"info":'
expect_status "регистрация с битым JSON" 400

echo "🔑 Вход и текущая сессия"
login "$LOGIN" "$PASSWORD"
expect_status "POST /api/auth/login" 200
expect_body "ответ содержит session_uuid (snake_case)" '"session_uuid"[[:space:]]*:[[:space:]]*"[0-9a-f-]{36}"'
SESSION="$(uuid_of session_uuid)"

request GET "$ENVOY_URL/api/auth/me" -H "Session-Uuid: $SESSION"
expect_status "GET /api/auth/me с сессией" 200
expect_body "ответ содержит логин пользователя" "\"login\"[[:space:]]*:[[:space:]]*\"$LOGIN\""

echo "🔩 Каталог"
# Первое обращение inventory к Mongo после простоя может упасть на устаревшем соединении (воспроизводится и
# напрямую по gRPC, к Envoy не относится), поэтому первый запрос повторяем.
for attempt in 1 2 3; do
  request GET "$ENVOY_URL/api/parts" -H "Session-Uuid: $SESSION"
  [ "$STATUS" != "504" ] && break
  echo "  ⏳ inventory не ответил за таймаут (попытка $attempt из 3), повторяем"
done
expect_status "GET /api/parts без фильтра" 200
expect_body "в ответе есть детали" '"part_uuid"[[:space:]]*:[[:space:]]*"[0-9a-f-]{36}"'
expect_body "int64 приходит строкой" '"stock_quantity"[[:space:]]*:[[:space:]]*"[0-9]+"'
PART_UUID="$(uuid_of part_uuid)"

request GET "$ENVOY_URL/api/parts?filter.category=CATEGORY_ENGINE" -H "Session-Uuid: $SESSION"
expect_status "GET /api/parts с фильтром по категории" 200
if grep -Eq '"category"[[:space:]]*:[[:space:]]*"CATEGORY_(FUEL|HULL|WING|UNSPECIFIED)"' "$BODY"; then
  fail "фильтр по категории пропустил детали других категорий"
else
  pass "фильтр по категории применён"
fi

request GET "$ENVOY_URL/api/parts"
expect_status "GET /api/parts без сессии" 401
expect_body "в теле ошибки есть message" '"message"'

echo "📦 Заказы через /api/v1"
request POST "$ENVOY_URL/api/v1/orders" "${JSON[@]}" -d "{\"user_uuid\":\"$USER_UUID\",\"part_uuids\":[\"$PART_UUID\"]}"
expect_status "создание заказа без сессии" 401

request POST "$ENVOY_URL/api/v1/orders" "${JSON[@]}" -H "Session-Uuid: $SESSION" -d "{\"user_uuid\":\"$USER_UUID\",\"part_uuids\":[\"$PART_UUID\"]}"
expect_status "POST /api/v1/orders" 200
ORDER_UUID="$(uuid_of order_uuid)"
if [ -z "$ORDER_UUID" ]; then
  echo "❌ Заказ не создан, проверки заказов пропущены"
  FAILED=1
else
  request GET "$ENVOY_URL/api/v1/orders/$ORDER_UUID" -H "Session-Uuid: $SESSION"
  expect_status "GET /api/v1/orders/{uuid}" 200

  request POST "$ENVOY_URL/api/v1/orders/$ORDER_UUID/pay" "${JSON[@]}" -H "Session-Uuid: $SESSION" -d '{"payment_method":"CARD"}'
  expect_status "оплата заказа" 200

  request POST "$ENVOY_URL/api/v1/orders/$ORDER_UUID/pay" "${JSON[@]}" -H "Session-Uuid: $SESSION" -d '{"payment_method":"CARD"}'
  expect_status "повторная оплата" 409

  request POST "$ENVOY_URL/api/v1/orders/$ORDER_UUID/cancel" -H "Session-Uuid: $SESSION"
  expect_status "отмена оплаченного заказа" 400
fi

request POST "$ENVOY_URL/api/v1/orders" "${JSON[@]}" -H "Session-Uuid: $SESSION" -d "{\"user_uuid\":\"$USER_UUID\",\"part_uuids\":[\"$PART_UUID\"]}"
CANCEL_UUID="$(uuid_of order_uuid)"
if [ -n "$CANCEL_UUID" ]; then
  request POST "$ENVOY_URL/api/v1/orders/$CANCEL_UUID/cancel" -H "Session-Uuid: $SESSION"
  expect_status "отмена неоплаченного заказа" 204
else
  fail "не удалось создать второй заказ для проверки отмены ($(cat "$BODY"))"
fi

request GET "$ENVOY_URL/api/v1/orders/11111111-1111-1111-1111-111111111111" -H "Session-Uuid: $SESSION"
expect_status "несуществующий заказ" 404
expect_body "в теле ошибки есть message" '"message"'

request POST "$ENVOY_URL/api/v1/orders" "${JSON[@]}" -H "Session-Uuid: $SESSION" -d '{"user_uuid":"not-a-uuid","part_uuids":[]}'
expect_status "создание заказа с неверным телом" 400

echo "🌐 CORS"
request OPTIONS "$ENVOY_URL/api/auth/login" -H "Origin: $ORIGIN" -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: content-type,session-uuid'
expect_status "preflight /api/auth/login" 200
expect_header "разрешён origin Vite" "^access-control-allow-origin: $ORIGIN"
expect_header "разрешён метод POST" '^access-control-allow-methods:.*POST'
expect_header "разрешены Content-Type и Session-Uuid" '^access-control-allow-headers:.*Content-Type.*Session-Uuid'

request OPTIONS "$ENVOY_URL/api/v1/orders" -H "Origin: $ORIGIN" -H 'Access-Control-Request-Method: POST'
expect_status "preflight /api/v1/orders" 200
expect_header "разрешён origin Vite" "^access-control-allow-origin: $ORIGIN"

request OPTIONS "$ENVOY_URL/api/parts" -H 'Origin: http://evil.example' -H 'Access-Control-Request-Method: GET'
if grep -Eiq '^access-control-allow-origin:' "$HEADERS"; then fail "чужой origin получил разрешение"; else pass "чужой origin не получает разрешения"; fi

request GET "$ENVOY_URL/api/parts" -H "Session-Uuid: $SESSION" -H "Origin: $ORIGIN"
expect_header "обычный ответ несёт access-control-allow-origin" "^access-control-allow-origin: $ORIGIN"

echo "🚫 Ошибки входа"
login "$LOGIN" "wrong-password"
expect_status "вход с неверным паролем" 401
expect_body "в теле ошибки есть message" '"message"'

request GET "$ENVOY_URL/api/auth/me"
expect_status "GET /api/auth/me без сессии" 400
expect_body "в теле ошибки есть message" '"message"'

# Последним: тест исчерпывает лимит попыток входа для адреса этой машины на минуту.
echo "🛡️  Ограничитель попыток входа"
BLOCKED=0
LAST=""
for i in 1 2 3 4 5 6; do
  # Каждый раз подставляем другой поддельный адрес: если бы Envoy верил заголовку, лимит ни разу не сработал бы.
  request POST "$ENVOY_URL/api/auth/login" "${JSON[@]}" -H "X-Forwarded-For: 203.0.113.$i" -d "{\"login\":\"$LOGIN\",\"password\":\"wrong-$i\"}"
  LAST="$STATUS"
  [ "$STATUS" = "429" ] && BLOCKED=$((BLOCKED + 1))
done
if [ "$BLOCKED" -gt 0 ] && [ "$LAST" = "429" ]; then
  pass "шесть неверных входов с разными поддельными X-Forwarded-For заблокированы ($BLOCKED из 6 получили 429)"
  expect_body "в теле блокировки есть message" '"message"'
else
  fail "ограничитель не сработал: поддельный X-Forwarded-For обходит лимит (последний статус $LAST, блокировок $BLOCKED)"
fi

if [ "$FAILED" -ne 0 ]; then
  echo "❌ Smoke-тест Envoy не пройден"
  exit 1
fi
echo "🎉 Smoke-тест Envoy пройден"
