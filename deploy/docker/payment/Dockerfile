# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.24-alpine AS builder

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum файлы для кеширования зависимостей
COPY backend/payment/go.mod backend/payment/go.sum ./payment/
COPY backend/shared/go.mod backend/shared/go.sum ./shared/
COPY backend/platform/go.mod backend/platform/go.sum ./platform/

# Загружаем зависимости
WORKDIR /app/payment
RUN go mod download

# Возвращаемся в корневую директорию и копируем исходный код
WORKDIR /app
COPY backend/payment/ ./payment/
COPY backend/shared/ ./shared/
COPY backend/platform/ ./platform/

# Собираем grpc-health-probe — утилиту для проверки состояния gRPC-сервиса. Через Go proxy, а не ADD
# с GitHub Releases: релизы периодически отвечают 504 и ломают сборку.
RUN CGO_ENABLED=0 GOBIN=/app go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.37

# Собираем приложение с оптимизациями
WORKDIR /app/payment
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o payment-service ./cmd/main.go

# Финальный минимальный образ
FROM scratch

# Копируем SSL сертификаты для HTTPS запросов
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем /etc/passwd, чтобы UID 65534 (nobody) разрешался в имя пользователя
COPY --from=builder /etc/passwd /etc/passwd

# Копируем бинарный файл
COPY --from=builder /app/payment/payment-service /payment-service

# grpc-health-probe для healthcheck в docker compose (в scratch-образе нет shell и curl)
COPY --from=builder /app/grpc-health-probe /grpc-health-probe

# Открываем порт для gRPC
EXPOSE 50052

# Запускаем от непривилегированного пользователя nobody (UID 65534).
# /etc/passwd нужен обязательно: OpenTelemetry определяет текущего пользователя через user.Current(),
# а в статическом бинарнике (CGO_ENABLED=0) без записи в passwd он падает на старте.
USER 65534:65534

ENTRYPOINT ["/payment-service"]
