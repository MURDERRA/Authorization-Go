FROM golang:1.23-alpine AS builder

WORKDIR /build

RUN apk update --no-cache && apk add --no-cache tzdata

# Копирование и загрузка зависимостей
COPY go.mod go.sum* ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Компиляция приложения
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/auth-service ./main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Установка необходимых пакетов
RUN apk update --no-cache && apk add --no-cache ca-certificates

# Копирование бинарного файла из builder
COPY --from=builder /build/auth-service /app/auth-service
COPY --from=builder /build/config.json /app/config.json

RUN chmod +x /app/auth-service

# Экспорт порта
EXPOSE 8101

# Запуск приложения
CMD ["/bin/sh", "-c", "/app/auth-service"]
