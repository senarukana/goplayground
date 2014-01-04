package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	fileType             = "INVOICES"
	magicNumber          = 0x125D
	fileVersion          = 100
	dateFormat           = "2006-02-02"
	nanosecondsToSeconds = 1e9
)

type Invoice struct {
	Id         int
	CustomerId int
	Raised     time.Time
	Due        time.Time
	Paid       bool
	Note       string
	Items      []*Item
}

type Item struct {
	Id       string
	Price    float64
	Quantity int
	Note     string
}

type InvoiceMarshaler interface {
	MarshalInvoices(writer io.Writer, invocies []*Invoice) error
}

type InvoiceUnMarshal interface {
	UnmarshalInvoices(reader io.Reader) ([]*Invoice, error)
}

func main() {
	log.SetFlags(0)
	report := false
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-t" || args[0] == "--time") {
		report = true
		args = args[1:]
	}
	if len(args) != 2 || args[0] == "-h" || args[0] == "--help" {
		log.Fatalf("usage: %s [-t|--time] infile.ext outfile.ext\n"+
			".ext may be any of .gob, .inv, .jsn, .json, .txt,"+
			"or .xml, optionally gzziped(e.g., .gob.gz)\n", filepath.Base(os.Args[0]))
	}
	inFilename, outFilename := args[0], args[1]
	if inFilename == outFilename {
		log.Fatalln("won't overwrite a file with itself")
	}

	start := time.Now()
	invoices, err := readInvoiceFile(inFilename)
	if err == nil && report {
		duration := time.Now().Sub(start)
		fmt.Printf("Read %s in %.3f seconds\n", inFilename,
			float64(duration)/nanosecondsToSeconds)
	}
	if err != nil {
		log.Fatalln("Failed to read:", err)
	}
	err = writeInvoiceFile(outFilename, invoices)
	if err == nil && report {
		duration := time.Now().Sub(start)
		fmt.Printf("Wrote %s in %.3f seconds\n", outFilename,
			float64(duration)/nanosecondsToSeconds)
	}
	if err != nil {
		log.Fatalln("Failed to write:", err)
	}
}

func readInvoiceFile(infileName string) ([]*Invoice, error) {
	file, closer, err := openInvoiceFile(infileName)
	if closer != nil {
		defer closer()
	}
	if err != nil {
		return nil, err
	}
	return readInvoices(file, suffixof(infileName))
}

func readInvoices(reader io.Reader, suffix string) ([]*Invoice, error) {
	var unmarshaler InvoiceUnMarshal
	switch suffix {
	case ".gob":
		unmarshaler = GobMarshaler{}
	case ".binary":
		unmarshaler = BinaryMarshaler{}
	case ".jsn", ".json":
		unmarshaler = JSONMarshaler{}
	case ".txt":
		unmarshaler = TxtMarshaler{}
	}
	if unmarshaler != nil {
		return unmarshaler.UnmarshalInvoices(reader)
	}
	return nil, fmt.Errorf("unrecognized input suffix: %s", suffix)
}

func writeInvoiceFile(filename string, invoices []*Invoice) error {
	writer, closer, err := createInvoiceFile(filename)
	if closer != nil {
		defer closer()
	}
	if err != nil {
		return err
	}
	return writeInvoices(writer, suffixof(filename), invoices)
}

func openInvoiceFile(filename string) (io.ReadCloser, func(), error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	closer := func() { file.Close() }
	var reader io.ReadCloser = file
	var decompressor *gzip.Reader
	if strings.HasSuffix(filename, ".gz") {
		if decompressor, err = gzip.NewReader(file); err != nil {
			return file, closer, err
		}
		closer = func() { decompressor.Close(); file.Close() }
		reader = decompressor
	}
	return reader, closer, nil
}

func createInvoiceFile(filename string) (io.WriteCloser, func(), error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, nil, err
	}
	closer := func() { file.Close() }
	var writer io.WriteCloser = file
	var compressor *gzip.Writer
	if strings.HasSuffix(filename, ".gz") {
		compressor = gzip.NewWriter(file)
		closer = func() { compressor.Close(); file.Close() }
		writer = compressor
	}
	return writer, closer, nil
}

func writeInvoices(writer io.Writer, suffix string,
	invoices []*Invoice) error {
	var marshaler InvoiceMarshaler
	switch suffix {
	case ".gob":
		marshaler = GobMarshaler{}
	case ".inv":
		marshaler = BinaryMarshaler{}
	case ".jsn", ".json":
		marshaler = JSONMarshaler{}
	case ".txt":
		marshaler = TxtMarshaler{}
	}
	if marshaler != nil {
		return marshaler.MarshalInvoices(writer, invoices)
	}
	return errors.New("unrecognized output suffix")
}

func suffixof(filename string) string {
	suffix := filepath.Ext(filename)
	if suffix == ".gz" {
		suffix = filepath.Ext(filename[:len(filename)-3])
	}
	return suffix

}
