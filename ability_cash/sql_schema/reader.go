package sql_schema

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
)

const (
	AccountsSql     = "SELECT Id, Name, StartingBalance, Currency FROM Accounts WHERE NOT Deleted"
	CurrenciesSql   = "SELECT Id, Code, Precision FROM Currencies WHERE NOT Deleted"
	CategoriesSql   = "SELECT Id, Name, Parent FROM Categories WHERE NOT Deleted ORDER BY Parent"
	RatesSql        = "SELECT RateDate, Currency1, Currency2, Value1, Value2 FROM CurrencyRates WHERE NOT Deleted ORDER BY RateDate"
	TxCategoriesSql = "SELECT Category, \"Transaction\" FROM TransactionCategories WHERE NOT Deleted"
	TxsSql          = "SELECT Id, BudgetDate, Locked, IncomeAccount, IncomeAmount, ExpenseAccount, ExpenseAmount, Comment FROM Transactions WHERE NOT Deleted AND Executed ORDER BY BudgetDate"
)

func ReadDatabase(fileName string) (schema.Database, error) {
	db := NewDatabase()

	base, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}

	err = query(CurrenciesSql, base, db.readCurrencies)
	if err != nil {
		return nil, err
	}

	err = query(RatesSql, base, db.readRates)
	if err != nil {
		return nil, err
	}

	err = query(AccountsSql, base, db.readAccounts)
	if err != nil {
		return nil, err
	}

	err = query(CategoriesSql, base, db.readCategories)
	if err != nil {
		return nil, err
	}

	err = query(TxCategoriesSql, base, db.readTxCategories)
	if err != nil {
		return nil, err
	}

	err = query(TxsSql, base, db.readTxs)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func query(query string, base *sql.DB, callback func(uid int, fetch FetchFunc) error) error {
	rows, err := base.Query(query)
	if err != nil {
		return err
	}

	var uid int

	for rows.Next() {
		err = callback(uid, rows.Scan)
		if err != nil {
			return err
		}
	}

	return rows.Close()
}
