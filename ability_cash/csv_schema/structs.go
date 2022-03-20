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
// 10         11           12       13                 14                    15
// Recurrence,Day of month,Interval,Category of Статья,Category of Поставщик,Category of Агент

type Database struct {
	Transactions []ledger.Transaction
}

func (d *Database) Fill(record []string) {
	if d.Transactions == nil {
		d.Transactions = make([]ledger.Transaction, 0)
	}

	tx := ledger.Transaction{
		Date:     parseDate(record[2]),
		Note:     record[9],
		Executed: record[0] == "+",
		Cleared:  record[0] == "+",
		Metadata: make(map[string]string),
	}

	if record[13] != "" {
		tx.Metadata["Category"] = record[13][1:]
	}
	if record[14] != "" {
		tx.Metadata["Provider"] = record[14][1:]
	}
	if record[15] != "" {
		tx.Metadata["Beneficiary"] = record[15][1:]
	}

	if record[3] != "" && record[6] != "" {
		tx.Items = []ledger.TxItem{
			txItemFromStrings(record[3], record[4]),
			txItemFromStrings(record[6], record[7]),
		}
	} else if record[3] != "" {
		tx.Items = []ledger.TxItem{
			txItemFromStrings(record[3], record[4]),
			{
				Account: record[13][1:],
			},
		}
	} else {
		item := txItemFromStrings(record[6], record[7])
		tx.Items = []ledger.TxItem{
			{
				Account: item.Account,
			},
			{
				Account:  record[13][1:],
				Currency: item.Currency,
				Amount:   -item.Amount,
			},
		}
	}

	d.Transactions = append(d.Transactions, tx)
}

func (d *Database) GetAccounts() *[]schema.Account {
	accounts := make([]schema.Account, 0)

	return &accounts
}

func (d *Database) GetTransactions() *[]ledger.Transaction {
	return &d.Transactions
}

func (d *Database) GetClassifiers() *schema.ClassifiersList {
	classifiers := make(schema.ClassifiersList)

	return &classifiers
}

func (d *Database) GetRates() *[]schema.Rate {
	rates := make([]schema.Rate, 0)

	return &rates
}

func parseDate(s string) time.Time {
	const format = "2006-01-02" // 2011-01-01

	parse, err := time.ParseInLocation(format, s, time.Local)

	if err != nil {
		log.Fatalln(err)
	}

	return parse
}

func txItemFromStrings(accountString, amountString string) ledger.TxItem {
	accountParts := strings.SplitN(accountString, " - ", 2)
	amount, err := strconv.ParseFloat(amountString, 64)

	if err != nil {
		log.Fatalln(err)
	}

	return ledger.TxItem{
		Account:  accountParts[1],
		Currency: accountParts[0],
		Amount:   amount,
	}
}
