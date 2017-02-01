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
)

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

type service struct {
	mu sync.RWMutex

	params *chaincfg.Params

	accounts          map[int64]account
	orderedAccountIDs []int64

	addresses map[string]int64

	transactions          map[int64]transaction
	unusedTxIDs           map[int64]struct{}
	orderedTransactionIDs []int64

	hooks map[string]struct{}

	ids map[int64]struct{}

	user string
	pass string

	handler http.Handler
}

func (s *service) nextID() int64 {
	for {
		// Add 1 so rand can never be 0.
		id := rand.Int63() + 1
		if _, exists := s.ids[id]; !exists {
			s.ids[id] = struct{}{}
			return id
		}
	}
}

type option func(*service)

// TestNet3 is an option that can be passed to New() in order to make the RTWire
// mock service simulate being on the testnet3 network. If this option is not
// passed the the mock service defaults to being on mainnet.
func TestNet3() func(*service) {
	return func(s *service) {
		s.params = &chaincfg.TestNet3Params
	}
}

// User is an option that can be passsed to New() to change the default user
// name from 'user'.
func User(user string) func(*service) {
	return func(s *service) {
		s.user = user
	}
}

// Pass is an option that can be passed to New() to change the default password
// from 'pass'.
func Pass(pass string) func(*service) {
	return func(s *service) {
		s.pass = pass
	}
}

// New returns a high fidelity mock RTWire service. The *service struct
// implements http.Handler and exposes HTTP endpoints identical to those
// described in https://rtwire.com/docs. This means *service can be combined
// httptest.Server to unit test your RTWire integration code locally. For
// example:
//
//    s := httptest.NewServer(service.New())
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
// https://github.com/rtwire/go/client/ for examples. Importantly one extra
// endpoint at POST /v1/[network]/addresses/[address] is exposed which allows
// the user of this service to send a JSON object "{ value: [value] }" that will
// credit the account assigned to [address] with [value] amount. This is
// equivalent to receiving bitcoins from the network.
//
// The default network is mainnet although it can be changed to testnet3 by
// passing in TestNet3 as an option to New. The default user name is 'user'
// and password is 'pass'. Both those can also be changed by passing in
// User([user name]) and Pass([password]) as options to New respectively.
func New(options ...option) *service {
	s := &service{
		params: &chaincfg.MainNetParams,

		accounts: make(map[int64]account),

		addresses: make(map[string]int64),

		hooks: make(map[string]struct{}),

		transactions: make(map[int64]transaction),
		unusedTxIDs:  make(map[int64]struct{}),

		ids: make(map[int64]struct{}),

		user: "user",
		pass: "pass",
	}

	s.initHandler()

	// All client accounts begin with an account where service fees can be sent
	// and deducted. Although this mock service doesn't implement a fees system
	// now, it may do in the future.
	s.CreateAccount()

	for _, op := range options {
		op(s)
	}
	return s
}

func (s *service) CreateAccount() account {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc := account{
		id:      s.nextID(),
		balance: 0,
	}
	s.accounts[acc.id] = acc
	s.orderedAccountIDs = append(s.orderedAccountIDs, acc.id)
	return acc
}

func (s *service) Accounts(limit, next int) []account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if next+limit > len(s.accounts) {
		limit = len(s.accounts) - next
	}
	accs := make([]account, limit)
	for i, id := range s.orderedAccountIDs[next : next+limit] {
		accs[i] = s.accounts[id]
	}
	return accs
}

func (s *service) AccountTransactions(accID int64,
	limit, next int) []transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if next+limit > len(s.transactions) {
		limit = len(s.transactions) - next
	}
	txns := make([]transaction, 0, limit)
	for _, id := range s.orderedTransactionIDs[next : next+limit] {

		tx := s.transactions[id]
		if tx.fromAccountID == accID || tx.toAccountID == accID {
			txns = append(txns, tx)
		}
	}
	return txns
}

var errAccountNotFound = errors.New("account not found")

func (s *service) CreateAddress(accountID int64) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accounts[accountID]; !exists {
		return "", errAccountNotFound
	}

	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", err
	}
	pubKey := privKey.PubKey()

	addrPubKey, err := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(),
		s.params)
	if err != nil {
		return "", err
	}
	addrPubKeyHash := addrPubKey.AddressPubKeyHash()

	addr := addrPubKeyHash.EncodeAddress()

	s.addresses[addr] = accountID

	return addr, nil
}

func (s *service) Account(id int64) (account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acc, exists := s.accounts[id]
	return acc, exists
}

func (s *service) CreditAddress(address string, value int64) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accID, exists := s.addresses[address]
	if !exists {
		return 0, false
	}

	acc := s.accounts[accID]
	acc.balance += value
	s.accounts[accID] = acc

	txID := s.nextID()
	tx := transaction{
		id:            txID,
		ty:            "credit",
		fromAccountID: accID,
		toAddress:     address,
		value:         value,
		created:       time.Now(),
	}
	s.transactions[txID] = tx
	s.orderedTransactionIDs = append(s.orderedTransactionIDs, txID)

	return txID, true
}

func (s *service) CreateTransactionID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID()
	s.unusedTxIDs[id] = struct{}{}
	return id
}

func (s *service) Transaction(id int64) (transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, exists := s.transactions[id]
	return tx, exists
}

func (s *service) Transfer(txID, fromAccID, toAccID, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.unusedTxIDs[txID]; !exists {
		return errors.New("invalid txID")
	}

	fromAcc, exists := s.accounts[fromAccID]
	if !exists {
		return errors.New("no from account")
	}

	toAcc, exists := s.accounts[toAccID]
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

	s.accounts[fromAccID] = fromAcc
	s.accounts[toAccID] = toAcc

	s.transactions[txID] = transaction{
		id:            txID,
		ty:            "transfer",
		fromAccountID: fromAccID,
		toAccountID:   toAccID,
		value:         value,
		created:       time.Now(),
	}
	s.orderedTransactionIDs = append(s.orderedTransactionIDs, txID)
	delete(s.unusedTxIDs, txID)

	return nil
}

func (s *service) Debit(txID, fromAccID int64,
	toAddr string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.unusedTxIDs[txID]; !exists {
		return errors.New("invalid txID")
	}

	fromAcc, exists := s.accounts[fromAccID]
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
	s.accounts[fromAccID] = fromAcc

	s.transactions[txID] = transaction{
		id:            txID,
		ty:            "debit",
		fromAccountID: fromAccID,
		toAddress:     toAddr,
		value:         value,
		created:       time.Now(),
	}
	s.orderedTransactionIDs = append(s.orderedTransactionIDs, txID)
	delete(s.unusedTxIDs, txID)

	return nil
}

func (s *service) Fees() []fee {
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

func (s *service) CreateHook(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hooks[url]; exists {
		return errHookExists
	}

	if len(s.hooks) > 2 {
		return errMaxHooks
	}
	s.hooks[url] = struct{}{}
	return nil
}

func (s *service) Hooks() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hooks := make([]string, 0, len(s.hooks))
	for hook := range s.hooks {
		hooks = append(hooks, hook)
	}
	return hooks
}

func (s *service) DeleteHook(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hooks[url]; !exists {
		return false
	}
	delete(s.hooks, url)
	return true
}
