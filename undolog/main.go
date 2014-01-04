package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const defaultBufferSize = 4 << 10

type ReverseByteReader struct {
	reader          io.ReaderAt
	unread          bool
	length          int16
	index           int16
	buffer          [defaultBufferSize]byte
	currentPosition int64
}

func newReverseByteReader(readerAt io.ReaderAt, length int64) *ReverseByteReader {
	rb := &ReverseByteReader{}
	rb.reader = readerAt
	rb.currentPosition = length
	rb.index = -1
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

var (
	START  = "START"
	END    = "END"
	COMMIT = "COMMIT"
	CKPT   = "START CKPT"
)

func main() {
	RecoverFromUndoLog("test1.txt")

}

func RecoverFromUndoLog(logName string) {
	file, err := os.Open(logName)
	if err != nil {
		log.Fatalf("Open log file error, %v", err)
	}
	defer file.Close()
	fileStat, _ := file.Stat()
	rb := newReverseByteReader(file, fileStat.Size())
	rl := newRecoverLog()
	for {
		line, err := rb.ReadString('\n')
		if line != "" {
			if rl.lineAnalyzer(line) {
				log.Print("Recover complete")
				break
			}

		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err.Error())
		}
	}
	for tid, tdata := range rl.recoverItems {
		log.Printf("Recover transaction [%d]\n", tid)
		for k, v := range tdata {
			log.Printf("\t %s = %s\n", k, v)
		}
	}
}

type RecoverLog struct {
	commitTransaction     map[int]bool
	recoverItems          map[int]map[string]string
	remainingTransactions map[int]bool
	status                int // -1 find nothing, 0: find out end ckpt first 1: find out start ckpt first
}

func newRecoverLog() *RecoverLog {
	rl := new(RecoverLog)
	rl.commitTransaction = make(map[int]bool)
	rl.recoverItems = make(map[int]map[string]string)
	rl.remainingTransactions = make(map[int]bool)
	rl.status = -1
	return rl
}

func (rl *RecoverLog) lineAnalyzer(line string) bool {
	switch {
	case strings.HasPrefix(line, END):
		if rl.status != -1 {
			log.Fatalf("%v, undo log is disruptted", line)
		}
		rl.status = 0
		log.Print("Find the END CKPT, change the status from -1 to 0")
		return false
	case strings.HasPrefix(line, CKPT):
		if rl.status == -1 {
			rl.status = 1
			strs := strings.Split(line, " ")
			if len(strs) != 3 {
				log.Fatalf("%v, undo log is disruppted", line)
			}
			transactions := strings.Split(strs[2], ",")
			for i := 0; i < len(transactions); i++ {
				tid, err := strconv.Atoi(transactions[i])
				if err != nil {
					log.Fatal("%v, invalid tid, undo log is disruptted", line)
				}
				rl.remainingTransactions[tid] = true
			}
			log.Print("Find the START CKPT, change the status from -1 to 1")
		} else if rl.status == 0 {
			return true // recover complete
		} else {
			log.Fatal("undo log is disruptted")
		}
		return false
	case strings.HasPrefix(line, COMMIT):
		strs := strings.Split(line, " ")
		if len(strs) != 2 {
			log.Fatalf("%v, invalid commit format, undo log is disruptted", line)
		}
		tid, err := strconv.Atoi(strs[1])
		if err != nil {
			log.Fatal("%v, invalid tid, undo log is disruptted", line)
		}
		if _, ok := rl.commitTransaction[tid]; !ok {
			rl.commitTransaction[tid] = true
		} else {
			log.Fatal("undo log is disruptted")
		}
		return false
	case strings.HasPrefix(line, START):
		if rl.status == 1 {
			strs := strings.Split(line, " ")
			if len(strs) != 2 {
				log.Fatalf("%v, undo log is disruppted", line)
			}
			tid, err := strconv.Atoi(strs[1])
			if err != nil {
				log.Fatalf("%v, undo log is disruppted", line)
			}
			if _, ok := rl.remainingTransactions[tid]; ok {
				delete(rl.remainingTransactions, tid)
				if len(rl.remainingTransactions) == 0 {
					return true
				}
			}
		}
		return false
	default:
		var tid int
		strs := strings.Split(line, ",")
		if len(strs) != 3 {
			log.Fatalf("%v, undo log is disruppted", line)
		}
		tid, err := strconv.Atoi(strs[0])
		if err != nil {
			log.Fatalf("%v, undo log is disruppted", line)
		}
		if _, ok := rl.commitTransaction[tid]; !ok {
			if rl.recoverItems[tid] == nil {
				rl.recoverItems[tid] = make(map[string]string)
			}
			rl.recoverItems[tid][strs[1]] = strs[2]
		}
		return false
	}
	panic("never go here")
}
