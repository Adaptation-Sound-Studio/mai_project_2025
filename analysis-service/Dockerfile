FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY .env .env

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
