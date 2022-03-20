package ability_cash

import (
	"encoding/csv"
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"

	"github.com/Bishop/abilitycash2ledger/ability_cash/csv_schema"
	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
	"github.com/Bishop/abilitycash2ledger/ability_cash/xml_schema"
)

func ReadXmlDatabase(fileName string) (schema.Database, error) {
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	db := new(xml_schema.Database)

	if err = xml.Unmarshal(data, db); err != nil {
		return nil, err
	}

	return db, nil
}

func ReadCsvDatabase(fileName string) (schema.Database, error) {
	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := csv.NewReader(file)

	db := new(csv_schema.Database)

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if record[0] == "Executed" {
			continue
		}

		db.Fill(record)
	}

	return db, nil
}
