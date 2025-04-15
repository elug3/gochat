package main

import (
	"log"
	"net/http"
)

var messageService MessageService

func main() {
	messageService = *NewMessageService()

	http.HandleFunc("GET /rooms/{roomid}/messages", handleGetMessages)
	http.HandleFunc("POST /rooms/{roomid}/messages", handlePostMessage)

	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatal(err)
	}
}
