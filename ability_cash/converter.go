package ability_cash

import (
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
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
		tx := ledger.Transaction{
			Date:    time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local),
			Payee:   "Opening Balance",
			Cleared: true,
			Items:   make([]ledger.TxItem, 0),
		}

		for _, account := range *c.Db.GetAccounts() {
			if account.InitBalance == 0 {
				continue
			}

			tx.Items = append(tx.Items, ledger.TxItem{
				Account:  account.Name,
				Currency: account.Currency,
				Amount:   account.InitBalance,
			})
		}

		tx.Items = append(tx.Items, ledger.TxItem{Account: ledger.OpeningBalance})
		txs <- tx
	}

	for _, tx := range *c.Db.GetTransactions() {
		if tx.Payee == "" && tx.Metadata[schema.PayeeClassifier] != "" {
			tx.Payee = c.lastPart(tx.Metadata[schema.PayeeClassifier])
		}
		txs <- tx
	}
}

func (c *LedgerConverter) lastPart(account string) string {
	parts := strings.Split(account, "\\")

	return parts[len(parts)-1]
}

func (c *LedgerConverter) AccountsList() <-chan string {
	list := make(chan string)

	go func() {
		for _, account := range *c.Db.GetAccounts() {
			list <- account.Name
		}
		if accounts, ok := (*c.Db.GetClassifiers())[schema.ExpensesClassifier]; ok {
			for _, account := range accounts {
				list <- account
			}
		}

		close(list)
	}()

	return list
}
