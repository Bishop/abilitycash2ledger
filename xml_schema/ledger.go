package xml_schema

import (
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
	Accounts          map[string]string
	Classifiers       map[string]map[string]string
	AccountClassifier string
	Db                *Database
}

func (c *LedgerConverter) Transactions() []ledger.Transaction {
	txs := make([]ledger.Transaction, len(c.Db.Transactions))

	for i, source := range c.Db.Transactions {
		txs[i] = ledger.Transaction{
			Date:        source.Date.Source(),
			Description: source.Comment,
			Items:       make([]ledger.TxItem, 2),
		}

		var statusSource txItem

		switch {
		case source.Transfer != nil:
			statusSource = source.Transfer.txItem

			txs[i].Items = []ledger.TxItem{
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
			statusSource = source.Expense.txItem

			classifier := source.Expense.Categories.Map()

			txs[i].Items = []ledger.TxItem{
				{
					Account:  c.account(source.Expense.ExpenseAccount.Name),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   source.Expense.ExpenseAmount,
				},
				{
					Account:  c.accountFromCategories(classifier),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   -source.Expense.ExpenseAmount,
				},
			}
		case source.Income != nil:
			statusSource = source.Income.txItem

			classifier := source.Expense.Categories.Map()

			txs[i].Items = []ledger.TxItem{
				{
					Account:  c.account(source.Income.IncomeAccount.Name),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   source.Income.IncomeAmount,
				},
				{
					Account:  c.accountFromCategories(classifier),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   -source.Income.IncomeAmount,
				},
			}
		case source.Balance != nil:
			statusSource = source.Balance.txItem
		}

		txs[i].Executed = statusSource.IsExecuted()
		txs[i].Locked = statusSource.IsLocked()
	}

	return txs
}

func (c *LedgerConverter) account(a string) string {
	return c.Accounts[a]
}

func (c *LedgerConverter) accountFromCategories(classifier map[string]string) string {
	return c.Classifiers[c.AccountClassifier][classifier[c.AccountClassifier]]
}
