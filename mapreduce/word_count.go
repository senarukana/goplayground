package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var workers = runtime.NumCPU()

func main() {
	if len(os.Args) < 3 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage: %s <output-file> <input-file1> <input-file2> ... <input-filen>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	filesChan := make(chan string, workers*10)
	reduceChans := make([]chan string, workers)
	for i := 0; i < workers; i++ {
		reduceChans[i] = make(chan string)
	}
	doneChan := make(chan map[string]int, workers)

	ofile, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Printf("unable to create output file, error %v\n", err)
		os.Exit(1)
	}

	prepare(filesChan, os.Args[2:])
	go readFile(reduceChans, filesChan)
	for i := 0; i < workers; i++ {
		go countWord(reduceChans[i], doneChan)
	}

	bw := bufio.NewWriter(ofile)
	for i := 0; i < workers; i++ {
		result := <-doneChan
		for word, count := range result {
			bw.WriteString(fmt.Sprintf("%s %d\n", word, count))
		}
	}
	bw.Flush()
	ofile.Close()
}

func prepare(filesChan chan<- string, files []string) {
	for _, fileName := range files {
		filesChan <- fileName
	}
	close(filesChan)
}

func simpleHash(word string) int {
	hash := 0
	for i := 0; i < len(word); i++ {
		hash += int(word[i])
	}
	return hash % workers
}

func readFile(reduceChans []chan string, filesChan <-chan string) {
	filedsFunc := func(char rune) bool {
		switch char {
		case '\t', ' ', ',', '.', '(', ')':
			return true
		}
		return false
	}
	for fileName := range filesChan {
		file, err := os.Open(fileName)
		if err != nil {
			log.Printf("unable to read %v file, error %v, ignore it\n", fileName, err)
			continue
		}
		br := bufio.NewReader(file)
		for {
			line, err := br.ReadString('\n')
			if line != "" {
				words := strings.FieldsFunc(line, filedsFunc)
				for _, word := range words {
					index := simpleHash(word)
					reduceChans[index] <- word
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("read file: %v encounters an error %v\n", fileName, err)
				}
				break
			}
		}
		file.Close()
	}
	for i := 0; i < workers; i++ {
		close(reduceChans[i])
	}
}

func countWord(reduceChan <-chan string, doneChan chan<- map[string]int) {
	wordMap := make(map[string]int)
	for word := range reduceChan {
		wordMap[word]++
	}

	doneChan <- wordMap
}
