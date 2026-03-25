FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
<<<<<<< HEAD

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o menu-service
=======
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o menu-service ./menu-service
>>>>>>> Ohm

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/menu-service .

EXPOSE 3002

CMD ["./menu-service"]