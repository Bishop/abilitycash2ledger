package scope

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/Bishop/abilitycash2ledger/ability_cash"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/Bishop/abilitycash2ledger/ability_cash/xml_schema"
)

type Scope struct {
	Datafiles []*Datafile                    `json:"datafiles"`
	Common    map[string]EmbeddedClassifiers `json:"common"`
}

type EmbeddedClassifiers struct {
	Accounts    map[string]string            `json:"accounts"`
	Classifiers map[string]map[string]string `json:"classifiers"`
}

type Datafile struct {
	Active bool   `json:"active"`
	Equity bool   `json:"equity"`
	Path   string `json:"path"`
	Target string `json:"target"`
	EmbeddedClassifiers
	PrimaryClassifier string `json:"primary_classifier"`
	CommonClassifiers string `json:"common_classifiers"`
	AccountNameLength int    `json:"account_name_length"`
	db                xml_schema.Database
}

func (s *Scope) AddFile(name string) error {
	for _, df := range s.Datafiles {
		if df.Path == name {
			return errors.New(fmt.Sprintf("newPath %s already in the list", name))
		}
	}

	s.Datafiles = append(s.Datafiles, &Datafile{
		Active: true,
		Equity: true,
		Path:   name,
		Target: strings.Replace(name, path.Ext(name), "", 1),
	})

	return nil
}

func (s *Scope) Validate() ([]string, error) {
	messages := make([]string, 0)

	if s.Common == nil {
		s.Common = make(map[string]EmbeddedClassifiers)
	}

	for _, datafile := range s.Datafiles {
		if err := datafile.readXmlDatabase(); err != nil {
			return nil, err
		}

		if datafile.CommonClassifiers != "" {
			datafile.EmbeddedClassifiers = s.Common[datafile.CommonClassifiers]
		}

		err := datafile.validate(&messages)

		if err != nil {
			return nil, err
		}

		if datafile.CommonClassifiers != "" {
			s.Common[datafile.CommonClassifiers] = datafile.EmbeddedClassifiers
			datafile.EmbeddedClassifiers = EmbeddedClassifiers{}
		}
	}

	return messages, nil
}

func (s *Scope) Export() error {
	if s.Common == nil {
		s.Common = make(map[string]EmbeddedClassifiers)
	}

	for _, datafile := range s.Datafiles {
		if !datafile.Active {
			continue
		}
		if err := datafile.readXmlDatabase(); err != nil {
			return err
		}
		if datafile.CommonClassifiers != "" {
			datafile.EmbeddedClassifiers = s.Common[datafile.CommonClassifiers]
		}
		if err := datafile.export(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Datafile) export() (err error) {
	if err = d.exportEntity("rates", d.db); err != nil {
		return
	}

	converter := ability_cash.LedgerConverter{
		Accounts:          d.Accounts,
		Classifiers:       d.Classifiers,
		AccountClassifier: d.PrimaryClassifier,
		GenerateEquity:    d.Equity,
		Db:                &d.db,
	}

	if err = d.exportEntity("txs", converter.Transactions()); err != nil {
		return
	}

	// $ ledger accounts
	if err = d.exportEntity("accounts", converter.AccountsList()); err != nil {
		return err
	}

	return
}

func (d *Datafile) exportEntity(entityName string, data interface{}) error {
	format := fmt.Sprintf("%%-%ds", d.AccountNameLength)

	t, err := getTemplate(entityName, template.FuncMap{
		"acc": func(name string) string {
			return fmt.Sprintf(format, name)
		},
	})

	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s-%s.dat", d.Target, entityName))

	if err != nil {
		return err
	}

	err = t.Execute(file, data)

	if err != nil {
		return err
	}

	err = file.Close()

	if err != nil {
		return err
	}

	return nil
}

func (d *Datafile) validate(messages *[]string) error {
	if len(d.db.AccountPlans) != 1 {
		return errors.New("something wrong with accounts plans")
	}

	accounts := d.db.AccountPlans[0].Mappings(func(duplicate string) {
		*messages = append(*messages, fmt.Sprintf("duplicate account name: %s", duplicate))
	})

	if d.Accounts == nil {
		d.Accounts = make(map[string]string)
	}

	for accountShort, accountFull := range accounts {
		if _, ok := d.Accounts[accountShort]; !ok {
			d.Accounts[accountShort] = accountFull
		}
		d.checkAccountLength(d.Accounts[accountShort])
	}

	if d.Classifiers == nil {
		d.Classifiers = make(map[string]map[string]string)
	}

	for _, c := range d.db.Classifiers {
		if _, ok := d.Classifiers[c.Name]; !ok {
			d.Classifiers[c.Name] = make(map[string]string)
		}

		for category := range c.Categories() {
			if _, ok := d.Classifiers[c.Name][category]; !ok {
				d.Classifiers[c.Name][category] = category
			}
			d.checkAccountLength(d.Classifiers[c.Name][category])
		}
	}

	*messages = append(*messages, fmt.Sprintf("file %s is ok; found %d transactions\n", d.Path, len(d.db.Transactions)))

	return nil
}

func (d *Datafile) readXmlDatabase() error {
	data, err := ioutil.ReadFile(d.Path)

	if err != nil {
		return err
	}

	if err = xml.Unmarshal(data, &d.db); err != nil {
		return err
	}

	return nil
}

func (d *Datafile) checkAccountLength(s string) {
	l := utf8.RuneCountInString(s)
	if l > d.AccountNameLength {
		d.AccountNameLength = l
	}
}

func getTemplate(name string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(fmt.Sprintf("%s.go.tmpl", name)).
		Funcs(funcs).
		Funcs(template.FuncMap{
			"signed": signed,
		}).
		ParseFiles(fmt.Sprintf("templates/%s.go.tmpl", name))
}

func signed(amount float64) string {
	return fmt.Sprintf("% g", amount)
}
