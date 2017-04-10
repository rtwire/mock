package service

import (
	"github.com/gorilla/mux"
)

func (c *chain) handler(router *mux.Router) {
	mw := middleware{
		c.basicAuthMiddleware(),
	}

	// The handler below is used for testing and does not represent an actual
	accounts := router.PathPrefix("/accounts").Subrouter()
	accounts.Handle("/", mw.Handler(c.postAccountsHandler)).Methods("POST")
	accounts.Handle("/", mw.Handler(c.getAccountsHandler)).Methods("GET")
	accounts.Handle("/labels/{account-label:_?[0-9a-zA-Z]+}/",
		mw.Handler(c.getAccountByLabelHandler)).Methods("GET")
	accounts.Handle("/{account-id:[0-9]+}",
		mw.Handler(c.getAccountHandler)).Methods("GET")
	accounts.Handle("/{account-id:[0-9]+}/addresses/",
		mw.Handler(c.postAccountAddresses)).Methods("POST")
	accounts.Handle("/{account-id:[0-9]+}/transactions/",
		mw.Handler(c.getAccountTransactions)).Methods("GET")

	transactions := router.PathPrefix("/transactions").Subrouter()
	transactions.Handle("/",
		mw.Handler(c.postTransactionsHandler)).Methods("POST")
	transactions.Handle("/",
		mw.Handler(c.putTransactionsHandler)).Methods("PUT")
	transactions.Handle("/{transaction-id:[0-9]+}",
		mw.Handler(c.getTransactionHandler)).Methods("GET")

	hooks := router.PathPrefix("/hooks").Subrouter()
	hooks.Handle("/", mw.Handler(c.postHookHandler)).Methods("POST")
	hooks.Handle("/", mw.Handler(c.getHooksHandler)).Methods("GET")
	hooks.Handle("/{url}", mw.Handler(c.deleteHookHandler)).Methods("DELETE")

	fees := router.PathPrefix("/fees").Subrouter()
	fees.Handle("/", mw.Handler(c.getFeesHandler)).Methods("GET")

	// production service end point.
	router.Handle("/addresses/{address}",
		mw.Handler(c.postAddressHandler)).Methods("POST")
}
