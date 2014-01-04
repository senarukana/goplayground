package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

const binaryDateForamt = "20060102"

var byteOrder = binary.LittleEndian

type binaryWriteFunc func(interface{}) error

func (write binaryWriteFunc) writeInvoice(invoice *Invoice) error {
	for _, i := range []int{invoice.Id, invoice.CustomerId} {
		if err := write(int32(i)); err != nil {
			return err
		}
	}
	for _, date := range []time.Time{invoice.Raised, invoice.Due} {
		if err := write.writeDate(date); err != nil {
			return err
		}
	}
	if err := write.writeBool(invoice.Paid); err != nil {
		return err
	}
	if err := write.writeString(invoice.Note); err != nil {
		return err
	}

	for _, item := range invoice.Items {
		if err := write.writeItem(item); err != nil {
			return err
		}
	}
	return nil
}

func (write binaryWriteFunc) writeItem(item *Item) error {
	if err := write.writeString(item.Id); err != nil {
		return err
	}

	if err := write(item.Price); err != nil {
		return err
	}

	if err := write(int16(item.Quantity)); err != nil {
		return err
	}

	return write.writeString(item.Note)
}

func (write binaryWriteFunc) writeDate(date time.Time) error {
	i, err := strconv.Atoi(date.Format(binaryDateForamt))
	if err != nil {
		return err
	}
	return write(int32(i))
}

func (write binaryWriteFunc) writeBool(b bool) error {
	var v int8
	if b {
		v = 1
	}
	return write(v)
}

func (write binaryWriteFunc) writeString(str string) error {
	if err := write(int32(len(str))); err != nil {
		return err
	}
	return write([]byte(str))
}

type BinaryMarshaler struct{}

func (BinaryMarshaler) MarshalInvoices(writer io.Writer, invoices []*Invoice) error {
	var write binaryWriteFunc = func(x interface{}) error {
		return binary.Write(writer, byteOrder, x)
	}

	if err := write(uint32(magicNumber)); err != nil {
		return err
	}

	if err := write(uint16(fileVersion)); err != nil {
		return err
	}

	if err := write(int32(len(invoices))); err != nil {
		return err
	}

	for _, invoice := range invoices {
		if err := write.writeInvoice(invoice); err != nil {
			return err
		}
	}
	return nil
}

func (BinaryMarshaler) UnmarshalInvoices(reader io.Reader) (invoices []*Invoice, err error) {
	if err := checkInvoiceVersion(reader); err != nil {
		return nil, err
	}
	var count int
	count, err = readIntFromInt32(reader)
	if err != nil {
		return nil, err
	}
	invoices = make([]*Invoice, 0, count)
	for i := 0; i < count; i++ {
		invoice, err := readInvoice(reader)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil

}

func checkInvoiceVersion(reader io.Reader) error {
	var magic uint32
	if err := binary.Read(reader, byteOrder, &magic); err != nil {
		return err
	}
	if magic != magicNumber {
		return errors.New("can't read non-invoices inv file")
	}

	var version uint16
	if err := binary.Read(reader, byteOrder, &version); err != nil {
		return err
	}
	if version > fileVersion {
		return fmt.Errorf("version %d is too new to read", version)
	}
	return nil
}

func readIntFromInt32(reader io.Reader) (int, error) {
	var i32 int32
	err := binary.Read(reader, byteOrder, &i32)
	return int(i32), err
}

func readIntFromInt16(reader io.Reader) (int, error) {
	var i16 int16
	err := binary.Read(reader, byteOrder, &i16)
	return int(i16), err
}

func readBoolFromInt8(reader io.Reader) (bool, error) {
	var i8 int8
	err := binary.Read(reader, byteOrder, &i8)
	return i8 == 1, err
}

func readInvDate(reader io.Reader) (time.Time, error) {
	var n int32
	if err := binary.Read(reader, byteOrder, &n); err != nil {
		return time.Time{}, err
	}
	return time.Parse(binaryDateForamt, fmt.Sprint(n))
}

func readString(reader io.Reader) (str string, err error) {
	var strlen int32
	if err = binary.Read(reader, byteOrder, &strlen); err != nil {
		return "", err
	}

	raw := make([]byte, strlen)
	if err = binary.Read(reader, byteOrder, raw); err != nil {
		return "", err
	}
	return string(raw), nil
}

func readInvoice(reader io.Reader) (invoice *Invoice, err error) {
	invoice = &Invoice{}
	for _, pid := range []*int{&invoice.Id, &invoice.CustomerId} {
		if *pid, err = readIntFromInt32(reader); err != nil {
			return nil, err
		}
	}

	for _, date := range []*time.Time{&invoice.Raised, &invoice.Due} {
		if *date, err = readInvDate(reader); err != nil {
			return nil, err
		}
	}

	if invoice.Paid, err = readBoolFromInt8(reader); err != nil {
		return nil, err
	}

	if invoice.Note, err = readString(reader); err != nil {
		return nil, err
	}

	var itemCount int
	if itemCount, err = readIntFromInt32(reader); err != nil {
		return nil, err
	}
	invoice.Items, err = readInvItems(reader, itemCount)
	return invoice, err
}

func readInvItems(reader io.Reader, count int) ([]*Item, error) {
	items := make([]*Item, 0, count)
	for i := 0; i < count; i++ {
		item, err := readInvItem(reader)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func readInvItem(reader io.Reader) (item *Item, err error) {
	item = &Item{}
	if item.Id, err = readString(reader); err != nil {
		return nil, err
	}
	if err = binary.Read(reader, byteOrder, &item.Price); err != nil {
		return nil, err
	}
	if item.Quantity, err = readIntFromInt16(reader); err != nil {
		return nil, err
	}
	item.Note, err = readString(reader)
	return item, nil
}
