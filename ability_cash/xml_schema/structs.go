package xml_schema

import (
	"encoding/xml"
)

type Database struct {
	XMLName      xml.Name      `xml:"ability-cash"`
	Currencies   []Currency    `xml:"currencies>currency"`
	Rates        []Rate        `xml:"rates>rate"`
	Accounts     []Account     `xml:"accounts>account"`
	AccountPlans []AccountPlan `xml:"account-plans>account-plan"`
	Transactions []Transaction `xml:"transactions>transaction"`
	Classifiers  []Classifier  `xml:"classifiers>classifier"`
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

func (d *Database) GetAccounts() *[]Account {
	return &d.Accounts
}

func (d *Database) GetTransactions() *[]Transaction {
	return &d.Transactions
}

func (d *Database) GetClassifiers() *[]Classifier {
	return &d.Classifiers
}

func (d *Database) GetAccountPlans() *[]AccountPlan {
	return &d.AccountPlans
}
