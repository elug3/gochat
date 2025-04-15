package main

import (
	"encoding/json"
	"net/http"
)

// GET rooms/{roomid}/messages
func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomid")
	messages, err := messageService.GetMessages(r.Context(), roomId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(messages)
}

// POST rooms/{roomid}/messages
func handlePostMessage(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomid")

	err := messageService.GetMessages()
	if err != nil {
		http.Error
	}
}
