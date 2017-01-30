package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/gorilla/mux"
)

func (s *service) postTransactionsHandler(w http.ResponseWriter,
	r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	if !contentTypeHeaderFound(w, r) {
		return
	}

	n := struct {
		N int `json:"n"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if n.N < 1 || n.N > 10 {
		sendError(w, http.StatusBadRequest, "n must be > 0 and <= 10")
		return
	}

	idsPayload := make([]struct {
		ID int64 `json:"id"`
	}, n.N)

	for i := range idsPayload {
		idsPayload[i].ID = s.CreateTransactionID()
	}

	sendPayload(w, http.StatusCreated, "transactions", "", idsPayload)
}

func (s *service) putTransactionsHandler(w http.ResponseWriter,
	r *http.Request) {

	pl := struct {
		ID            int64
		FromAccountID int64
		ToAccountID   int64
		ToAddress     string
		Value         int64
	}{}

	if err := json.NewDecoder(r.Body).Decode(&pl); err != nil {
		sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if pl.ID <= 0 {
		sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if pl.FromAccountID <= 0 {
		sendError(w, http.StatusBadRequest, "invalid fromAccountID")
		return
	}

	if pl.Value <= 0 {
		sendError(w, http.StatusBadRequest, "invalid value")
		return
	}

	if pl.ToAddress == "" {
		if pl.ToAccountID <= 0 {
			sendError(w, http.StatusBadRequest, "invalid toAccountID")
			return
		}
		err := s.Transfer(pl.ID, pl.FromAccountID, pl.ToAccountID, pl.Value)
		if err != nil {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else {

		toAddr, err := btcutil.DecodeAddress(pl.ToAddress, s.params)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid toAddress")
			return
		}
		if _, ok := toAddr.(*btcutil.AddressPubKeyHash); !ok {
			sendError(w, http.StatusBadRequest,
				"toAddress not public key hash")
			return
		}
		if !toAddr.IsForNet(s.params) {
			sendError(w, http.StatusBadRequest,
				"toAddress for wrong chain")
			return
		}

		if err := s.Debit(pl.ID,
			pl.FromAccountID, pl.ToAddress, pl.Value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

type transactionPayload struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`

	FromAccountID int64 `json:"fromAccountID"`
	ToAccountID   int64 `json:"toAccountID"`

	FromAccountBalance int64 `json:"fromAccountBalance"`
	ToAccountBalance   int64 `json:"toAccountBalance"`

	FromAccountTxID int64 `json:"fromAccountTxID"`
	ToAccountTxID   int64 `json:"toAccountTxID"`

	Value int64 `json:"value"`

	Created time.Time `json:"created"`

	TxHashes []string `json:"txHashes,omitempty"`
	TxIndex  int64    `json:"txIndex,omitempty"`
}

func (s *service) getTransactionHandler(w http.ResponseWriter,
	r *http.Request) {

	txIDValue := mux.Vars(r)["transaction-id"]
	txID, err := strconv.ParseInt(txIDValue, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, exists := s.Transaction(txID)
	if !exists {
		errStr := fmt.Sprintf("transaction with ID %v not found", txID)
		http.Error(w, errStr, http.StatusNotFound)
		return
	}

	sendPayload(w, http.StatusOK, "transactions", "",
		[]transactionPayload{
			{
				ID:   txID,
				Type: tx.ty,

				FromAccountID: tx.fromAccountID,
				ToAccountID:   tx.toAccountID,

				Value:   tx.value,
				Created: tx.created,
			},
		})
}

func (s *service) getAccountTransactions(w http.ResponseWriter,
	r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	accIDValue := mux.Vars(r)["account-id"]
	accID, err := strconv.ParseInt(accIDValue, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if accID < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limitValue := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitValue)
	if limitValue == "" {
		limit = getAccountsLimit
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if limit > getAccountsLimitMax {
		errStr := fmt.Sprintf("limit > %d", getAccountsLimitMax)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	nextValue := r.URL.Query().Get("next")
	next, err := strconv.Atoi(nextValue)
	if nextValue == "" {
		next = 0
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload := []transactionPayload{}
	for i, tx := range s.AccountTransactions(accID, limit, next) {
		if i < next {
			continue
		}
		payload = append(payload, transactionPayload{
			ID:   tx.id,
			Type: tx.ty,

			FromAccountID: tx.fromAccountID,
			ToAccountID:   tx.toAccountID,

			Value:   tx.value,
			Created: tx.created,
		})
	}
	sendPayload(w, http.StatusOK, "transactions", "", payload)
}
