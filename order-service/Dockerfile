FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order-service ./order-service

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/order-service .

EXPOSE 3003

CMD ["./order-service"]