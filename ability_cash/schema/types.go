package schema

import (
	"time"

	"github.com/Bishop/abilitycash2ledger/ledger"
)

type Database interface {
	GetAccounts() *[]Account
	GetTransactions() *[]ledger.Transaction
	GetClassifiers() *ClassifiersList
	GetRates() *[]Rate
}

type Account struct {
	Name        string
	Currency    string
	InitBalance float64
}

type Rate struct {
	Date      time.Time
	Currency1 string
	Currency2 string
	Amount1   float64
	Amount2   float64
}

type AccountsMap map[string]string
type ClassifiersMap map[string]AccountsMap
type ClassifiersList map[string][]string
