package ledger

import "time"

type Transaction struct {
	Date        time.Time
	Description string
	Executed    bool
	Locked      bool
	Items       []TxItem
	Tags        map[string]string
}

type TxItem struct {
	Account  string
	Currency string
	Amount   float64
}

type Source interface {
	Transactions() <-chan Transaction
}
