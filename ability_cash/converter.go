package ability_cash

import (
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
	Accounts          schema.AccountsMap
	Classifiers       schema.ClassifiersMap
	AccountClassifier string
	GenerateEquity    bool
	Db                schema.Database
}

func (c *LedgerConverter) Transactions() <-chan ledger.Transaction {
	txs := make(chan ledger.Transaction)

	go func() {
		c.transactions(txs)
		close(txs)
	}()

	return txs
}

func (c *LedgerConverter) transactions(txs chan<- ledger.Transaction) {
	if c.GenerateEquity {
		for _, account := range *c.Db.GetAccounts() {
			if account.InitBalance == 0 {
				continue
			}

			txs <- ledger.Transaction{
				Date:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local),
				Payee: "Opening Balance",
				Items: []ledger.TxItem{
					{
						Account:  account.Name,
						Currency: account.Currency,
						Amount:   account.InitBalance,
					},
					{
						Account:  ledger.OpeningBalance,
						Currency: account.Currency,
						Amount:   -account.InitBalance,
					},
				},
				Cleared: true,
			}
		}
	}
	for _, tx := range *c.Db.GetTransactions() {
		txs <- tx
	}
}

func (c *LedgerConverter) AccountsList() <-chan string {
	list := make(chan string)

	go func() {
		for _, account := range c.Accounts {
			list <- account
		}
		for _, account := range c.Classifiers[c.AccountClassifier] {
			list <- account
		}
		close(list)
	}()

	return list
}
