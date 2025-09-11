FROM gcr.io/distroless/base
# TODO: use go build

COPY gochat-message_server ./gochat-message_server

EXPOSE 8080
ENTRYPOINT ["./gochat-message_server"] 

CMD ["./gochat-message_server"]