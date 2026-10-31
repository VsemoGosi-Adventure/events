FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /app/events-service ./cmd/app

FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -u 10001 -g '' appuser

COPY --from=builder /app/events-service /app/events-service
COPY --from=builder /app/migrations /app/migrations

RUN mkdir -p /app/internal/config && touch /app/internal/config/.env
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/events-service"]
