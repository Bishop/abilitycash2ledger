package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"unicode/utf8"

	"github.com/Bishop/abilitycash2ledger/xml_schema"
)

type Scope struct {
	Datafiles []*Datafile `json:"datafile"`
}

type Datafile struct {
	Active            bool                         `json:"active"`
	Path              string                       `json:"path"`
	Target            string                       `json:"target"`
	Accounts          map[string]string            `json:"accounts"`
	Classifiers       map[string]map[string]string `json:"classifiers"`
	PrimaryClassifier string                       `json:"primary_classifier"`
	AccountNameLength int                          `json:"account_name_length"`
	db                xml_schema.Database
}

func (d *Datafile) Export() (err error) {
	if !d.Active {
		return nil
	}

	if err = d.readXmlDatabase(); err != nil {
		return
	}

	if err = d.exportEntity("rates", d.db); err != nil {
		return
	}

	converter := xml_schema.LedgerConverter{
		Accounts:          d.Accounts,
		Classifiers:       d.Classifiers,
		AccountClassifier: d.PrimaryClassifier,
		Db:                &d.db,
	}

	if err = d.exportEntity("txs", converter.Transactions()); err != nil {
		return
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

func (d *Datafile) Validate() ([]string, error) {
	messages := make([]string, 0)

	if err := d.readXmlDatabase(); err != nil {
		return nil, err
	}

	if len(d.db.AccountPlans) != 1 {
		return nil, errors.New("something wrong with accounts plans")
	}

	accounts := d.db.AccountPlans[0].Mappings(func(duplicate string) {
		messages = append(messages, fmt.Sprintf("duplicate account name: %s", duplicate))
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

	messages = append(messages, fmt.Sprintf("file %s is ok; found %d transactions\n", d.Path, len(d.db.Transactions)))

	return messages, nil
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
