package service_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rtwire/mock/service"
)

func TestFees(t *testing.T) {
	s := service.New()

	const url = "/v1/mainnet/fees/"

	r := httptest.NewRequest("GET", url, nil)
	r.Header.Set("Accept", "application/json")
	r.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatal("expected StatusOK got", w.Code, w.Body.String())
	}
}
