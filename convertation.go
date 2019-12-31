package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/Bishop/abilitycash2ledger/xml_schema"
)

func export(db *xml_schema.Database, outFilePrefix string) error {
	for _, entity := range []string{"rates", "txs"} {
		err := exportEntity(db, outFilePrefix, entity)

		if err != nil {
			return err
		}
	}

	return nil
}

func exportEntity(db *xml_schema.Database, outFilePrefix string, entity string) error {
	t, err := getTemplate(entity)

	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s-%s.dat", outFilePrefix, entity))

	if err != nil {
		return err
	}

	err = t.Execute(file, db)

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
		ParseFiles(fmt.Sprintf("templates/%s.go.tmpl", name))
}
