package xml_schema

import (
	"github.com/Bishop/abilitycash2ledger/ledger"
)

func (d *Database) LedgerTransactions(categoryClassifier string) []ledger.Transaction {
	txs := make([]ledger.Transaction, len(d.Transactions))

	for i, source := range d.Transactions {
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
					Account:  source.Transfer.ExpenseAccount.Name,
					Currency: source.Transfer.ExpenseAccount.Currency,
					Amount:   source.Transfer.ExpenseAmount,
				},
				{
					Account:  source.Transfer.IncomeAccount.Name,
					Currency: source.Transfer.IncomeAccount.Currency,
					Amount:   source.Transfer.IncomeAmount,
				},
			}
		case source.Expense != nil:
			statusSource = source.Expense.txItem

			classifier := source.Expense.Categories.Map()

			txs[i].Items = []ledger.TxItem{
				{
					Account:  source.Expense.ExpenseAccount.Name,
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   source.Expense.ExpenseAmount,
				},
				{
					Account:  classifier[categoryClassifier],
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   -source.Expense.ExpenseAmount,
				},
			}
		case source.Income != nil:
			statusSource = source.Income.txItem

			classifier := source.Expense.Categories.Map()

			txs[i].Items = []ledger.TxItem{
				{
					Account:  source.Income.IncomeAccount.Name,
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   source.Income.IncomeAmount,
				},
				{
					Account:  classifier[categoryClassifier],
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
