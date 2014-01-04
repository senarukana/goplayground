package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type UndoRedoRecoverLog struct {
	commitedTransactions  map[int]bool
	abortTransactions     map[int]bool
	redoItems             map[int]map[string]Value
	remainingTransactions map[int]bool
	status                int // -1 find nothing, 0: find out end ckpt first 1: find out start ckpt first
}

func RecoverFromRedoUndoLog(file *os.File) {
	defer file.Close()
	fileStat, _ := file.Stat()
	rb := newReverseByteReader(file, fileStat.Size())
	rl := newUndoRedoRecoverLog()
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
	_, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Fatalf("seek log to end error, %v", err)
	}
	for tid, tdata := range rl.redoItems {
		log.Printf("Redo transaction [%d]\n", tid)
		for k, v := range tdata {
			log.Printf("\t %s = %s\n", k, v.newValue)
		}
	}

	for tid, _ := range rl.abortTransactions {
		_, err = file.WriteString(fmt.Sprintf("ABORT %d\n", tid))
		log.Printf("ABORT %d\n", tid)
		if err != nil {
			log.Fatalf("write abort error, %v", err)
		}
	}
	log.Println("Recover complete")
}

type Value struct {
	oldValue string
	newValue string
}

func newUndoRedoRecoverLog() *UndoRedoRecoverLog {
	rl := new(UndoRedoRecoverLog)
	rl.commitedTransactions = make(map[int]bool)
	rl.abortTransactions = make(map[int]bool)
	rl.redoItems = make(map[int]map[string]Value)
	rl.remainingTransactions = make(map[int]bool)
	rl.status = -1
	return rl
}

func (rl *UndoRedoRecoverLog) lineAnalyzer(line string) bool {
	switch {
	case strings.HasPrefix(line, END):
		if rl.status != -1 {
			log.Fatalf("%v, redo-undo log is disruptted", line)
		}
		rl.status = 0
		log.Print("Find the END CKPT, change the status from -1 to 0")
		return false
	case strings.HasPrefix(line, CKPT):
		if rl.status == -1 {
			rl.status = 1
			strs := strings.Split(line, " ")
			if len(strs) != 3 {
				log.Fatalf("%v, redo-undo log is disruppted", line)
			}
			transactions := strings.Split(strs[2], ",")
			for i := 0; i < len(transactions); i++ {
				tid, err := strconv.Atoi(transactions[i])
				if err != nil {
					log.Fatal("%v, invalid tid, redo-undo log is disruptted", line)
				}
				rl.remainingTransactions[tid] = true
			}
			log.Print("Find the START CKPT, change the status from -1 to 1")
		} else if rl.status == 0 {
			return true // recover complete
		} else {
			log.Fatal("redo-undo log is disruptted")
		}
		return false
	case strings.HasPrefix(line, COMMIT):
		strs := strings.Split(line, " ")
		if len(strs) != 2 {
			log.Fatalf("%v, invalid commit format, redo-undo log is disruptted", line)
		}
		tid, err := strconv.Atoi(strs[1])
		if err != nil {
			log.Fatal("%v, invalid tid, redo-undo log is disruptted", line)
		}
		if _, ok := rl.commitedTransactions[tid]; !ok {
			rl.commitedTransactions[tid] = true
		} else {
			log.Fatal("redo-undo log is disruptted")
		}
		return false
	case strings.HasPrefix(line, START):
		if rl.status == 1 {
			strs := strings.Split(line, " ")
			if len(strs) != 2 {
				log.Fatalf("%v, redo-undo log is disruppted", line)
			}
			tid, err := strconv.Atoi(strs[1])
			if err != nil {
				log.Fatalf("%v, redo-undo log is disruppted", line)
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
		if len(strs) != 4 {
			log.Fatalf("%v, redo-undo log is disruppted", line)
		}
		tid, err := strconv.Atoi(strs[0])
		if err != nil {
			log.Fatalf("%v, redo-undo log is disruppted", line)
		}
		if _, ok := rl.commitedTransactions[tid]; ok {
			if rl.redoItems[tid] == nil {
				rl.redoItems[tid] = make(map[string]Value)
			}
			rl.redoItems[tid][strs[1]] = Value{strs[2], strs[3]}
		} else {
			rl.abortTransactions[tid] = true
		}
		return false
	}
	panic("never go here")
}
