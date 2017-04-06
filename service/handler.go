package service

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *service) initHandler() {
	mw := middleware{
		s.basicAuthMiddleware(),
	}

	router := mux.NewRouter()

	ver := router.PathPrefix("/v1").Subrouter()

	chain := ver.PathPrefix("/" + s.params.Name).Subrouter()

	accounts := chain.PathPrefix("/accounts").Subrouter()
	accounts.Handle("/", mw.Handler(s.postAccountsHandler)).Methods("POST")
	accounts.Handle("/", mw.Handler(s.getAccountsHandler)).Methods("GET")
	accounts.Handle("/labels/{account-label:_?[0-9a-zA-Z]+}/",
		mw.Handler(s.getAccountByLabelHandler)).Methods("GET")
	accounts.Handle("/{account-id:[0-9]+}",
		mw.Handler(s.getAccountHandler)).Methods("GET")
	accounts.Handle("/{account-id:[0-9]+}/addresses/",
		mw.Handler(s.postAccountAddresses)).Methods("POST")
	accounts.Handle("/{account-id:[0-9]+}/transactions/",
		mw.Handler(s.getAccountTransactions)).Methods("GET")

	transactions := chain.PathPrefix("/transactions").Subrouter()
	transactions.Handle("/",
		mw.Handler(s.postTransactionsHandler)).Methods("POST")
	transactions.Handle("/",
		mw.Handler(s.putTransactionsHandler)).Methods("PUT")
	transactions.Handle("/{transaction-id:[0-9]+}",
		mw.Handler(s.getTransactionHandler)).Methods("GET")

	hooks := chain.PathPrefix("/hooks").Subrouter()
	hooks.Handle("/", mw.Handler(s.postHookHandler)).Methods("POST")
	hooks.Handle("/", mw.Handler(s.getHooksHandler)).Methods("GET")
	hooks.Handle("/{url}", mw.Handler(s.deleteHookHandler)).Methods("DELETE")

	fees := chain.PathPrefix("/fees").Subrouter()
	fees.Handle("/", mw.Handler(s.getFeesHandler)).Methods("GET")

	// The handler below is used for testing and does not represent an actual
	// production service end point.
	chain.Handle("/addresses/{address}",
		mw.Handler(s.postAddressHandler)).Methods("POST")

	s.handler = router
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
