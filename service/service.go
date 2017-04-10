package service

import (
	"errors"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/gorilla/mux"
)

// Network specifies the Bitcoin network.
type Network string

const (
	// TestNet3 represents the testnet3 Bitcoin network.
	TestNet3 Network = "testnet3"

	// MainNet represents the mainnet Bitcoin network.
	MainNet Network = "mainnet"
)

var networkParams = map[Network]*chaincfg.Params{
	TestNet3: &chaincfg.TestNet3Params,
	MainNet:  &chaincfg.MainNetParams,
}

type account struct {
	id      int64
	balance int64
}

type transaction struct {
	id int64
	ty string

	fromAccountID int64
	toAccountID   int64
	toAddress     string
	value         int64

	created time.Time
}

type fee struct {
	feePerByte  int64
	blockHeight int64
}

type chain struct {
	mu sync.RWMutex

	network Network

	accounts          map[int64]account
	orderedAccountIDs []int64
	accountLabels     map[string]int64

	addresses map[string]int64

	transactions          map[int64]transaction
	unusedTxIDs           map[int64]struct{}
	orderedTransactionIDs []int64

	hooks map[string]struct{}

	ids map[int64]struct{}

	user string
	pass string
}

func (c *chain) params() *chaincfg.Params {
	return networkParams[c.network]
}

func (c *chain) nextID() int64 {
	for {
		// Add 1 so rand can never be 0.
		id := rand.Int63() + 1
		if _, exists := c.ids[id]; !exists {
			c.ids[id] = struct{}{}
			return id
		}
	}
}

func (c *chain) CreateAccount() account {
	c.mu.Lock()
	defer c.mu.Unlock()

	acc := account{
		id:      c.nextID(),
		balance: 0,
	}
	c.accounts[acc.id] = acc
	c.orderedAccountIDs = append(c.orderedAccountIDs, acc.id)
	return acc
}

func (c *chain) Accounts(limit, next int) []account {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if next+limit > len(c.accounts) {
		limit = len(c.accounts) - next
	}
	accs := make([]account, limit)
	for i, id := range c.orderedAccountIDs[next : next+limit] {
		accs[i] = c.accounts[id]
	}
	return accs
}

func (c *chain) AccountTransactions(accID int64,
	limit, next int) []transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if next+limit > len(c.transactions) {
		limit = len(c.transactions) - next
	}
	txns := make([]transaction, 0, limit)
	for _, id := range c.orderedTransactionIDs[next : next+limit] {

		tx := c.transactions[id]
		if tx.fromAccountID == accID || tx.toAccountID == accID {
			txns = append(txns, tx)
		}
	}
	return txns
}

var errAccountNotFound = errors.New("account not found")

func (c *chain) CreateAddress(accountID int64) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.accounts[accountID]; !exists {
		return "", errAccountNotFound
	}

	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", err
	}
	pubKey := privKey.PubKey()

	addrPubKey, err := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(),
		c.params())
	if err != nil {
		return "", err
	}
	addrPubKeyHash := addrPubKey.AddressPubKeyHash()

	addr := addrPubKeyHash.EncodeAddress()

	c.addresses[addr] = accountID

	return addr, nil
}

func (c *chain) Account(id int64) (account, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	acc, exists := c.accounts[id]
	return acc, exists
}

func (c *chain) AccountByLabel(label string) (account, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	accID := c.accountLabels[label]
	acc, exists := c.accounts[accID]
	return acc, exists
}

func (c *chain) CreditAddress(address string, value int64) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	accID, exists := c.addresses[address]
	if !exists {
		return 0, false
	}

	acc := c.accounts[accID]
	acc.balance += value
	c.accounts[accID] = acc

	txID := c.nextID()
	tx := transaction{
		id:            txID,
		ty:            "credit",
		fromAccountID: accID,
		toAddress:     address,
		value:         value,
		created:       time.Now(),
	}
	c.transactions[txID] = tx
	c.orderedTransactionIDs = append(c.orderedTransactionIDs, txID)

	return txID, true
}

func (c *chain) CreateTransactionID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID()
	c.unusedTxIDs[id] = struct{}{}
	return id
}

func (c *chain) Transaction(id int64) (transaction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tx, exists := c.transactions[id]
	return tx, exists
}

func (c *chain) Transfer(txID, fromAccID, toAccID, value int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.unusedTxIDs[txID]; !exists {
		return errors.New("invalid txID")
	}

	fromAcc, exists := c.accounts[fromAccID]
	if !exists {
		return errors.New("no from account")
	}

	toAcc, exists := c.accounts[toAccID]
	if !exists {
		return errors.New("no to account")
	}

	if value <= 0 {
		return errors.New("invalid balance")
	}

	if fromAcc.balance < value {
		return errors.New("insufficient funds")
	}

	fromAcc.balance = fromAcc.balance - value
	toAcc.balance = toAcc.balance + value

	c.accounts[fromAccID] = fromAcc
	c.accounts[toAccID] = toAcc

	c.transactions[txID] = transaction{
		id:            txID,
		ty:            "transfer",
		fromAccountID: fromAccID,
		toAccountID:   toAccID,
		value:         value,
		created:       time.Now(),
	}
	c.orderedTransactionIDs = append(c.orderedTransactionIDs, txID)
	delete(c.unusedTxIDs, txID)

	return nil
}

