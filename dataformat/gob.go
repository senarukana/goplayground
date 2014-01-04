package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
)

type GobMarshaler struct{}

func (GobMarshaler) MarshalInvoices(writer io.Writer, invoices []*Invoice) error {
	encoder := gob.NewEncoder(writer)
	if err := encoder.Encode(magicNumber); err != nil {
		return err
	}
	if err := encoder.Encode(fileVersion); err != nil {
		return err
	}
	return encoder.Encode(invoices)
}

func (GobMarshaler) UnmarshalInvoices(reader io.Reader) (invoices []*Invoice, err error) {
	decoder := gob.NewDecoder(reader)
	var magic int
	if err := decoder.Decode(&magic); err != nil {
		return nil, err
	}
	if magic != magicNumber {
		return nil, errors.New("can't read non-invoices gob file")
	}
	var version int
	if err := decoder.Decode(&version); err != nil {
		return nil, err
	}
	if version > fileVersion {
		return nil, fmt.Errorf("version %d is too new to read", version)
	}
	err = decoder.Decode(invoices)
	return invoices, err
}
