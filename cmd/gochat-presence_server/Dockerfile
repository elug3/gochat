FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-presence_server ./cmd/gochat-presence_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-presence_server /gochat-presence_server

EXPOSE 8080
ENTRYPOINT ["/gochat-presence_server"]
