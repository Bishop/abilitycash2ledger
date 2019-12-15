package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/Bishop/abilitycash2ledger/xml_schema"
	"github.com/urfave/cli/v2"
)

const scopeFile = "./scope.json"

var scope = Scope{}

func main() {
	readScope()

	app := cli.App{
		Name:    "abilitycash db to ledger converter",
		Version: "0.0.1",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add datafile to scope",
				Action:  add,
			},
		},
		Action: process,
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
	return nil
}

func process(c *cli.Context) error {
	path := c.String("datafile")

	ensureFileExist(path)

	data, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	db := xml_schema.Database{}

	if err = xml.Unmarshal(data, &db); err != nil {
		return err
	}

	log.Printf("%+v", db.Accounts)

	for _, tx := range db.Transactions {
		if tx.Income == nil && tx.Expense == nil && tx.Transfer == nil {
			dump, _ := json.MarshalIndent(tx, "", "  ")

			log.Println(string(dump))
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
