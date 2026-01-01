FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/data-plane/go.mod services/data-plane/go.sum ./

RUN go mod download

COPY services/data-plane/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /goproxy-bin ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /goproxy-bin .

COPY services/data-plane/prometheus.yml .

EXPOSE 8080

CMD ["./goproxy-bin"]
