package main

import (
	"io"
)

const defaultBufferSize = 4 << 10

type ReverseByteReader struct {
	reader          io.ReaderAt
	unread          bool
	length          int16
	index           int16
	buffer          []byte
	currentPosition int64
}

func newReverseByteReader(readerAt io.ReaderAt, length int64) *ReverseByteReader {
	rb := &ReverseByteReader{}
	rb.reader = readerAt
	rb.currentPosition = length
	rb.index = -1
	if length > defaultBufferSize {
		rb.buffer = make([]byte, defaultBufferSize)
	} else {
		rb.buffer = make([]byte, length)
	}
	return rb
}

func (rb *ReverseByteReader) ReadByte() (c byte, err error) {
	if rb.index < 0 {
		if rb.currentPosition == 0 {
			return 0, io.EOF
		}
		bufferSize := int64(defaultBufferSize)
		if rb.currentPosition < bufferSize {
			bufferSize = rb.currentPosition
		}
		rb.currentPosition -= bufferSize
		rb.length = int16(bufferSize)
		rb.index = rb.length - 1
		_, err = rb.reader.ReadAt(rb.buffer[0:bufferSize], rb.currentPosition)
		if err != nil {
			return 0, err
		}
	}
	c = rb.buffer[rb.index]
	rb.index--
	rb.unread = true
	return c, nil
}

func (rl *ReverseByteReader) ReadString(delim byte) (line string, err error) {
	var readData, data []byte
	var c byte
	for {
		c, err = rl.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if c == delim {
			break
		}
		readData = append(readData, c)
	}
	for i := len(readData) - 1; i >= 0; i-- {
		data = append(data, readData[i])
	}
	return string(data), err
}
