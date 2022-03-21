package xml_schema

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/Bishop/abilitycash2ledger/ability_cash/schema"
)

func ReadDatabase(fileName string) (schema.Database, error) {
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	db := new(Database)

	if err = xml.Unmarshal(data, db); err != nil {
		return nil, err
	}

	return db, nil
}
