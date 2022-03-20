package scope

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
)

func NewScope() *scope {
	return new(scope)
}

type scope struct {
	Datafiles   []*datafile           `json:"datafiles"`
	Accounts    schema.AccountsMap    `json:"accounts"`
	Classifiers schema.ClassifiersMap `json:"classifiers"`
}

func (s *scope) AddFile(name string) error {
	for _, df := range s.Datafiles {
		if df.Path == name {
			return errors.New(fmt.Sprintf("path %s already in the list", name))
		}
	}

	s.Datafiles = append(s.Datafiles, &datafile{
		Active: true,
		Equity: true,
		Path:   name,
		Target: strings.TrimSuffix(name, path.Ext(name)),
	})

	return nil
}

func (s *scope) Validate() ([]string, error) {
	messages := make([]string, 0)

	s.init()

	var err error

	for _, datafile := range s.Datafiles {
		if !datafile.Active {
			continue
		}

		if datafile.db, err = datafile.readDb(); err != nil {
			return nil, err
		}

		for _, account := range *datafile.db.GetAccounts() {
			if _, ok := s.Accounts[account.Name]; !ok {
				s.Accounts[account.Name] = account.Name
			}
		}

		for name, c := range *datafile.db.GetClassifiers() {
			if _, ok := s.Classifiers[name]; !ok {
				s.Classifiers[name] = make(schema.AccountsMap)
			}

			for _, category := range c {
				if _, ok := s.Classifiers[name][category]; !ok {
					s.Classifiers[name][category] = category
				}
			}
		}

		messages = append(messages, fmt.Sprintf("file %s is ok; found %d transactions\n", datafile.Path, len(*datafile.db.GetTransactions())))

		if err != nil {
			return nil, err
		}
	}

	return messages, nil
}

func (s *scope) Export() error {
	s.init()

	var err error

	for _, datafile := range s.Datafiles {
		if !datafile.Active {
			continue
		}
		if datafile.db, err = datafile.readDb(); err != nil {
			return err
		}

		if err := datafile.export(s); err != nil {
			return err
		}
	}
	return nil
}

func (s *scope) init() {
	if s.Accounts == nil {
		s.Accounts = make(schema.AccountsMap)
	}

	if s.Classifiers == nil {
		s.Classifiers = make(schema.ClassifiersMap)
	}
}
