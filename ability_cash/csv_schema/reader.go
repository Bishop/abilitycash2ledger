package csv_schema

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
)

func ReadDatabase(fileName string) (schema.Database, error) {
	db := NewDatabase()

	err := readCsv(fileName, "rates.csv", db.AddRate)
	if err != nil {
		return nil, err
	}

	err = readCsv(fileName, "structure.csv", db.AddAccountMap)
	if err != nil {
		return nil, err
	}

	err = readCsv(fileName, "accounts.csv", db.AddAccount)
	if err != nil {
		return nil, err
	}

	err = readCsv(fileName, "categories.csv", db.AddCategory)
	if err != nil {
		return nil, err
	}

	err = readCsv(fileName, "txs.csv", db.AddTx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readCsv(dirName string, fileName string, handler func([]string)) error {
	file, err := os.Open(filepath.Join(dirName, fileName))

	if err != nil {
		return err
	}

	defer file.Close()

	reader := csv.NewReader(file)
	_, _ = reader.Read() // skip header

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		handler(record)
	}

	return nil
}
