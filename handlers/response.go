package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(&data); err != nil {
		log.Printf("encoding data to json: %v", err)
	}
}

func WriteError(w http.ResponseWriter, message string, status int) {
	WriteJSON(w, status, struct {
		Error string `json:"error"`
	}{
		Error: message,
	})
}
