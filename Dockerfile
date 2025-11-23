FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/reviewer-service ./cmd/reviewer-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/reviewer-service .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./reviewer-service"]

