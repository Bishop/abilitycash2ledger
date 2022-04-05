package sql_schema

import (
	"database/sql"
	"github.com/Bishop/abilitycash2ledger/ledger"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
)

const AccountsSql = "SELECT Id, Name, StartingBalance, Currency FROM Accounts WHERE NOT Deleted"
const CurrenciesSql = "SELECT Id, Code, Precision FROM Currencies WHERE NOT Deleted"
const CategoriesSql = "SELECT Id, Name, Parent FROM Categories WHERE NOT Deleted ORDER BY Parent"
const RatesSql = "SELECT RateDate, Currency1, Currency2, Value1, Value2 FROM CurrencyRates WHERE NOT Deleted ORDER BY RateDate"
const TxCategoriesSql = "SELECT Category, \"Transaction\" FROM TransactionCategories WHERE NOT Deleted"
const TxsSql = "SELECT Id, BudgetDate, Locked, IncomeAccount, IncomeAmount, ExpenseAccount, ExpenseAmount, Comment FROM Transactions WHERE NOT Deleted AND Executed ORDER BY BudgetDate"

func ReadDatabase(fileName string) (schema.Database, error) {
	db := NewDatabase()

	base, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}

	err = readCurrencies(db, base)
	if err != nil {
		return nil, err
	}

	err = readRates(db, base)
	if err != nil {
		return nil, err
	}

	err = readAccounts(db, base)
	if err != nil {
		return nil, err
	}

	err = readCategories(db, base)
	if err != nil {
		return nil, err
	}

	err = readTxCategories(db, base)
	if err != nil {
		return nil, err
	}

	err = readTxs(db, base)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readCurrencies(db *Database, base *sql.DB) error {
	return query(CurrenciesSql, db, base, func(uid int, rows *sql.Rows) error {
		currency := Currency{}

		err := rows.Scan(&uid, &currency.Code, &currency.Precision)
		if err != nil {
			return err
		}

		db.currenciesIndexI[uid] = &currency
		db.currenciesIndexS[currency.Code] = &currency

		return nil
	})
}

func readRates(db *Database, base *sql.DB) error {
	return query(RatesSql, db, base, func(uid int, rows *sql.Rows) error {
		var currency1, currency2 int
		var value1, value2 float64
		var date int64

		err := rows.Scan(&date, &currency1, &currency2, &value1, &value2)
		if err != nil {
			return err
		}

		db.Rates = append(db.Rates, schema.Rate{
			Date:      time.Unix(date, 0),
			Currency1: db.currenciesIndexI[currency1].Code,
			Currency2: db.currenciesIndexI[currency2].Code,
			Amount1:   db.currenciesIndexI[currency1].ConvertAmount(value1),
			Amount2:   db.currenciesIndexI[currency2].ConvertAmount(value2),
		})

		return nil
	})
}

func readAccounts(db *Database, base *sql.DB) error {
	return query(AccountsSql, db, base, func(uid int, rows *sql.Rows) error {
		var currencyId int
		account := schema.Account{}

		err := rows.Scan(&uid, &account.Name, &account.InitBalance, &currencyId)
		if err != nil {
			return err
		}

		account.Currency = db.currenciesIndexI[currencyId].Code
		account.InitBalance = db.currenciesIndexI[currencyId].ConvertAmount(account.InitBalance)

		db.Accounts = append(db.Accounts, account)
		db.accountIndex[uid] = &account

		return nil
	})
}

func readCategories(db *Database, base *sql.DB) error {
	return query(CategoriesSql, db, base, func(uid int, rows *sql.Rows) error {
		var parentId sql.NullInt32
		var name string

		err := rows.Scan(&uid, &name, &parentId)
		if err != nil {
			return err
		}

		if parentId.Valid {
			name = db.categoriesIndex[int(parentId.Int32)] + "\\" + name
		}

		db.categoriesIndex[uid] = name

		return nil
	})
}

func readTxCategories(db *Database, base *sql.DB) error {
	return query(TxCategoriesSql, db, base, func(uid int, rows *sql.Rows) error {
		var category, tx int

		err := rows.Scan(&category, &tx)
		if err != nil {
			return err
		}

		_, ok := db.txCategoriesIndex[tx]
		if !ok {
			db.txCategoriesIndex[tx] = make([]int, 0)
		}
		db.txCategoriesIndex[tx] = append(db.txCategoriesIndex[tx], category)

		return nil
	})
}

func readTxs(db *Database, base *sql.DB) error {
	return query(TxsSql, db, base, func(uid int, rows *sql.Rows) error {
		var iaccout, eaccount sql.NullInt32
		var iamount, eamount sql.NullFloat64
		var date int64
		var locked bool
		var comment string

		err := rows.Scan(&uid, &date, &locked, &iaccout, &iamount, &eaccount, &eamount, &comment)
		if err != nil {
			return err
		}

		tx := ledger.Transaction{
			Date:     time.Unix(date, 0),
			Cleared:  locked,
			Note:     comment,
			Metadata: make(map[string]string),
			Items:    []ledger.TxItem{},
		}

		if iaccout.Valid {
			account := db.accountIndex[int(iaccout.Int32)]

			item := ledger.TxItem{
				Account:  account.Name,
				Currency: account.Currency,
				Amount:   db.currenciesIndexS[account.Currency].ConvertAmount(iamount.Float64),
			}

			tx.Items = append(tx.Items, item)
		}

		if eaccount.Valid {
			account := db.accountIndex[int(eaccount.Int32)]

			item := ledger.TxItem{
				Account:  account.Name,
				Currency: account.Currency,
				Amount:   db.currenciesIndexS[account.Currency].ConvertAmount(eamount.Float64),
			}

			tx.Items = append(tx.Items, item)
		}

		if iaccout.Valid && eaccount.Valid {
			if tx.Items[0].Currency == tx.Items[1].Currency {
				tx.Payee = "Transfer"
			} else {
				tx.Payee = "Exchange"
			}
		}

		if categories, ok := db.txCategoriesIndex[uid]; ok {
			for _, category := range categories {
				name := db.categoriesIndex[category]

				nameParts := strings.SplitN(name, "\\", 2)

				classifier := ""
				switch nameParts[0] {
				case "Income", "Expenses":
					classifier = schema.ExpensesClassifier
				case "Payee":
					classifier = schema.PayeeClassifier
				case "Agents":
					classifier = "Agent"
				}

				tx.Metadata[classifier] = name
			}
		}

		db.Transactions = append(db.Transactions, tx)

		return nil
	})
}

func query(query string, db *Database, base *sql.DB, callback func(uid int, rows *sql.Rows) error) error {
	rows, err := base.Query(query)
	if err != nil {
		return err
	}

	var uid int

	for rows.Next() {
		err = callback(uid, rows)
		if err != nil {
			return err
		}
	}

	return rows.Close()
}
