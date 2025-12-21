FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-chatcmd_server ./cmd/gochat-chatcmd_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-chatcmd_server /gochat-chatcmd_server

EXPOSE 8080
ENTRYPOINT ["/gochat-chatcmd_server"]
