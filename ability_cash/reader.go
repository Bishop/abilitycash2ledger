package ability_cash

import (
	"encoding/xml"
	"io/ioutil"

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
