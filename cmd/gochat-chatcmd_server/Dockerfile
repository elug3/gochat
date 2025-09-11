# COPY go.mod go.sum* ./
# RUN go mod download

# COPY *.go ./

# RUN go build

# CMD ["./gochat-chatcmd_server"]


FROM gcr.io/distroless/base
# TODO: use go build



COPY gochat-chatcmd_server ./gochat-chatcmd_server

EXPOSE 8080
ENTRYPOINT ["./gochat-chatcmd_server"] 

CMD ["./gochat-chatcmd_server"]