package csv_schema

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

// 0        1      2    3              4             5              6               7              8               9
// Executed,Locked,Date,Income account,Income amount,Income balance,Expense account,Expense amount,Expense balance,Comment,
// 10                   11                   12
// Category of Category,Category of Provider,Category of Agent
// 10         11           12       13                 14                    15
// Recurrence,Day of month,Interval,Category of Category,Category of Provider,Category of Agent

type Database struct {
	Rates        []schema.Rate
	Accounts     []schema.Account
	AccountsMap  schema.AccountsMap
	Transactions []ledger.Transaction
}

func NewDatabase() *Database {
	db := new(Database)

	db.Rates = make([]schema.Rate, 0)
	db.Accounts = make([]schema.Account, 0)
	db.AccountsMap = make(schema.AccountsMap)
	db.Transactions = make([]ledger.Transaction, 0)

	return db
}

func (d *Database) AddTx(record []string) {
	tx := ledger.Transaction{
		Date:     parseDate(record[2]),
		Note:     record[9],
		Executed: record[0] == "+",
		Cleared:  record[1] == "+",
		Metadata: make(map[string]string),
		Items:    []ledger.TxItem{},
	}

	for _, category := range []string{record[10], record[11], record[12]} {
		if category != "" {
			category = category[1:]
			tx.Metadata[schema.CategoryClassifier(category)] = category
		}
	}

	if record[3] != "" {
		tx.Items = append(tx.Items, d.txItemFromStrings(record[3], record[4]))
	}

	if record[6] != "" {
		tx.Items = append(tx.Items, d.txItemFromStrings(record[6], record[7]))
	}

	d.Transactions = append(d.Transactions, tx)
}

func (d *Database) AddRate(record []string) {
	rate := schema.Rate{
		Date:      parseDate(record[0]),
		Currency1: record[1],
		Currency2: record[3],
		Amount1:   parseFloat(record[2]),
		Amount2:   parseFloat(record[4]),
	}

	d.Rates = append(d.Rates, rate)
}

func (d *Database) AddAccountMap(record []string) {
	dir := strings.Replace(record[0], "\\Root", "", 1)
	dir = strings.Replace(dir, "\\", "", 1)

	account := record[1]

	if dir != "" {
		account = strings.Join([]string{dir, account}, "\\")
	}

	d.AccountsMap[record[1]] = account
}

func (d *Database) AddAccount(record []string) {
	account := schema.Account{
		Name:        d.account(record[0]),
		Currency:    record[1],
		InitBalance: parseFloat(record[2]),
	}

	d.Accounts = append(d.Accounts, account)
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

func (d *Database) account(a string) string {
	account, ok := d.AccountsMap[a]
	if ok {
		return account
	} else {
		return a
	}
}

func (d *Database) txItemFromStrings(accountString, amountString string) ledger.TxItem {
	accountParts := strings.SplitN(accountString, " - ", 2)

	return ledger.TxItem{
		Account:  d.account(accountParts[1]),
		Currency: accountParts[0],
		Amount:   parseFloat(amountString),
	}
}

func parseDate(s string) time.Time {
	const format = "2006-01-02" // 2011-01-01

	parse, err := time.ParseInLocation(format, s, time.Local)

	if err != nil {
		log.Fatalln(err)
	}

	return parse
}

func parseFloat(s string) float64 {
	amount, err := strconv.ParseFloat(s, 64)

	if err != nil {
		log.Fatalln(err)
	}

	return amount
}