func (c *chain) Debit(txID, fromAccID int64,
	toAddr string, value int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.unusedTxIDs[txID]; !exists {
		return errors.New("invalid txID")
	}

	fromAcc, exists := c.accounts[fromAccID]
	if !exists {
		return errors.New("no from account")
	}

	if toAddr == "" {
		return errors.New("no to address")
	}

	if value <= 0 {
		return errors.New("invalid balance")
	}

	if fromAcc.balance < value {
		return errors.New("insufficient funds")
	}

	fromAcc.balance = fromAcc.balance - value
	c.accounts[fromAccID] = fromAcc

	c.transactions[txID] = transaction{
		id:            txID,
		ty:            "debit",
		fromAccountID: fromAccID,
		toAddress:     toAddr,
		value:         value,
		created:       time.Now(),
	}
	c.orderedTransactionIDs = append(c.orderedTransactionIDs, txID)
	delete(c.unusedTxIDs, txID)

	return nil
}

func (c *chain) Fees() []fee {
	return []fee{
		{
			feePerByte:  100,
			blockHeight: 451000,
		},
	}
}

var (
	errHookExists = errors.New("hook exists")
	errMaxHooks   = errors.New("max hooks")
)

func (c *chain) CreateHook(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.hooks[url]; exists {
		return errHookExists
	}

	if len(c.hooks) > 2 {
		return errMaxHooks
	}
	c.hooks[url] = struct{}{}
	return nil
}

func (c *chain) Hooks() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hooks := make([]string, 0, len(c.hooks))
	for hook := range c.hooks {
		hooks = append(hooks, hook)
	}
	return hooks
}

func (c *chain) DeleteHook(url string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.hooks[url]; !exists {
		return false
	}
	delete(c.hooks, url)
	return true
}

type service struct {
	router *mux.Router
}

// New returns a high fidelity mock RTWire service. The *service struct
// implements http.Handler and exposes HTTP endpoints identical to those
// described in https://rtwire.com/docs. This means *service can be combined
// httptest.Server to unit test your RTWire integration code locally. For
// example:
//
//    s := httptest.NewServer(cervice.New())
//    defer s.Close()
//
//    req, err := http.NewRequest("GET", s.URL+"/v1/mainnet/accounts", nil)
//    req.SetBasicAuth("user", "pass")
//    req.Header.Set("Content-Type", "application/json")
//
//    resp, err := http.DefaultClient.Do(req)
//    // Handle error if not nil. resp.Body will contain a JSON object with all
//    // accounts created so far with the mock service.
//
// This will allow the s.URL to expose the same HTTP endpoints as RTWire's. See
// https://github.com/rtwire/go/client/ for examples. Importantly this mock
// service exposes one extra endpoint:
//
//     POST /v1/[network]/addresses/[address]
//
// This endpoint can be passed the folliwng JSON object that will credit the
// account at address [address] with value [value].
//
//    {
//      "value": [value]
//	  }
//
// Note that you must specify a "Content-Type: application/json" header with
// this endpoint. Using this endpoint is the equivalent to receiving bitcoins
// from the network.
//
// The default basic authentication username is user is 'user' and password is
// 'pass'. Both can be changed by using the UserPass option.
func New(options ...option) *service {
	s := &service{
		router: mux.NewRouter().PathPrefix("/v1").Subrouter(),
	}

	for _, net := range []Network{TestNet3, MainNet} {
		c := &chain{
			network: net,

			accounts:      make(map[int64]account),
			accountLabels: make(map[string]int64),

			addresses: make(map[string]int64),

			hooks: make(map[string]struct{}),

			transactions: make(map[int64]transaction),
			unusedTxIDs:  make(map[int64]struct{}),

			ids: make(map[int64]struct{}),

			user: "user",
			pass: "pass",
		}

		for _, op := range options {
			op(c)
		}

		// All client accounts begin with an account where service fees can be
		// sent and deducted.
		feeAcc := c.CreateAccount()
		c.accountLabels["_fee"] = feeAcc.id

		c.handler(s.router.PathPrefix("/" + c.params().Name).Subrouter())

	}
	return s
}

type option func(*chain)

// UserPass is an option that can be passsed to New() to change the default user
// and pass authentication credentials for the specified network.
func UserPass(network Network, user, pass string) option {
	return func(c *chain) {
		if c.network == network {
			c.user = user
			c.pass = pass
		}
	}
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
