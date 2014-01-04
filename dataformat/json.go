package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

type JSONInvoice struct {
	Id         int
	CustomerId int
	Raised     string // time.Time in Invoice struct
	Due        string // time.Time in Invoice struct
	Paid       bool
	Note       string
	Items      []*Item
}

func (invoice Invoice) MarshalJSON() ([]byte, error) {
	jsonInvoice := JSONInvoice{
		invoice.Id,
		invoice.CustomerId,
		invoice.Raised.Format(dateFormat),
		invoice.Due.Format(dateFormat),
		invoice.Paid,
		invoice.Note,
		invoice.Items,
	}
	return json.Marshal(jsonInvoice)
}

func (invoice *Invoice) UnmarshalJSON(data []byte) (err error) {
	var jsonInvoice JSONInvoice
	if err = json.Unmarshal(data, &jsonInvoice); err != nil {
		return err
	}
	var raised, due time.Time
	if raised, err = time.Parse(dateFormat, jsonInvoice.Raised); err != nil {
		return err
	}
	if due, err = time.Parse(dateFormat, jsonInvoice.Due); err != nil {
		return err
	}
	*invoice = Invoice{
		jsonInvoice.Id,
		jsonInvoice.CustomerId,
		raised,
		due,
		jsonInvoice.Paid,
		jsonInvoice.Note,
		jsonInvoice.Items,
	}
	return nil
}

type JSONMarshaler struct{}

func (JSONMarshaler) MarshalInvoices(writer io.Writer, invoices []*Invoice) error {
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(fileType); err != nil {
		return err
	}
	if err := encoder.Encode(fileVersion); err != nil {
		return err
	}
	return encoder.Encode(invoices)
}

func (JSONMarshaler) UnmarshalInvoices(reader io.Reader) ([]*Invoice,
	error) {
	decoder := json.NewDecoder(reader)
	var kind string
	if err := decoder.Decode(&kind); err != nil {
		return nil, err
	}
	if kind != fileType {
		return nil, errors.New("cannot read non-invoices json file")
	}
	var version int
	if err := decoder.Decode(&version); err != nil {
		return nil, err
	}
	if version > fileVersion {
		return nil, fmt.Errorf("version %d is too new to read", version)
	}
	var invoices []*Invoice
	err := decoder.Decode(&invoices)
	return invoices, err
}
