package service_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rtwire/mock/service"
)

func TestHooks(t *testing.T) {
	s := service.New()

	const hookURL = "https://testurl.com/"

	urls := struct {
		URL string `json:"url"`
	}{
		URL: hookURL,
	}
	var req bytes.Buffer
	enc := json.NewEncoder(&req)
	if err := enc.Encode(urls); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/mainnet/hooks/", &req)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
	}

	// Try and add the same URL again.
	req.Reset()
	enc = json.NewEncoder(&req)
	if err := enc.Encode(urls); err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("POST", "/v1/mainnet/hooks/", &req)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Content-Type", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %v got %v", http.StatusBadRequest, w.Code)
	}
}
