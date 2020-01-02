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
	Path              string            `json:"path"`
	Target            string            `json:"target"`
	Accounts          map[string]string `json:"accounts"`
	Classifiers       []string          `json:"classifiers"`
	PrimaryClassifier string            `json:"primary_classifier"`
	accountNameLength int
	db                xml_schema.Database
}

func (d *Datafile) Export() (err error) {
	if err = d.readXmlDatabase(); err != nil {
		return
	}

	for _, accountName := range d.Accounts {
		l := utf8.RuneCountInString(accountName)
		if l > d.accountNameLength {
			d.accountNameLength = l
		}
	}

	if err = d.exportEntity("rates", d.db); err != nil {
		return
	}
	if err = d.exportEntity("txs", d.db.LedgerTransactions(d.PrimaryClassifier)); err != nil {
		return
	}

	return
}

func (d *Datafile) exportEntity(entityName string, data interface{}) error {
	t, err := getTemplate(entityName, template.FuncMap{
		"acc": d.account,
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
	}

	d.Classifiers = make([]string, 0)
	for _, c := range d.db.Classifiers {
		d.Classifiers = append(d.Classifiers, c.Name)
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

func getTemplate(name string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(fmt.Sprintf("%s.go.tmpl", name)).
		Funcs(funcs).
		Funcs(template.FuncMap{
			"signed": signed,
		}).
		ParseFiles(fmt.Sprintf("templates/%s.go.tmpl", name))
}

func (d *Datafile) account(name string) string {
	format := fmt.Sprintf("%%-%ds", d.accountNameLength)

	account, ok := d.Accounts[name]
	if !ok {
		account = name
	}

	return fmt.Sprintf(format, account)
}

func signed(amount float64) string {
	return fmt.Sprintf("% g", amount)
}
