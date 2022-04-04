package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/Bishop/abilitycash2ledger/scope"
)

const scopeFile = "./scope.json"

var config = scope.NewScope()

func main() {
	readScope()

	app := cli.App{
		Name:    "abilitycash2ledger",
		Usage:   "abilitycash db to ledger converter",
		Version: "0.0.4",
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

	newPath := c.Args().First()

	ensureFileExist(newPath)

	if err := config.AddFile(newPath); err == nil {
		saveScope()
		return nil
	} else {
		return err
	}
}

func prepare(c *cli.Context) error {
	messages, err := config.Validate()

	if err != nil {
		return err
	}

	for _, m := range messages {
		log.Println(m)
	}

	saveScope()

	return nil
}

func convert(c *cli.Context) error {
	return config.Export()
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

	readConfig(scopeFile, &config)
}

func saveScope() {
	if err := saveConfig(scopeFile, &config); err != nil {
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
