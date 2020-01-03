package xml_schema

import (
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type LedgerConverter struct {
	Accounts          map[string]string
	Classifiers       map[string]map[string]string
	AccountClassifier string
	GenerateEquity    bool
	Db                *Database
}

func (c *LedgerConverter) Transactions() []ledger.Transaction {
	shift := 0
	if c.GenerateEquity {
		shift = len(c.Db.Accounts)
	}

	txs := make([]ledger.Transaction, len(c.Db.Transactions)+shift)

	if c.GenerateEquity {
		for i, account := range c.Db.Accounts {
			txs[i] = ledger.Transaction{
				Date:        account.ChangedAt.Source(),
				Description: "Opening Balance",
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
				Executed: true,
				Locked:   true,
			}
		}
	}

	for i, source := range c.Db.Transactions {
		txs[i+shift] = ledger.Transaction{
			Date:        source.Date.Source(),
			Description: source.Comment,
			Items:       make([]ledger.TxItem, 2),
		}
		pTx := &txs[i+shift]

		var statusSource txItem

		switch {
		case source.Transfer != nil:
			statusSource = source.Transfer.txItem

			pTx.Items = []ledger.TxItem{
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

			pTx.Tags = source.Expense.Categories.Map()
			pTx.Items = []ledger.TxItem{
				{
					Account:  c.account(source.Expense.ExpenseAccount.Name),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   source.Expense.ExpenseAmount,
				},
				{
					Account:  c.accountFromCategories(pTx.Tags),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   -source.Expense.ExpenseAmount,
				},
			}
		case source.Income != nil:
			statusSource = source.Income.txItem

			pTx.Tags = source.Income.Categories.Map()
			pTx.Items = []ledger.TxItem{
				{
					Account:  c.account(source.Income.IncomeAccount.Name),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   source.Income.IncomeAmount,
				},
				{
					Account:  c.accountFromCategories(pTx.Tags),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   -source.Income.IncomeAmount,
				},
			}
		case source.Balance != nil:
			statusSource = source.Balance.txItem
		}

		pTx.Executed = statusSource.IsExecuted()
		pTx.Locked = statusSource.IsLocked()
	}

	return txs
}

func (c *LedgerConverter) account(a string) string {
	return c.Accounts[a]
}

func (c *LedgerConverter) accountFromCategories(classifier map[string]string) string {
	return c.Classifiers[c.AccountClassifier][classifier[c.AccountClassifier]]
}
