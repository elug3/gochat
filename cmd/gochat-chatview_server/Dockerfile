FROM golang:1.25.3 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-chatview_server ./cmd/gochat-chatview_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-chatview_server /gochat-chatview_server

EXPOSE 8080
ENTRYPOINT ["/gochat-chatview_server"]
