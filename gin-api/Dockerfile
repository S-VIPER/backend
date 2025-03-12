# Используем официальный образ Go для сборки
FROM golang:1.20-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o gin-api ./cmd/main.go

# Используем минимальный образ Alpine для финального контейнера
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из стадии сборки
COPY --from=builder /app/gin-api .

# Копируем статические файлы (если есть)
COPY ./config ./config

# Открываем порт для приложения
EXPOSE 8080

# Запускаем приложение
CMD ["./sviper-dj"]