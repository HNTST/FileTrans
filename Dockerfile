# Stage 1: Сборка бинарного файла
FROM golang:1.23.5 AS builder
WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и собираем бинарный файл
COPY . .
RUN CGO_ENABLED=0 go build -o /myapp

# Stage 2: Создание минимального образа
FROM gcr.io/distroless/static-debian12
WORKDIR /app

# Копируем бинарный файл из Stage 1
COPY --from=builder /myapp /myapp

# Указываем команду запуска
CMD ["/myapp"]