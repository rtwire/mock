package service_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rtwire/mock/service"
)

func TestTransations(t *testing.T) {
	s := service.New()

	// Create two accounts and give them balances.
	accIDs := make([]int64, 2)
	for i := range accIDs {
		r := httptest.NewRequest("POST", "/v1/mainnet/accounts/", nil)
		r.SetBasicAuth("user", "pass")
		r.Header.Add("Accept", "application/json")
		w := httptest.NewRecorder()

		s.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
		}

		newAccRes := struct {
			Type    string
			Payload []struct {
				ID      int64
				Balance int64
			}
		}{}

		if err := json.NewDecoder(w.Body).Decode(&newAccRes); err != nil {
			t.Fatal(err)
		}

		accID := newAccRes.Payload[0].ID
		accIDs[i] = accID

		// Get an account address.
		addrURL := fmt.Sprintf("/v1/mainnet/accounts/%d/addresses/", accID)
		r = httptest.NewRequest("POST", addrURL, nil)
		r.SetBasicAuth("user", "pass")
		r.Header.Add("Accept", "application/json")
		w = httptest.NewRecorder()

		s.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
		}

		accAddrRes := struct {
			Type    string
			Payload []struct {
				Address string
			}
		}{}

		if err := json.NewDecoder(w.Body).Decode(&accAddrRes); err != nil {
			t.Fatal(err)
		}
		accAddr := accAddrRes.Payload[0].Address

		// Now credit the address.
		var req bytes.Buffer
		if err := json.NewEncoder(&req).Encode(struct {
			Value int64 `json:"value"`
		}{
			Value: 10,
		}); err != nil {
			t.Fatal(err)
		}

		url := fmt.Sprintf("/v1/mainnet/addresses/%s", accAddr)
		r = httptest.NewRequest("POST", url, &req)
		r.SetBasicAuth("user", "pass")
		r.Header.Add("Content-Type", "application/json")
		w = httptest.NewRecorder()

		s.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
		}
	}

	// Create transaction ID.
	var req bytes.Buffer
	err := json.NewEncoder(&req).Encode(struct {
		N int64 `json:"n"`
	}{
		N: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("POST", "/v1/mainnet/transactions/", &req)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
	}

	res := struct {
		Type    string
		Payload []struct {
			ID int64
		}
	}{}

	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	txID := res.Payload[0].ID

	// Create a transaction.
	req.Reset()
	if err := json.NewEncoder(&req).Encode(struct {
		ID            int64 `json:"id"`
		FromAccountID int64 `json:"fromAccountID"`
		ToAccountID   int64 `json:"toAccountID"`
		Value         int64 `json:"value"`
	}{
		ID:            txID,
		FromAccountID: accIDs[0],
		ToAccountID:   accIDs[1],
		Value:         5,
	}); err != nil {
		t.Fatal(err)
	}

	r = httptest.NewRequest("PUT", "/v1/mainnet/transactions/", &req)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v: %v", http.StatusCreated, w.Code,
			w.Body.String())
	}

	// Get the transaction.
	url := fmt.Sprintf("/v1/mainnet/transactions/%d", txID)
	r = httptest.NewRequest("GET", url, nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	// Get accIDs[0] transactions.
	url = fmt.Sprintf("/v1/mainnet/accounts/%d/transactions/", accIDs[0])
	r = httptest.NewRequest("GET", url, nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	txns := &struct {
		Status  string
		Payload []struct {
			ID   int64
			Type string
		}
	}{}

	if err := json.NewDecoder(w.Body).Decode(txns); err != nil {
		t.Fatal(err)
	}

	if len(txns.Payload) != 2 {
		t.Fatal("incorrect number of transactions")
	}
}
