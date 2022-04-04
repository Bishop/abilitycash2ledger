package scope

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

func NewScope() *scope {
	return new(scope)
}

type scope struct {
	Datafiles []*datafile `json:"datafiles"`
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
	var err error

	messages := make([]string, 0)

	for _, datafile := range s.Datafiles {
		if !datafile.Active {
			continue
		}

		if datafile.db, err = datafile.readDb(); err != nil {
			return nil, err
		}

		_ = *datafile.db.GetAccounts()
		_ = *datafile.db.GetClassifiers()

		messages = append(messages, fmt.Sprintf("file %s is ok; found %d transactions\n", datafile.Path, len(*datafile.db.GetTransactions())))
	}

	return messages, nil
}

func (s *scope) Export() error {
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
