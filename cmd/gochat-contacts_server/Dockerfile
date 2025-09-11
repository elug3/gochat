# FROM golang:1.25 AS builder 

# WORKDIR /app

# COPY go.mod go.sum* main.go ./
# RUN go mod tidy
# RUN go build .


FROM gcr.io/distroless/base
# TODO: use go build
COPY gochat-contacts_server ./gochat-contacts_server

EXPOSE 8080
ENTRYPOINT ["/gochat-contacts_server"] 

CMD ["./gochat-contacts_server"]