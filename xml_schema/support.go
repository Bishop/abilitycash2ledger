package xml_schema

import (
	"encoding/json"
	"encoding/xml"
	"time"
)

type acTime struct {
	t time.Time
}

func (a *acTime) UnmarshalXMLAttr(attr xml.Attr) error {
	const format = "2006-01-02T15:04:05" // 2011-09-02T20:40:53

	if parse, err := time.ParseInLocation(format, attr.Value, time.Local); err != nil {
		return err
	} else {
		*a = acTime{parse}
	}

	return nil
}

func (a acTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.t)
}

type acDate struct {
	d time.Time
}

func (a *acDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const format = "2006-01-02" // 2011-01-01

	var s string

	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	if parse, err := time.ParseInLocation(format, s, time.Local); err != nil {
		return err
	} else {
		*a = acDate{parse}
	}

	return nil
}

func (a acDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.d)
}

func (a *acDate) Format(layout string) string {
	return a.d.Format(layout)
}

func (a *acDate) Source() time.Time {
	return a.d
}

func (tx *txItem) IsExecuted() bool {
	return tx.Executed != nil
}

func (tx *txItem) IsLocked() bool {
	return tx.Locked != nil
}
