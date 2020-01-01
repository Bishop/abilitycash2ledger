package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

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
		return errors.New("newPath to datafile is needed for add command")
	}

	newPath := c.Args().First()

	ensureFileExist(newPath)

	for _, df := range scope.Datafiles {
		if df.Path == newPath {
			log.Printf("newPath %s already in the list", newPath)
			return nil
		}
	}

	scope.Datafiles = append(scope.Datafiles, &Datafile{Path: newPath,
		Target: strings.Replace(newPath, path.Ext(newPath), "", 1)})

	saveScope()

	return nil
}

func prepare(c *cli.Context) error {
	for _, datafile := range scope.Datafiles {
		messages, err := datafile.Validate()

		if err != nil {
			log.Fatal(err)
		}

		for _, m := range messages {
			log.Println(m)
		}
	}

	saveScope()

	return nil
}

func convert(c *cli.Context) error {
	for _, datafile := range scope.Datafiles {
		if err := datafile.Export(); err != nil {
			log.Fatal(err)
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

func logEntity(e interface{}) {
	dump, _ := json.MarshalIndent(e, "", "  ")

	log.Println(string(dump))
}
