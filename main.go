package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Bishop/abilitycash2ledger/xml_schema"
	"github.com/urfave/cli/v2"
)

const scopeFile = "./scope.json"

var scope = Scope{}

func main() {
	readScope()

	app := cli.App{
		Name:    "abilitycash2ledger",
		Usage:   "abilitycash db to ledger converter",
		Version: "0.0.2",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add datafile to scope",
				Action:  add,
			},
			{
				Name:    "prepare",
				Aliases: []string{"p"},
				Usage:   "Analyze datafiles and fill config file",
				Action:  prepare,
			},
			{
				Name:    "convert",
				Aliases: []string{"c"},
				Usage:   "Convert added datafiles to ledger format",
				Action:  convert,
			},
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func add(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("path to datafile is needed for add command")
	}

	path := c.Args().First()

	ensureFileExist(path)

	for _, df := range scope.Datafiles {
		if df.Path == path {
			log.Printf("path %s already in the list", path)
			return nil
		}
	}

	scope.Datafiles = append(scope.Datafiles, &Datafile{Path: path})

	saveScope()

	return nil
}

func prepare(c *cli.Context) error {
	for _, datafile := range scope.Datafiles {
		dataFile := datafile.Path

		db := readXmlDatabase(dataFile)

		datafile.Accounts = make([]string, len(db.Accounts))
		for i, account := range db.Accounts {
			datafile.Accounts[i] = fmt.Sprintf("%s - %s", account.Name, account.Currency)
		}

		for _, tx := range db.Transactions {
			if tx.Income == nil && tx.Expense == nil && tx.Transfer == nil && tx.Balance == nil {
				dump, _ := json.MarshalIndent(tx, "", "  ")

				log.Println(string(dump))
			}
		}

		fmt.Printf("file: %s\n%d transactions\n", dataFile, len(db.Transactions))
	}

	saveScope()

	return nil
}

func convert(c *cli.Context) error {
	for _, datafile := range scope.Datafiles {
		dataFile := datafile.Path

		db := readXmlDatabase(dataFile)

		outFile := strings.Replace(path.Base(dataFile), path.Ext(dataFile), "", 1)

		t, err := template.New("rates.go.tmpl").ParseFiles("templates/rates.go.tmpl")

		if err != nil {
			return err
		}

		file, err := os.Create(fmt.Sprintf("%s-rates.dat", outFile))

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
	}

	return nil
}

func ensureFileExist(path string) {
	if !checkFileExist(path) {
		log.Fatalf("File %v does not exist\n", path)
	}
}

func checkFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func readScope() {
	if !checkFileExist(scopeFile) {
		return
	}

	readConfig(scopeFile, &scope)
}

func saveScope() {
	if err := saveConfig(scopeFile, &scope); err != nil {
		log.Fatal(err)
	}
}

func readConfig(filename string, config interface{}) {
	data, _ := ioutil.ReadFile(filename)

	err := json.Unmarshal(data, config)

	if err != nil {
		log.Fatal(err)
	}

	return
}

func saveConfig(filename string, config interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0600)
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
