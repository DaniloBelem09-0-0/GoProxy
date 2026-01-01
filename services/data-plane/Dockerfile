# Stage de Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copia os arquivos de dependência (Agora com o caminho completo a partir da raiz)
COPY services/data-plane/go.mod services/data-plane/go.sum ./

# 2. Baixa as dependências
RUN go mod download

# 3. Copia todo o código fonte da pasta data-plane
COPY services/data-plane/ .

# 4. Compila o binário (Ajuste o caminho do main se necessário)
RUN CGO_ENABLED=0 GOOS=linux go build -o /goproxy-bin ./cmd/main.go

# Stage Final (Produção)
FROM alpine:latest

WORKDIR /root/

# Copia o binário do stage anterior
COPY --from=builder /goproxy-bin .

# (Opcional) Se você precisar copiar o prometheus.yml para dentro da imagem também
COPY services/data-plane/prometheus.yml .

EXPOSE 8080

CMD ["./goproxy-bin"]