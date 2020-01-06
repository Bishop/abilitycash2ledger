package xml_schema

import (
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

func (a *acTime) Source() time.Time {
	return a.t
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

func (a *acDate) Format(layout string) string {
	return a.d.Format(layout)
}

func (a *acDate) Source() time.Time {
	return a.d
}

func (tx *Transaction) Item() *txItem {
	switch {
	case tx.Transfer != nil:
		return &tx.Transfer.txItem
	case tx.Expense != nil:
		return &tx.Expense.txItem
	case tx.Income != nil:
		return &tx.Income.txItem
	case tx.Balance != nil:
		return &tx.Balance.txItem
	}

	return nil
}

func (tx *Transaction) IsExecuted() bool {
	return tx.Item().Executed != nil
}

func (tx *Transaction) IsLocked() bool {
	return tx.Item().Locked != nil
}
