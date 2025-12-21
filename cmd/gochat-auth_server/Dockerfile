FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-auth_server ./cmd/gochat-auth_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-auth_server /gochat-auth_server

EXPOSE 8080
ENTRYPOINT ["/gochat-auth_server"]
