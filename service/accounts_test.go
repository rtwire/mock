package service_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rtwire/mock/service"
)

func TestAccounts(t *testing.T) {
	s := service.New()

	// Check we have a fee account.
	r := httptest.NewRequest("GET", "/v1/mainnet/accounts/labels/_fee/", nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	r = httptest.NewRequest("POST", "/v1/mainnet/accounts/", nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
	}

	r = httptest.NewRequest("GET", "/v1/mainnet/accounts/", nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	res := struct {
		Type    string
		Payload []struct {
			ID      int64
			Balance int64
		}
	}{}

	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	accID := res.Payload[0].ID

	// We should really extract the account ID from earlier and use it here
	// instead of hardcoding it.
	url := fmt.Sprintf("/v1/mainnet/accounts/%d", accID)
	r = httptest.NewRequest("GET", url, nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

}

func TestAccountsNext(t *testing.T) {
	t.SkipNow()
}

func TestAccountsLimit(t *testing.T) {
	t.SkipNow()
}

func TestAccountAddresses(t *testing.T) {
	s := service.New()

	r := httptest.NewRequest("POST", "/v1/mainnet/accounts/", nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
	}

	res := struct {
		Type    string
		Payload []struct {
			ID      int64
			Balance int64
		}
	}{}

	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	accID := res.Payload[0].ID

	url := fmt.Sprintf("/v1/mainnet/accounts/%d/addresses/", accID)
	r = httptest.NewRequest("POST", url, nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
	}
}
