package xml_schema

func (a *AccountPlan) Mappings(logger func(s string)) map[string]string {
	a.mapping = make(map[string]string)

	a.fillAccounts("", a.mapping, logger)

	return a.mapping
}

func (a *AccountPlan) fillAccounts(prefix string, target map[string]string, logger func(s string)) {
	for _, account := range a.Accounts {
		if _, ok := target[account.Name]; ok {
			logger(account.Name)
		} else {
			target[account.Name] = prefix + account.Name
		}
	}

	for _, folder := range a.Folders {
		folder.fillAccounts(folder.Name+":", target, logger)
	}
}
