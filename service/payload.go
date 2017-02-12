package service

import (
	"encoding/json"
	"net/http"
)

type jsonMessage struct {
	Type    string      `json:"type"`
	Next    string      `json:"next,omitempty"`
	Payload interface{} `json:"payload"`
}

func sendError(w http.ResponseWriter, code int, message string) {
	resp := jsonMessage{
		Type: "errors",
		Payload: []struct {
			Message string `json:"message"`
		}{
			{Message: message},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendPayload(w http.ResponseWriter, code int,
	ty, next string, payload interface{}) {

	resp := jsonMessage{
		Type:    ty,
		Next:    next,
		Payload: payload,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
