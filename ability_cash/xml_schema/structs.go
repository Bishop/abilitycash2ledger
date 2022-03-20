package xml_schema

import (
	"encoding/xml"
	"log"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ledger"
)

type Database struct {
	XMLName      xml.Name      `xml:"ability-cash"`
	Currencies   []Currency    `xml:"currencies>currency"`
	Rates        []Rate        `xml:"rates>rate"`
	Accounts     []Account     `xml:"accounts>account"`
	AccountPlans []AccountPlan `xml:"account-plans>account-plan"`
	Transactions []Transaction `xml:"transactions>transaction"`
	Classifiers  []Classifier  `xml:"classifiers>classifier"`
	AccountsMap  schema.AccountsMap
}

type Currency struct {
	item
	Name      string `xml:"name"`
	Code      string `xml:"code"`
	Precision uint   `xml:"precision"`
}

type Rate struct {
	item
	Date      acDate  `xml:"date"`
	Currency1 string  `xml:"currency-1"`
	Currency2 string  `xml:"currency-2"`
	Amount1   float64 `xml:"amount-1"`
	Amount2   float64 `xml:"amount-2"`
}

type Account struct {
	item
	Name        string  `xml:"name"`
	Currency    string  `xml:"currency"`
	InitBalance float64 `xml:"init-balance"`
}

type AccountPlan struct {
	item
	Name     string        `xml:"name"`
	Comment  string        `xml:"comment"`
	Accounts []Account     `xml:"account"`
	Folders  []AccountPlan `xml:"folder"`
}

type Classifier struct {
	item
	Name       string         `xml:"singular-name"`
	PluralName string         `xml:"plural-name"`
	Income     []txCategoryTI `xml:"income-tree>category"`
	Expense    []txCategoryTI `xml:"expense-tree>category"`
	Single     []txCategoryTI `xml:"single-tree>category"`
}

type Transaction struct {
	item
	Date     acDate    `xml:"date"`
	Comment  string    `xml:"comment"`
	Transfer *Transfer `xml:"transfer"`
	Income   *Income   `xml:"income"`
	Expense  *Expense  `xml:"expense"`
	Balance  *Balance  `xml:"balance"`
}

type Transfer struct {
	txItem
	txIncome
	txExpense
}

type Income struct {
	txItem
	txIncome
	Categories txCategories `xml:"category"`
}

type Expense struct {
	txItem
	txExpense
	Categories txCategories `xml:"category"`
}

type Balance struct {
	txItem
	txIncome
}

type item struct {
	Oid       string `xml:"oid,attr"`
	ChangedAt acTime `xml:"changed-at,attr"`
}

type txItem struct {
	item
	Executed *struct{} `xml:"executed"`
	Locked   *struct{} `xml:"locked"`
}

type txAccount struct {
	Name     string `xml:"name"`
	Currency string `xml:"currency"`
}

type txIncome struct {
	IncomeAccount txAccount `xml:"income-account"`
	IncomeAmount  float64   `xml:"income-amount"`
	IncomeBalance float64   `xml:"income-balance"`
}

type txExpense struct {
	ExpenseAccount txAccount `xml:"expense-account"`
	ExpenseAmount  float64   `xml:"expense-amount"`
	ExpenseBalance float64   `xml:"expense-balance"`
}

type txCategory struct {
	Classifier string      `xml:"classifier,attr"`
	Name       string      `xml:"name"`
	Category   *txCategory `xml:"category"`
}

type txCategories []txCategory

type txCategoryTI struct {
	Name       string          `xml:"name"`
	Categories *[]txCategoryTI `xml:"category"`
}

func (d *Database) GetAccounts() *[]schema.Account {
	d.cacheAccountsMap()

	accounts := make([]schema.Account, len(d.Accounts))

	for i, account := range d.Accounts {
		accounts[i] = schema.Account{
			Name:        d.account(account.Name),
			Currency:    account.Currency,
			InitBalance: account.InitBalance,
		}
	}

	return &accounts
}

func (d *Database) GetRates() *[]schema.Rate {
	rates := make([]schema.Rate, len(d.Rates))

	for i, rate := range d.Rates {
		rates[i] = schema.Rate{
			Date:      rate.Date.Source(),
			Currency1: rate.Currency1,
			Currency2: rate.Currency2,
			Amount1:   rate.Amount1,
			Amount2:   rate.Amount2,
		}
	}

	return &rates
}

func (d *Database) GetTransactions() *[]ledger.Transaction {
	d.cacheAccountsMap()

	txs := make([]ledger.Transaction, len(d.Transactions))

	for i, source := range d.Transactions {
		if !source.IsExecuted() {
			continue
		}

		tx := ledger.Transaction{
			Date:    source.Date.Source(),
			Note:    source.Comment,
			Cleared: source.IsLocked(),
		}

		switch {
		case source.Transfer != nil:
			tx.Items = []ledger.TxItem{
				{
					Account: d.account(source.Transfer.ExpenseAccount.Name),
				},
				{
					Account:  d.account(source.Transfer.IncomeAccount.Name),
					Currency: source.Transfer.IncomeAccount.Currency,
					Amount:   source.Transfer.IncomeAmount,
				},
			}

			if source.Transfer.IncomeAccount.Currency != source.Transfer.ExpenseAccount.Currency {
				tx.Items[0].Currency = source.Transfer.ExpenseAccount.Currency
				tx.Items[0].Amount = source.Transfer.ExpenseAmount
			}
		case source.Expense != nil:
			tx.Metadata = source.Expense.Categories.Map()
			tx.Items = []ledger.TxItem{
				{
					Account: d.account(source.Expense.ExpenseAccount.Name),
				},
				{
					Account:  d.accountFromCategories(tx.Metadata),
					Currency: source.Expense.ExpenseAccount.Currency,
					Amount:   -source.Expense.ExpenseAmount,
				},
			}
		case source.Income != nil:
			tx.Metadata = source.Income.Categories.Map()
			tx.Items = []ledger.TxItem{
				{
					Account:  d.account(source.Income.IncomeAccount.Name),
					Currency: source.Income.IncomeAccount.Currency,
					Amount:   source.Income.IncomeAmount,
				},
				{
					Account: d.accountFromCategories(tx.Metadata),
				},
			}
		case source.Balance != nil:
			tx.Items = []ledger.TxItem{
				{
					Account:          d.account(source.Balance.IncomeAccount.Name),
					Currency:         source.Balance.IncomeAccount.Currency,
					BalanceAssertion: source.Balance.IncomeBalance,
				},
				{
					Account: ledger.Adjustment,
				},
			}
		}

		txs[i] = tx
	}

	return &txs
}

func (d *Database) GetClassifiers() *schema.ClassifiersList {
	classifiers := make(schema.ClassifiersList)

	for _, classifier := range d.Classifiers {
		classifiers[classifier.Name] = make([]string, 0)

		for category := range classifier.Categories() {
			classifiers[classifier.Name] = append(classifiers[classifier.Name], category)
		}
	}

	return &classifiers
}

func (d *Database) account(a string) string {
	account, ok := d.AccountsMap[a]
	if ok {
		return account
	} else {
		return a
	}
}

func (d *Database) accountFromCategories(classifier map[string]string) string {
	return classifier["Статья"]
}

func (d *Database) cacheAccountsMap() {
	if d.AccountsMap != nil {
		return
	}

	if len(d.AccountPlans) != 1 {
		log.Fatalln("something wrong with accounts plans")
	}

	d.AccountsMap = d.AccountPlans[0].Check(func(duplicate string) {
		log.Fatalf("duplicate account name: %s", duplicate)
	})
}
