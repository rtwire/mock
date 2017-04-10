package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	getAccountsLimit    = 10
	getAccountsLimitMax = 50
)

type accountPayload struct {
	ID      int64 `json:"id"`
	Balance int64 `json:"balance"`
}

func (c *chain) postAccountsHandler(w http.ResponseWriter, r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	acc := c.CreateAccount()

	sendPayload(w, http.StatusCreated, "accounts", "",
		[]accountPayload{
			{acc.id, acc.balance},
		})
}

func (c *chain) getAccountsHandler(w http.ResponseWriter, r *http.Request) {

	if !acceptHeaderFound(w, r) {
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

	accountsPayload := []accountPayload{}
	for i, acc := range c.Accounts(limit, next) {
		if i < next {
			continue
		}
		accountsPayload = append(accountsPayload, accountPayload{
			ID:      acc.id,
			Balance: acc.balance,
		})
	}

	sendPayload(w, http.StatusOK, "accounts", "", accountsPayload)
}

func (c *chain) getAccountByLabelHandler(w http.ResponseWriter,
	r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	accLabel := mux.Vars(r)["account-label"]

	acc, exists := c.AccountByLabel(accLabel)
	if !exists {
		errStr := fmt.Sprintf(
			"account with label %v not found", accLabel)
		http.Error(w, errStr, http.StatusNotFound)
		return
	}

	sendPayload(w, http.StatusOK, "accounts", "",
		[]accountPayload{
			{acc.id, acc.balance},
		})
}

func (c *chain) getAccountHandler(w http.ResponseWriter, r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	accIDValue := mux.Vars(r)["account-id"]
	accID, err := strconv.ParseInt(accIDValue, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acc, exists := c.Account(accID)
	if !exists {
		errStr := fmt.Sprintf("account with ID %v not found", accID)
		http.Error(w, errStr, http.StatusNotFound)
		return
	}

	sendPayload(w, http.StatusOK, "accounts", "",
		[]accountPayload{
			{acc.id, acc.balance},
		})
}

type addressPayload struct {
	Address string `json:"address"`
}

func (c *chain) postAccountAddresses(w http.ResponseWriter, r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	accIDValue := mux.Vars(r)["account-id"]
	accID, err := strconv.ParseInt(accIDValue, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addr, err := c.CreateAddress(accID)
	if err == errAccountNotFound {
		errStr := fmt.Sprintf("account ID %v not found", accID)
		http.Error(w, errStr, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendPayload(w, http.StatusCreated, "addresses", "",
		[]addressPayload{
			{Address: addr},
		})
}
