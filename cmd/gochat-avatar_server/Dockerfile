FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-avatar_server ./cmd/gochat-avatar_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-avatar_server /gochat-avatar_server

ENTRYPOINT ["/gochat-avatar_server"]
