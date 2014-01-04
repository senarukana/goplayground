package buffer

import (
	"io"
)

type Buffer struct {
	buf    []byte
	idx    int
	length int
	rd     io.Reader
}

const defaultBufferSize = 4096

func (b *Buffer) fill(need int) error {
	if b.idx > 0 && b.length > 0 {
		copy(b.buf[0:b.length], b.buf[b.idx:])
	}
	if need > len(b.buf) {
		newBuffer := make([]byte, defaultBufferSize*((need-1)/defaultBufferSize+1))
		copy(newBuffer, b.buf[0:b.length])
		b.buf = newBuffer
	}

	b.idx = 0
	for {
		n, err := b.rd.Read(b.buf[b.idx:])
		b.length += n
		if err != nil {
			if err == io.EOF && b.length > need {
				return nil
			}
			return err
		}
		if b.length >= need {
			return nil
		}
	}
}

func (b *Buffer) ReadNext(n int) (data []byte, err error) {
	if b.length < n {
		err = b.fill(n)
		if err != nil {
			return nil, err
		}
	}
	data = b.buf[b.idx : b.idx+n]
	b.idx += n
	b.length -= n
	return data, nil
}

func (b *Buffer) Peek(n int) (data []byte, err error) {
	if b.length < n {
		err = b.fill(n)
		if err != nil {
			return nil, err
		}
	}
	data = b.buf[b.idx : b.idx+n]
	return data, nil
}
