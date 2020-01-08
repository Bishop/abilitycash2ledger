package ledger

import "time"

type Transaction struct {
	Date          time.Time
	Payee         string
	Note          string
	Cleared       bool
	Pending       bool
	Notes         []string
	Items         []TxItem
	Metadata      map[string]string
	TypedMetadata map[string]interface{}
	Tags          []string
}

type TxItem struct {
	Account  string
	Currency string
	Amount   float64
	Note     string
	Cleared  bool
	Pending  bool
	Payee    string
	Virtual  bool
	Balanced bool
}

type Source interface {
	Transactions() <-chan Transaction
}
