package schema

import "github.com/Bishop/abilitycash2ledger/ability_cash/xml_schema"

type Database interface {
	GetAccounts() *[]xml_schema.Account
	GetTransactions() *[]xml_schema.Transaction
	GetClassifiers() *[]xml_schema.Classifier
	GetAccountPlans() *[]xml_schema.AccountPlan
}

type Accounts map[string]string
