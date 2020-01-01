package main

import (
	"fmt"
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

type view struct {
	Database *xml_schema.Database
	Accounts map[string]string
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

	for _, entity := range []string{"rates", "txs"} {
		err := d.exportEntity(outFilePrefix, entity)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (d *Datafile) exportEntity(outFilePrefix string, entity string) error {
	t, err := getTemplate(entity, template.FuncMap{
		"acc": d.account,
	})

	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s-%s.dat", outFilePrefix, entity))

	if err != nil {
		return err
	}

	err = t.Execute(file, view{
		Database: d.db,
		Accounts: d.Accounts,
	})

	if err != nil {
		return err
	}

	err = file.Close()

	if err != nil {
		return err
	}

	return nil
}

func (d *Datafile) account(name string) string {
	format := fmt.Sprintf("%%-%ds", d.accountNameLength)
	return fmt.Sprintf(format, d.Accounts[name])
}
