FROM golang:1.24.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/main ./cmd/lyrics-library/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main/ .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["/app/main"]