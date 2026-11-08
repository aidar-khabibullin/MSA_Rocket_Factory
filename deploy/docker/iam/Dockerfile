# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.24-alpine AS builder

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum файлы для кеширования зависимостей
COPY backend/iam/go.mod backend/iam/go.sum ./iam/
COPY backend/shared/go.mod backend/shared/go.sum ./shared/
COPY backend/platform/go.mod backend/platform/go.sum ./platform/

# Загружаем зависимости
WORKDIR /app/iam
RUN go mod download

# Возвращаемся в корневую директорию и копируем исходный код
WORKDIR /app
COPY backend/iam/ ./iam/
COPY backend/shared/ ./shared/
COPY backend/platform/ ./platform/

# Собираем grpc-health-probe — утилиту для проверки состояния gRPC-сервиса. Через Go proxy, а не ADD
# с GitHub Releases: релизы периодически отвечают 504 и ломают сборку.
RUN CGO_ENABLED=0 GOBIN=/app go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.37

# Собираем приложение с оптимизациями
WORKDIR /app/iam
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o iam-service ./cmd/main.go

# Финальный образ на основе Alpine для поддержки миграций
FROM alpine:3.20

# Устанавливаем ca-certificates для HTTPS запросов и создаем пользователя
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл
COPY --from=builder /app/iam/iam-service .

# Копируем миграции
COPY --from=builder /app/iam/migrations ./migrations

# Копируем grpc-health-probe для health checks
COPY --from=builder /app/grpc-health-probe .

# Изменяем владельца файлов
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт для gRPC
EXPOSE 50050

# Запускаем приложение
ENTRYPOINT ["./iam-service"]