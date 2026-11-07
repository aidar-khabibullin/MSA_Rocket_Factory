# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.25-alpine AS builder

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum файлы для кеширования зависимостей
COPY assembly/go.mod assembly/go.sum ./assembly/
COPY shared/go.mod shared/go.sum ./shared/
COPY platform/go.mod platform/go.sum ./platform/

# Загружаем зависимости
WORKDIR /app/assembly
RUN go mod download

# Возвращаемся в корневую директорию и копируем исходный код
WORKDIR /app
COPY assembly/ ./assembly/
COPY shared/ ./shared/
COPY platform/ ./platform/

# Собираем приложение с оптимизациями
WORKDIR /app/assembly
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o assembly-service ./cmd/main.go

# Финальный минимальный образ
FROM scratch

# Копируем SSL сертификаты для HTTPS запросов
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем /etc/passwd, чтобы UID 65534 (nobody) разрешался в имя пользователя
COPY --from=builder /etc/passwd /etc/passwd

# Копируем бинарный файл
COPY --from=builder /app/assembly/assembly-service /assembly-service

# Запускаем от непривилегированного пользователя nobody (UID 65534).
# /etc/passwd нужен обязательно: OpenTelemetry определяет текущего пользователя через user.Current(),
# а в статическом бинарнике (CGO_ENABLED=0) без записи в passwd он падает на старте.
USER 65534:65534

ENTRYPOINT ["/assembly-service"]
