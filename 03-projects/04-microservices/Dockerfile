# Etapa de compilacion: construye los tres binarios
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar go.mod primero para cachear dependencias
COPY go.mod ./
RUN go mod download

# Copiar codigo fuente
COPY . .

# Compilar los tres servicios como binarios estaticos
RUN CGO_ENABLED=0 GOOS=linux go build -o /userservice ./cmd/userservice
RUN CGO_ENABLED=0 GOOS=linux go build -o /orderservice ./cmd/orderservice
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./cmd/gateway

# Etapa final: imagen minima con los binarios
FROM alpine:3.19

# Certificados para conexiones HTTPS (buena practica)
RUN apk --no-cache add ca-certificates

# Copiar los tres binarios desde la etapa de compilacion
COPY --from=builder /userservice /userservice
COPY --from=builder /orderservice /orderservice
COPY --from=builder /gateway /gateway

# El comando se especifica en docker-compose.yml
ENTRYPOINT ["/gateway"]
