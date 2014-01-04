package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const noteSep = ":"

type TxtMarshaler struct{}

type writeFunc func(string, ...interface{}) error

func (write writeFunc) writeInvoice(invoice *Invoice) error {
	note := ""
	if invoice.Note != "" {
		note = noteSep + " " + invoice.Note
	}
	if err := write("INVOICE ID=%d CUSTOMER=%d RAISED=%s DUE=%s "+
		"PAID=%t%s\n", invoice.Id, invoice.CustomerId,
		invoice.Raised.Format(dateFormat),
		invoice.Due.Format(dateFormat), invoice.Paid, note); err != nil {
		return err
	}
	if err := write.writeItems(invoice.Items); err != nil {
		return err
	}
	return write("\f\n")
}

func (write writeFunc) writeItems(items []*Item) error {
	for _, item := range items {
		note := ""
		if item.Note != "" {
			note = noteSep + " " + item.Note
		}
		if err := write("ITEM ID=%s PRICE=%.2f QUANTITY=%d%s\n", item.Id,
			item.Price, item.Quantity, note); err != nil {
			return err
		}
	}
	return nil
}

func (TxtMarshaler) MarshalInvoices(writer io.Writer, invoices []*Invoice) error {
	bufferedWriter := bufio.NewWriter(writer)
	defer bufferedWriter.Flush()

	var write writeFunc = func(format string, args ...interface{}) error {
		_, err := fmt.Fprintf(bufferedWriter, format, args)
		return err
	}

	if err := write("%s %d\n", fileType, fileVersion); err != nil {
		return err
	}

	for _, invoice := range invoices {
		if err := write.writeInvoice(invoice); err != nil {
			return err
		}
	}
	return nil
}

func (TxtMarshaler) UnmarshalInvoices(reader io.Reader) (invoices []*Invoice, err error) {
	bufferedReader := bufio.NewReader(reader)
	if err = checkTxtVersion(bufferedReader); err != nil {
		return nil, err
	}
	eof := false
	for lino := 2; !eof; lino++ {
		line, err := bufferedReader.ReadString('\n')
		if err == io.EOF {
			eof = true
			err = nil
		} else if err != nil {
			return nil, err
		}
		if err = parseTxtLine(lino, line, invoices); err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

func checkTxtVersion(reader *bufio.Reader) error {
	var version int
	if _, err := fmt.Fscanf(reader, "INVOICES %d\n", &version); err != nil {
		return errors.New("can't read non-invoices text file")
	} else if version > fileVersion {
		return fmt.Errorf("version %d is too new to read", version)
	}
	return nil
}

func parseTxtLine(lino int, line string, invoices []*Invoice) (err error) {
	if strings.HasPrefix(line, "INVOICE") {
		var invoice *Invoice
		invoice, err = parseTxtInvoice(lino, line)
		if err != nil {
			return err
		}
		invoices = append(invoices, invoice)
	} else if strings.HasPrefix(line, "ITEM") {
		var item *Item
		item, err = parseTxtItem(lino, line)
		items := &invoices[len(invoices)-1].Items
		*items = append(*items, item)
	}
	return err
}

func parseTxtInvoice(lino int, line string) (invoice *Invoice, err error) {
	invoice = &Invoice{}
	var raised, due string
	if _, err = fmt.Sscanf(line, "INVOICE ID=%d CUSTOMER=%d"+
		"RASIED=%s DUE=%s PAID=%t", &invoice.Id, &invoice.CustomerId, &raised,
		&due, &invoice.Paid); err != nil {
		return nil, fmt.Errorf("invalid invoice %v line %d", err, lino)
	}
	if invoice.Raised, err = time.Parse(dateFormat, raised); err != nil {
		return nil, fmt.Errorf("invalid raised %v line %d", err, lino)
	}
	if invoice.Due, err = time.Parse(dateFormat, due); err != nil {
		return nil, fmt.Errorf("invalid raised %v line %d", err, lino)
	}
	if i := strings.Index(line, noteSep); i > -1 {
		invoice.Note = line[i+len(noteSep):]
	}
	return invoice, nil
}

func parseTxtItem(lino int, line string) (item *Item, err error) {
	item = &Item{}
	if _, err = fmt.Sscanf(line, "ITEM ID=%s PRICE=%f QUANTITY=%d",
		&item.Id, &item.Price, &item.Quantity); err != nil {
		return nil, fmt.Errorf("invalid item %v line %d", err, lino)
	}
	if i := strings.Index(line, noteSep); i > -1 {
		item.Note = strings.TrimSpace(line[i+len(noteSep):])
	}
	return item, nil
}
