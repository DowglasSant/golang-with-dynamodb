# Etapa 1: Build
# Usa a imagem oficial do Go para compilar o binario.
# O "AS builder" da um nome a essa etapa para referenciamos depois.
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copia go.mod e go.sum primeiro para aproveitar cache de camadas do Docker.
# Se as dependencias nao mudaram, o Docker reutiliza essa camada sem baixar tudo de novo.
COPY go.mod go.sum ./
RUN go mod download

# Agora copia o restante do codigo e compila.
# CGO_ENABLED=0 gera um binario estatico (sem dependencia de libc),
# necessario para rodar em imagens minimas como alpine.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Etapa 2: Runtime
# Usa alpine (5MB) em vez da imagem Go (800MB+).
# O container final so tem o binario compilado — nada de Go, git ou ferramentas de build.
FROM alpine:3.20

# Certificados SSL para o SDK da AWS conseguir fazer requests HTTPS.
RUN apk --no-cache add ca-certificates

COPY --from=builder /api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
