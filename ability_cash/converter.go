package ability_cash

import (
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
	GenerateEquity bool
	Db             schema.Database
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
				Account:  c.account(account.Name),
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
			delete(tx.Metadata, schema.PayeeClassifier)
		}

		if tx.Payee == "" && len(tx.Items) == 2 {
			if tx.Items[0].Currency == tx.Items[1].Currency {
				tx.Payee = "Transfer"
			} else {
				tx.Payee = "Exchange"
			}
		}

		if tx.Metadata[schema.ExpensesClassifier] != "" {
			account := tx.Metadata[schema.ExpensesClassifier]
			delete(tx.Metadata, schema.ExpensesClassifier)

			if tx.Payee == "" && strings.Count(account, "\\") == 3 {
				tx.Payee = c.lastPart(account)
				account = strings.Replace(account, "\\"+tx.Payee, "", 1)
			}

			tx.Items = append(tx.Items, ledger.TxItem{Account: account})

			if tx.Items[0].Amount < 0 {
				tx.Items[1].Amount = -tx.Items[0].Amount
				tx.Items[0].Amount = 0
				tx.Items[1].Currency = tx.Items[0].Currency
				tx.Items[0].Currency = ""
			}
		}

		tx.Items[0].Account = c.account(tx.Items[0].Account)
		tx.Items[1].Account = c.account(tx.Items[1].Account)

		txs <- tx
	}
}

func (c *LedgerConverter) lastPart(account string) string {
	parts := strings.Split(account, "\\")

	return parts[len(parts)-1]
}

func (c *LedgerConverter) account(s string) string {
	s = strings.Replace(s, "Assets\\", "", 1)
	s = strings.Replace(s, "\\", ":", -1)

	return s
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
