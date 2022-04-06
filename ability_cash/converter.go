package ability_cash

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
	GenerateEquity bool
	Db             schema.Database
	Categories     map[string]string
	accounts       map[string]string
}

type Tags struct {
	Payee     string
	ItemPayee string
	Account   string
	Tags      map[string]string
}

func (c *LedgerConverter) Transactions() <-chan ledger.Transaction {
	c.accounts = make(map[string]string)

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
		tags := c.createTags(tx.Tags)
		tx.Tags = nil
		tx.Metadata = tags.Tags

		if tx.Payee == "" {
			tx.Payee = tags.Payee
		}

		if tx.Payee == "" && len(tx.Items) == 2 {
			if tx.Items[0].Currency == tx.Items[1].Currency {
				tx.Payee = "Transfer"

				if math.Abs(tx.Items[0].Amount) == math.Abs(tx.Items[1].Amount) {
					index := 0
					if tx.Items[1].Amount < 0 {
						index = 1
					}
					tx.Items[index].Amount = 0
					tx.Items[index].Currency = ""
				}
			} else {
				tx.Payee = "Exchange"
			}
		}

		if tags.Account != "" {
			tx.Items = append(tx.Items, ledger.TxItem{Account: tags.Account})

			if tx.Items[0].Amount < 0 {
				tx.Items[1].Amount, tx.Items[0].Amount = -tx.Items[0].Amount, 0
				tx.Items[1].Currency, tx.Items[0].Currency = tx.Items[0].Currency, ""
			}
		}

		if tx.Payee == "" {
			tx.Payee = tags.ItemPayee
		} else {
			tx.Items[1].Payee = tags.ItemPayee
		}

		tx.Items[0].Account = c.account(tx.Items[0].Account)
		tx.Items[1].Account = c.account(tx.Items[1].Account)

		txs <- tx
	}
}

func (c *LedgerConverter) Accounts() []string {
	list := make([]string, 0, len(c.accounts))

	for _, account := range c.accounts {
		list = append(list, account)
	}

	sort.Strings(list)

	return list
}

func (c *LedgerConverter) account(s string) string {
	a, ok := c.accounts[s]

	if !ok {
		a = strings.Replace(s, "Assets\\", "", 1)
		a = strings.Replace(a, "\\", ":", -1)
		c.accounts[s] = a
	}

	return a
}

func (c *LedgerConverter) createTags(tags []string) *Tags {
	t := new(Tags)
	t.Tags = make(map[string]string)

	for _, tag := range tags {
		parts := strings.SplitN(tag, "\\", 2)
		switch c.Categories[parts[0]] {
		case "payee":
			t.Payee = c.lastPart(tag)
		case "account":
			t.Account = tag

			if strings.Count(t.Account, "\\") == 3 {
				t.ItemPayee = c.lastPart(t.Account)
				t.Account = t.Account[0 : len(t.Account)-len(t.ItemPayee)-1]
			}
		default:
			t.Tags[parts[0]] = c.lastPart(tag)
		}
	}

	return t
}

func (c *LedgerConverter) lastPart(account string) string {
	parts := strings.Split(account, "\\")

	return parts[len(parts)-1]
}
