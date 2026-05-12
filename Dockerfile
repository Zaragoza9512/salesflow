FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o salesflow ./cmd/api

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/salesflow .

EXPOSE 8080

CMD ["./salesflow"]