#!/usr/bin/env bash
# Готовит пользователя, сессию и деталь, затем запускает нагрузочный сценарий.
# Требует поднятый стек (task up-all) и bin/grpcurl (task grpcurl:install).
# Параметры передаются в main.go: ./scripts/loadtest/run.sh -workers 50 -duration 20s -pay 20
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GRPCURL="$ROOT/bin/grpcurl"
LOGIN="load_$(date +%s)"
PASSWORD='TestP@ss123!'

REG=$("$GRPCURL" -plaintext -d "{\"info\":{\"userInfo\":{\"login\":\"$LOGIN\",\"email\":\"$LOGIN@example.com\",\"notificationMethods\":[]},\"password\":\"$PASSWORD\"}}" localhost:50053 iam.v1.UserService/Register)
USER_UUID=$(echo "$REG" | grep -o '"userUuid": "[^"]*' | cut -d'"' -f4)

SESSION=$("$GRPCURL" -plaintext -d "{\"login\":\"$LOGIN\",\"password\":\"$PASSWORD\"}" localhost:50053 iam.v1.AuthService/Login \
  | grep -o '"sessionUuid": "[^"]*' | cut -d'"' -f4)

PART=$("$GRPCURL" -plaintext -H "session-uuid: $SESSION" -d '{"filter":{}}' localhost:50051 inventory.v1.InventoryService/ListParts \
  | grep -o '"partUuid": "[^"]*' | head -1 | cut -d'"' -f4)

exec go run "$ROOT/scripts/loadtest/main.go" -session "$SESSION" -user "$USER_UUID" -part "$PART" "$@"
