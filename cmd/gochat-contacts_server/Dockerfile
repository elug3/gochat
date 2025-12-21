FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/gochat-contacts_server ./cmd/gochat-contacts_server

FROM gcr.io/distroless/base

COPY --from=builder /bin/gochat-contacts_server /gochat-contacts_server

EXPOSE 8080
ENTRYPOINT ["/gochat-contacts_server"]
