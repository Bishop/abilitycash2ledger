package ability_cash

import (
	"github.com/Bishop/abilitycash2ledger/ability_cash/xml_schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
	Accounts          map[string]string
	Classifiers       map[string]map[string]string
	AccountClassifier string
	GenerateEquity    bool
	Db                *xml_schema.Database
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
		for _, account := range c.Db.Accounts {
			if account.InitBalance == 0 {
				continue
			}

			txs <- ledger.Transaction{
				Date: account.ChangedAt.Source(),
				Note: "Opening Balance",
				Items: []ledger.TxItem{
					{
						Account:  c.account(account.Name),
						Currency: account.Currency,
						Amount:   account.InitBalance,
					},
					{
						Account:  "equity:opening balances",
						Currency: account.Currency,
						Amount:   -account.InitBalance,
					},
				},
				Cleared: true,
			}
		}
	}

	for _, source := range c.Db.Transactions {
		if !source.IsExecuted() {
			continue
		}

		tx := ledger.Transaction{
			Date: source.Date.Source(),
			Note: source.Comment,
		}

		switch {
		case source.Transfer != nil:
			tx.Items = []ledger.TxItem{
				{
					Account:  c.account(source.Transfer.ExpenseAccount.Name),
					Currency: source.Transfer.ExpenseAccount.Currency,
					Amount:   source.Transfer.ExpenseAmount,
				},
				{
					Account:  c.account(source.Transfer.IncomeAccount.Name),
					Currency: source.Transfer.IncomeAccount.Currency,
					Amount:   source.Transfer.IncomeAmount,
				},
			}
		case source.Expense != nil:
			tx.Tags = source.Expense.Categories.Map()
			tx.Items = []ledger.TxItem{
				{
					Account:  c.account(source.Expense.ExpenseAccount.Name),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   source.Expense.ExpenseAmount,
				},
				{
					Account:  c.accountFromCategories(tx.Tags),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   -source.Expense.ExpenseAmount,
				},
			}
		case source.Income != nil:
			tx.Tags = source.Income.Categories.Map()
			tx.Items = []ledger.TxItem{
				{
					Account:  c.account(source.Income.IncomeAccount.Name),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   source.Income.IncomeAmount,
				},
				{
					Account:  c.accountFromCategories(tx.Tags),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   -source.Income.IncomeAmount,
				},
			}
		case source.Balance != nil:
			tx.Items = []ledger.TxItem{
				{
					Account:  c.account(source.Balance.IncomeAccount.Name),
					Currency: source.Balance.IncomeAccount.Currency,
					Amount:   source.Balance.IncomeAmount,
				},
				{
					Account:  "equity:correction",
					Currency: source.Balance.IncomeAccount.Currency,
					Amount:   -source.Balance.IncomeAmount,
				},
			}
		}

		tx.Cleared = source.IsLocked()

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

func (c *LedgerConverter) account(a string) string {
	return c.Accounts[a]
}

func (c *LedgerConverter) accountFromCategories(classifier map[string]string) string {
	return c.Classifiers[c.AccountClassifier][classifier[c.AccountClassifier]]
}
