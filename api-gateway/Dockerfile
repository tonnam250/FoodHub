FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-gateway

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]