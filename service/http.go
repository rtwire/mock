package service

import (
	"fmt"
	"net/http"
)

func headerFound(w http.ResponseWriter, r *http.Request,
	key, value string) bool {
	if r.Header.Get(key) != value {
		errStr := fmt.Sprintf("header %v: %v not found", key, value)
		http.Error(w, errStr, http.StatusBadRequest)
		return false
	}
	return true
}

func acceptHeaderFound(w http.ResponseWriter, r *http.Request) bool {
	return headerFound(w, r, "Accept", "application/json")
}

func contentTypeHeaderFound(w http.ResponseWriter, r *http.Request) bool {
	return headerFound(w, r, "Content-Type", "application/json")
}
