package ledger

import "time"

type Transaction struct {
	Date    time.Time
	Note    string
	Cleared bool
	Pending bool
	Notes   []string
	Items   []TxItem
	Tags    map[string]string
}

type TxItem struct {
	Account  string
	Currency string
	Amount   float64
	Note     string
	Cleared  bool
	Pending  bool
	Payee    string
}

type Source interface {
	Transactions() <-chan Transaction
}
