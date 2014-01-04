package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type RedoRecoverLog struct {
	commitedTransactions  map[int]bool
	recoverItems          map[int]map[string]string
	remainingTransactions map[int]bool
	abortTransactions     map[int]bool
	status                int // -1 find nothing, 0: find out end ckpt first 1: find out start ckpt first
}

func RecoverFromRedoLog(file *os.File) {
	defer file.Close()
	fileStat, _ := file.Stat()
	rb := newReverseByteReader(file, fileStat.Size())
	rl := newRecoverRedoLog()
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
		log.Printf("Redo transaction [%d]\n", tid)
		for k, v := range tdata {
			log.Printf("\t %s = %s\n", k, v)
		}
	}
	_, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Fatalf("seek log to end error, %v", err)
	}
	for tid, _ := range rl.abortTransactions {
		_, err = file.WriteString(fmt.Sprintf("ABORT %d\n", tid))
		if err != nil {
			log.Fatalf("write abort error, %v", err)
		}
	}
	log.Println("Recover complete")
}

func newRecoverRedoLog() *RedoRecoverLog {
	rl := new(RedoRecoverLog)
	rl.commitedTransactions = make(map[int]bool)
	rl.abortTransactions = make(map[int]bool)
	rl.recoverItems = make(map[int]map[string]string)
	rl.remainingTransactions = make(map[int]bool)
	rl.status = -1
	return rl
}

func (rl *RedoRecoverLog) lineAnalyzer(line string) bool {
	switch {
	case strings.HasPrefix(line, END):
		rl.status = 0
		log.Println("Find the END CKPT, change the status from -1 to 0, need to find")
		return false
	case strings.HasPrefix(line, CKPT):
		if rl.status == -1 {
			log.Println("find the START CKPT, but the status is -1, need to find END CKPT")
		} else if rl.status == 0 {
			rl.status = 1
			strs := strings.Split(line, " ")
			if len(strs) != 3 {
				log.Fatalf("%v, redo log is disruppted", line)
			}
			transactions := strings.Split(strs[2], ",")
			for i := 0; i < len(transactions); i++ {
				tid, err := strconv.Atoi(transactions[i])
				if err != nil {
					log.Fatal("%v, invalid tid, redo log is disruptted", line)
				}
				if _, ok := rl.commitedTransactions[tid]; ok {
					rl.remainingTransactions[tid] = true
				}
			}
			log.Println("Find the START CKPT, change the status from 0 to 1")
		}
		return false
	case strings.HasPrefix(line, COMMIT):
		strs := strings.Split(line, " ")
		if len(strs) != 2 {
			log.Fatalf("%v, invalid commit format, redo log is disruptted", line)
		}
		tid, err := strconv.Atoi(strs[1])
		if err != nil {
			log.Fatal("%v, invalid tid, redo log is disruptted", line)
		}
		if _, ok := rl.commitedTransactions[tid]; !ok {
			rl.commitedTransactions[tid] = true
		} else {
			log.Fatalf("%v, redo log is disruptted", line)
		}
		return false
	case strings.HasPrefix(line, x, START):
		if rl.status == 1 {
			strs := strings.Split(line, " ")
			if len(strs) != 2 {
				log.Fatalf("%v, redo log is disruppted", line)
			}
			tid, err := strconv.Atoi(strs[1])
			if err != nil {
				log.Fatalf("%v, redo log is disruppted", line)
			}
			if _, ok := rl.remainingTransactions[tid]; ok {
				delete(rl.remainingTransactions, tid)
				if len(rl.remainingTransactions) == 0 {
					return true
				}
			}
		}
		return false
	case strings.HasPrefix(line, ABORT):

	default:
		var tid int
		strs := strings.Split(line, ",")
		if len(strs) != 3 {
			log.Fatalf("%v, redo log is disruppted", line)
		}
		tid, err := strconv.Atoi(strs[0])
		if err != nil {
			log.Fatalf("%v, redo log is disruppted", line)
		}
		if _, ok := rl.commitedTransactions[tid]; ok {
			if rl.recoverItems[tid] == nil {
				rl.recoverItems[tid] = make(map[string]string)
			}
			rl.recoverItems[tid][strs[1]] = strs[2]
		} else {
			rl.abortTransactions[tid] = true
		}
		return false
	}
	panic("never go here")
}
