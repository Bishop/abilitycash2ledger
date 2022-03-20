package xml_schema

import "github.com/Bishop/abilitycash2ledger/ability_cash/schema"

func (a *AccountPlan) Check(logger func(s string)) schema.AccountsMap {
	return a.mappings(logger)
}

func (a *AccountPlan) Mappings() schema.AccountsMap {
	mapping := make(map[string]string)

	a.fillAccounts("", mapping, func(string) {})

	return mapping
}

func (a *AccountPlan) mappings(logger func(s string)) schema.AccountsMap {
	mapping := make(schema.AccountsMap)

	a.fillAccounts("", mapping, logger)

	return mapping
}

func (a *AccountPlan) fillAccounts(prefix string, target schema.AccountsMap, logger func(s string)) {
	for _, account := range a.Accounts {
		if _, ok := target[account.Name]; ok {
			logger(account.Name)
		} else {
			target[account.Name] = prefix + account.Name
		}
	}

	for _, folder := range a.Folders {
		folder.fillAccounts(folder.Name+"\\", target, logger)
	}
}
