FROM golang:1.25 AS builder 

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download && go mod verify

COPY main.go ./
RUN go build .


FROM gcr.io/distroless/base
COPY --from=builder /app/* ./

EXPOSE 8080
ENTRYPOINT ["/gochat-contacts_server"] 

CMD ["./gochat-contacts_server"]