package scope

import (
	"errors"
	"fmt"
	"github.com/Bishop/abilitycash2ledger/ability_cash/sql_schema"
	"os"
	"path"
	"text/template"

	"github.com/Bishop/abilitycash2ledger/ability_cash"
	"github.com/Bishop/abilitycash2ledger/ability_cash/csv_schema"
	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ability_cash/xml_schema"
)

type datafile struct {
	Active bool   `json:"active"`
	Equity bool   `json:"equity"`
	Path   string `json:"path"`
	Target string `json:"target"`
	db     schema.Database
}

func (d *datafile) readDb() (schema.Database, error) {
	switch d.format() {
	case ".xml":
		return xml_schema.ReadDatabase(d.Path)
	case "", ".csv":
		return csv_schema.ReadDatabase(d.Path)
	case ".cash":
		return sql_schema.ReadDatabase(d.Path)
	default:
		return nil, errors.New("unknown format")
	}
}

func (d *datafile) format() string {
	return path.Ext(d.Path)
}

func (d *datafile) export() (err error) {
	if err = d.exportEntity("rates", d.db.GetRates()); err != nil {
		return
	}

	converter := &ability_cash.LedgerConverter{
		GenerateEquity: d.Equity,
		Db:             d.db,
	}

	err = d.exportEntity("txs", converter.Transactions())

	if err = d.exportEntity("accounts", converter.Accounts()); err != nil {
		return err
	}

	return
}

func (d *datafile) exportEntity(entityName string, data interface{}) error {
	t, err := getTemplate(entityName)

	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s-%s.journal", d.Target, entityName))

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

func getTemplate(name string) (*template.Template, error) {
	return template.New(fmt.Sprintf("%s.go.tmpl", name)).
		Funcs(template.FuncMap{
			"acc":    acc,
			"signed": signed,
		}).
		ParseFiles(fmt.Sprintf("templates/%s.go.tmpl", name))
}

func acc(account string) string {
	return fmt.Sprintf("%-40s", account)
}

func signed(amount float64) string {
	// suppress exponent format floats
	// print 110778000, not 1.10778e+08, and not 110778000.000000
	return fmt.Sprintf("% .10g", amount)
}
