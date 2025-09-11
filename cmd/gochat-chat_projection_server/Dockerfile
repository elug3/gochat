FROM golang:1.25.1-alpine   

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY *.go ./

RUN go build
CMD ["./app/gochat-chat_projection_server"]