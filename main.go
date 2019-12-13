package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"

	"github.com/Bishop/abilitycash2ledger/abilitycash"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:    "abilitycash db to ledger converter",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "datafile",
				Aliases:  []string{"f"},
				Usage:    "Load transactions from `FILE`",
				Required: true,
			},
		},
		Action: process,
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func process(c *cli.Context) error {
	path := c.String("datafile")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("File %v does not exist\n", path)
	}

	data, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	db := abilitycash.Database{}

	if err = xml.Unmarshal(data, &db); err != nil {
		return err
	}

	log.Printf("%+v", db.Accounts)

	return nil
}
