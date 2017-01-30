package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *service) postAddressHandler(w http.ResponseWriter, r *http.Request) {

	if !contentTypeHeaderFound(w, r) {
		return
	}

	pl := struct {
		Value int64 `json:"value"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&pl); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addr := mux.Vars(r)["address"]

	txID, exists := s.CreditAddress(addr, pl.Value)
	if !exists {
		http.Error(w, "address not found", http.StatusBadRequest)
		return
	}

	if err := s.sendHookCreditEvent(txID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *service) sendHookCreditEvent(txID int64) error {

	tx, exists := s.Transaction(txID)
	if !exists {
		return errors.New("transaction does not exist")
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(jsonMessage{
		Type: "transactions",
		Payload: []transactionPayload{{
			ID:          txID,
			Type:        tx.ty,
			ToAccountID: tx.toAccountID,
			Value:       tx.value,
			Created:     tx.created,
			TxHashes:    []string{},
		}},
	}); err != nil {
		return err
	}

	for _, url := range s.Hooks() {
		go func(hook string) {
			res, err := http.Post(url, "application/json", &body)
			if err != nil {
				log.Printf("Error sending to hook %s: %v.", hook, err)
				return
			}
			if res.StatusCode != http.StatusOK {
				log.Printf("Error response from hook %s: %d.", hook,
					res.StatusCode)
			}
		}(url)
	}
	return nil
}
