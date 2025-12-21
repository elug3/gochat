FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-message_server ./cmd/gochat-message_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-message_server /gochat-message_server

EXPOSE 8080
ENTRYPOINT ["/gochat-message_server"]
