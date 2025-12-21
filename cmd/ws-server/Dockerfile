FROM golang:1.25.4 AS builder

WORKDIR /app

COPY . .
COPY go.work go.work.sum ./
ENV GOWORK=/app/go.work

RUN go build -o /bin/ws-server ./cmd/ws-server

FROM gcr.io/distroless/base

COPY --from=builder /bin/ws-server /ws-server

EXPOSE 12345
ENTRYPOINT ["/ws-server"]
