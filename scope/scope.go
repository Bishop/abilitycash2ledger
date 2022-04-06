package scope

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

func NewScope() *scope {
	s := new(scope)

	s.Categories = map[string]string{"Payee": "payee", "Expenses": "account", "Income": "account"}

	return s
}

type scope struct {
	Datafiles  []*datafile       `json:"datafiles"`
	Categories map[string]string `json:"categories"`
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

	_ = s.iterateDatafiles(func(d *datafile) error {
		_ = *d.db.GetAccounts()

		messages = append(messages, fmt.Sprintf("file %s is ok; found %d transactions\n", d.Path, len(*d.db.GetTransactions())))

		return nil
	})

	return messages, nil
}

func (s *scope) Export() error {
	return s.iterateDatafiles(func(d *datafile) error {
		return d.export(s.Categories)
	})
}

func (s *scope) iterateDatafiles(callback func(*datafile) error) error {
	var err error

	for _, datafile := range s.Datafiles {
		if !datafile.Active {
			continue
		}
		if datafile.db, err = datafile.readDb(); err != nil {
			return err
		}

		err = callback(datafile)

		if err != nil {
			return err
		}
	}

	return nil
}
