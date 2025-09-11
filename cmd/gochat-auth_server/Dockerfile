FROM gcr.io/distroless/base
# TODO: use go build
COPY gochat-auth_server ./gochat-contacts_server

EXPOSE 8080
ENTRYPOINT ["/gochat-contacts_server"] 

CMD ["./gochat-contacts_server"]