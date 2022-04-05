package schema

import (
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ledger"
)

const (
	PayeeClassifier    = "Provider"
	ExpensesClassifier = "Category"
)

type Database interface {
	GetAccounts() *[]Account
	GetTransactions() *[]ledger.Transaction
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

func CategoryClassifier(category string) string {
	parts := strings.SplitN(category, "\\", 2)

	switch parts[0] {
	case "Income", "Expenses":
		return ExpensesClassifier
	case "Payee":
		return PayeeClassifier
	case "Agents":
		return "Agent"
	default:
		return ""
	}
}
