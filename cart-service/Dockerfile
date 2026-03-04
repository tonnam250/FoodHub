FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o menu-service ./menu-service

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/cart-service .

EXPOSE 3005

CMD ["./menu-service"]