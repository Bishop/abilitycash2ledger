package sql_schema

import (
	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
	"math"
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

func (c *Currency) ConvertAmount(amount float64) float64 {
	return amount / math.Pow(10, c.Precision+2)
}
