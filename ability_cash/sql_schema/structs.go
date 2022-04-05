package sql_schema

import (
	"database/sql"
	"math"
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type Database struct {
	Rates             []schema.Rate
	Classifiers       schema.ClassifiersList
	Accounts          []schema.Account
	AccountsMap       schema.AccountsMap
	Transactions      []ledger.Transaction
	accountIndex      map[int]*schema.Account
	currenciesIndexI  map[int]*Currency
	currenciesIndexS  map[string]*Currency
	categoriesIndex   map[int]string
	txCategoriesIndex map[int][]int
}

type Currency struct {
	Code      string
	Precision float64
}

type FetchFunc func(dest ...any) error

func NewDatabase() *Database {
	db := new(Database)

	db.Rates = make([]schema.Rate, 0)
	db.Classifiers = make(schema.ClassifiersList)
	db.Classifiers[schema.ExpensesClassifier] = make([]string, 0)
	db.Classifiers[schema.PayeeClassifier] = make([]string, 0)
	db.Classifiers["Agent"] = make([]string, 0)
	db.Accounts = make([]schema.Account, 0)
	db.AccountsMap = make(schema.AccountsMap)
	db.Transactions = make([]ledger.Transaction, 0)

	db.accountIndex = make(map[int]*schema.Account)
	db.currenciesIndexI = make(map[int]*Currency)
	db.currenciesIndexS = make(map[string]*Currency)
	db.categoriesIndex = make(map[int]string)
	db.txCategoriesIndex = make(map[int][]int)

	return db
}

func (d *Database) GetAccounts() *[]schema.Account {
	return &d.Accounts
}

func (d *Database) GetTransactions() *[]ledger.Transaction {
	return &d.Transactions
}

func (d *Database) GetClassifiers() *schema.ClassifiersList {
	return &d.Classifiers
}

func (d *Database) GetRates() *[]schema.Rate {
	return &d.Rates
}

func (d *Database) readCurrencies(uid int, fetch FetchFunc) error {
	currency := Currency{}

	err := fetch(&uid, &currency.Code, &currency.Precision)
	if err != nil {
		return err
	}

	d.currenciesIndexI[uid] = &currency
	d.currenciesIndexS[currency.Code] = &currency

	return nil
}

func (d *Database) readRates(uid int, fetch FetchFunc) error {
	var currency1, currency2 int
	var value1, value2 float64
	var date int64

	err := fetch(&date, &currency1, &currency2, &value1, &value2)
	if err != nil {
		return err
	}

	d.Rates = append(d.Rates, schema.Rate{
		Date:      time.Unix(date, 0),
		Currency1: d.currenciesIndexI[currency1].Code,
		Currency2: d.currenciesIndexI[currency2].Code,
		Amount1:   d.currenciesIndexI[currency1].ConvertAmount(value1),
		Amount2:   d.currenciesIndexI[currency2].ConvertAmount(value2),
	})

	return nil
}

func (d *Database) readAccounts(uid int, fetch FetchFunc) error {
	var currencyId int
	account := schema.Account{}

	err := fetch(&uid, &account.Name, &account.InitBalance, &currencyId)
	if err != nil {
		return err
	}

	account.Currency = d.currenciesIndexI[currencyId].Code
	account.InitBalance = d.currenciesIndexI[currencyId].ConvertAmount(account.InitBalance)

	d.Accounts = append(d.Accounts, account)
	d.accountIndex[uid] = &account

	return nil
}

func (d *Database) readCategories(uid int, fetch FetchFunc) error {
	var parentId sql.NullInt32
	var name string

	err := fetch(&uid, &name, &parentId)
	if err != nil {
		return err
	}

	if parentId.Valid {
		name = d.categoriesIndex[int(parentId.Int32)] + "\\" + name
	}

	d.categoriesIndex[uid] = name

	return nil
}

func (d *Database) readTxCategories(uid int, fetch FetchFunc) error {
	var category, tx int

	err := fetch(&category, &tx)
	if err != nil {
		return err
	}

	_, ok := d.txCategoriesIndex[tx]
	if !ok {
		d.txCategoriesIndex[tx] = make([]int, 0)
	}
	d.txCategoriesIndex[tx] = append(d.txCategoriesIndex[tx], category)

	return nil
}

func (d *Database) readTxs(uid int, fetch FetchFunc) error {
	var iaccout, eaccount sql.NullInt32
	var iamount, eamount sql.NullFloat64
	var date int64
	var locked bool
	var comment string

	err := fetch(&uid, &date, &locked, &iaccout, &iamount, &eaccount, &eamount, &comment)
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
		account := d.accountIndex[int(iaccout.Int32)]

		item := ledger.TxItem{
			Account:  account.Name,
			Currency: account.Currency,
			Amount:   d.currenciesIndexS[account.Currency].ConvertAmount(iamount.Float64),
		}

		tx.Items = append(tx.Items, item)
	}

	if eaccount.Valid {
		account := d.accountIndex[int(eaccount.Int32)]

		item := ledger.TxItem{
			Account:  account.Name,
			Currency: account.Currency,
			Amount:   d.currenciesIndexS[account.Currency].ConvertAmount(eamount.Float64),
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

	if categories, ok := d.txCategoriesIndex[uid]; ok {
		for _, category := range categories {
			name := d.categoriesIndex[category]

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

	d.Transactions = append(d.Transactions, tx)

	return nil
}

func (c *Currency) ConvertAmount(amount float64) float64 {
	return amount / math.Pow(10, c.Precision+2)
}