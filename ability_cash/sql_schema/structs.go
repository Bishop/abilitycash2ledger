package sql_schema

import (
	"database/sql"
	"math"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type Database struct {
	Rates             []schema.Rate
	Accounts          []schema.Account
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
	db.Accounts = make([]schema.Account, 0)
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
		Currency1: d.currency(currency1).Code,
		Currency2: d.currency(currency2).Code,
		Amount1:   d.currency(currency1).ConvertAmount(value1),
		Amount2:   d.currency(currency2).ConvertAmount(value2),
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

	account.Currency = d.currency(currencyId).Code
	account.InitBalance = d.currency(currencyId).ConvertAmount(account.InitBalance)

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
		Note:     comment,
		Cleared:  locked,
		Metadata: make(map[string]string),
		Items:    []ledger.TxItem{},
	}

	if iaccout.Valid {
		tx.Items = append(tx.Items, d.makeTxItem(iaccout, iamount))
	}

	if eaccount.Valid {
		tx.Items = append(tx.Items, d.makeTxItem(eaccount, eamount))
	}

	if categories, ok := d.txCategoriesIndex[uid]; ok {
		for _, category := range categories {
			name := d.categoriesIndex[category]
			tx.Metadata[schema.CategoryClassifier(name)] = name
		}
	}

	d.Transactions = append(d.Transactions, tx)

	return nil
}

func (d *Database) makeTxItem(accountId sql.NullInt32, amount sql.NullFloat64) ledger.TxItem {
	account := d.accountIndex[int(accountId.Int32)]

	return ledger.TxItem{
		Account:  account.Name,
		Currency: account.Currency,
		Amount:   d.currency(account.Currency).ConvertAmount(amount.Float64),
	}
}

func (d *Database) currency(id any) *Currency {
	switch v := id.(type) {
	case int:
		return d.currenciesIndexI[v]
	case string:
		return d.currenciesIndexS[v]
	default:
		panic("Unknown id type")
	}
}

func (c *Currency) ConvertAmount(amount float64) float64 {
	return amount / math.Pow(10, c.Precision+2)
}
