package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
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
	accountNameLength int
	db                *xml_schema.Database
}

func (d *Datafile) Export(reader func(path string) *xml_schema.Database) {
	d.db = reader(d.Path)

	for _, accountName := range d.Accounts {
		l := utf8.RuneCountInString(accountName)
		if l > d.accountNameLength {
			d.accountNameLength = l
		}
	}

	outFilePrefix := strings.Replace(path.Base(d.Path), path.Ext(d.Path), "", 1)

	err := d.exportEntity(outFilePrefix, "rates", d.db)
	if err != nil {
		log.Fatal(err)
	}

	err = d.exportEntity(outFilePrefix, "txs", d.db.LedgerTransactions())
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Datafile) exportEntity(outFilePrefix string, entityName string, data interface{}) error {
	t, err := getTemplate(entityName, template.FuncMap{
		"acc": d.account,
	})

	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s-%s.dat", outFilePrefix, entityName))

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

func readXmlDatabase(path string) *xml_schema.Database {
	ensureFileExist(path)

	data, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	db := xml_schema.Database{}

	if err = xml.Unmarshal(data, &db); err != nil {
		log.Fatal(err)
	}

	return &db
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
	return fmt.Sprintf(format, d.Accounts[name])
}

func signed(amount float64) string {
	return fmt.Sprintf("% g", amount)
}
