package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	START  = "START"
	END    = "END"
	COMMIT = "COMMIT"
	CKPT   = "START CKPT"
	ABORT  = "ABORT"
)

func main() {
	if len(os.Args) != 3 {
		log.Printf("usage %s <logtype>(undo,redo,redo-undo) <logfile>\n, ", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	file, err := os.OpenFile(os.Args[2], os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Open log file %v error, %v\n", os.Args[1], err)
	}
	switch strings.ToLower(os.Args[1]) {
	case "undo":
		RecoverFromUndoLog(file)
	case "redo":
		RecoverFromRedoLog(file)
	case "redo-undo":
	case "undo-redo":
		RecoverFromRedoUndoLog(file)
	default:
		log.Fatal("Unsupported log type")
	}
}
