FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
<<<<<<< HEAD

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order-service
=======
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order-service ./order-service
>>>>>>> Ohm

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/order-service .

EXPOSE 3003

CMD ["./order-service"]