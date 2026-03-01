FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./user-service

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/user-service .

EXPOSE 3001

CMD ["./user-service"]