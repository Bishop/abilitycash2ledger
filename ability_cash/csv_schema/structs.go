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
	Classifiers  schema.ClassifiersList
	Accounts     []schema.Account
	AccountsMap  schema.AccountsMap
	Transactions []ledger.Transaction
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

	return db
}

func (d *Database) AddTx(record []string) {
	tx := ledger.Transaction{
		Date:     parseDate(record[2]),
		Note:     record[9],
		Executed: record[0] == "+",
		Cleared:  record[1] == "+",
		Metadata: make(map[string]string),
	}

	if record[10] != "" {
		tx.Metadata[schema.ExpensesClassifier] = record[10][1:]
	}
	if record[11] != "" {
		tx.Metadata[schema.PayeeClassifier] = record[11][1:]
	}
	if record[12] != "" {
		tx.Metadata["Agent"] = record[12][1:]
	}

	if record[3] != "" && record[6] != "" {
		tx.Items = []ledger.TxItem{
			d.txItemFromStrings(record[3], record[4]),
			d.txItemFromStrings(record[6], record[7]),
		}

		if tx.Items[0].Currency == tx.Items[1].Currency {
			tx.Payee = "Transfer"
		} else {
			tx.Payee = "Exchange"
		}
	} else if record[3] != "" {
		tx.Items = []ledger.TxItem{
			d.txItemFromStrings(record[3], record[4]),
			{
				Account: d.account(tx.Metadata[schema.ExpensesClassifier]),
			},
		}
	} else {
		item := d.txItemFromStrings(record[6], record[7])
		tx.Items = []ledger.TxItem{
			{
				Account: item.Account,
			},
			{
				Account:  d.account(tx.Metadata[schema.ExpensesClassifier]),
				Currency: item.Currency,
				Amount:   -item.Amount,
			},
		}
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

func (d *Database) AddCategory(record []string) {
	category := record[0][1:]
	categoryParts := strings.SplitN(category, "\\", 2)

	switch categoryParts[0] {
	case "Income", "Expenses":
		d.Classifiers[schema.ExpensesClassifier] = append(d.Classifiers[schema.ExpensesClassifier], category)
	case "Payee":
		d.Classifiers[schema.PayeeClassifier] = append(d.Classifiers[schema.PayeeClassifier], category)
	case "Agents":
		d.Classifiers["Agent"] = append(d.Classifiers["Agent"], category)
	}
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

func (d *Database) GetClassifiers() *schema.ClassifiersList {
	return &d.Classifiers
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
