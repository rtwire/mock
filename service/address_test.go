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

func TestCreditFeeAddress(t *testing.T) {
	s := service.New()

	r := httptest.NewRequest("GET", "/v1/mainnet/accounts/labels/_fee/", nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	type accPayload struct {
		Payload []struct {
			ID      int64
			Balance int64
		}
	}

	feeAccPayload := &accPayload{}

	if err := json.NewDecoder(w.Body).Decode(feeAccPayload); err != nil {
		t.Fatal(err)
	}

	feeAccID := feeAccPayload.Payload[0].ID

	// Get address for fee account.
	url := fmt.Sprintf("/v1/mainnet/accounts/%d/addresses/", feeAccID)
	r = httptest.NewRequest("POST", url, nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %v got %v", http.StatusCreated, w.Code)
	}

	type addrPayload struct {
		Payload []struct {
			Address string
		}
	}

	feeAddrPayload := &addrPayload{}

	if err := json.NewDecoder(w.Body).Decode(feeAddrPayload); err != nil {
		t.Fatal(err)
	}

	feeAddr := feeAddrPayload.Payload[0].Address
	if feeAddr == "" {
		t.Fatal("no fee address")
	}

	// Credit the fee account using its address.
	const creditValue = int64(1234)
	url = fmt.Sprintf("/v1/mainnet/addresses/%s", feeAddr)

	body, err := json.Marshal(struct {
		Value int64
	}{
		Value: creditValue,
	})
	if err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	r.SetBasicAuth("user", "pass")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	// Ensure that the account is credited.
	url = fmt.Sprintf("/v1/mainnet/accounts/%d", feeAccID)
	r = httptest.NewRequest("GET", url, nil)
	r.SetBasicAuth("user", "pass")
	r.Header.Add("Accept", "application/json")
	w = httptest.NewRecorder()

	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %v got %v", http.StatusOK, w.Code)
	}

	feeAccPayload = &accPayload{}

	if err := json.NewDecoder(w.Body).Decode(feeAccPayload); err != nil {
		t.Fatal(err)
	}

	if creditValue != feeAccPayload.Payload[0].Balance {
		t.Fatal("incrrect balance", feeAccPayload.Payload[0].Balance)
	}
}
