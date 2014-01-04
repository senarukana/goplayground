package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/senarukana/test/safeslice"
)

var workers = runtime.NumCPU()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage: %s <file.log>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	lines := make(chan string, workers*10)
	done := make(chan struct{}, workers)
	pageSlice := safeSlice.New()
	go readLog(lines, os.Args[1])
	getRx := regexp.MustCompile(`GET [ \t]+[^ \t\n]+[.]html?)`)
	for i := 0; i < workers; i++ {
		go processLog(lines, done, pageSlice, getRx)
	}
	waitUntil(done)
	showResults(pageSlice)
}

func readLog(lineChan chan<- string, logName string) {
	file, err := os.Open(logName)
	if err != nil {
		log.Fatalf("Read file %v error, %v", logName, err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			lineChan <- line
		}
		if err != nil {
			if err != io.EOF {
				log.Println("Failed to finish reading file:", err)
			}
			break
		}
	}
	close(lineChan)
}

func processLog(lineChan <-chan string, doneChan chan<- struct{},
	pageSlice safeSlice.SafeSlice, getRx *regexp.Regexp) {
	for line := range lineChan {
		if matches := getRx.FindStringSubmatch(line); matches != nil {
			pageSlice.Append(matches[1])
		}
	}
	doneChan <- struct{}{}
}

func waitUntil(donechan <-chan struct{}) {
	for i := 0; i < workers; i++ {
		<-donechan
	}
}

func showResults(pageList safeSlice.SafeSlice) {
	list := pageList.Close()
	counts := make(map[string]int)
	for _, page := range list { // uniquify
		counts[page.(string)] += 1
	}
	for page, count := range counts {
		fmt.Printf("%8d %s\n", count, page)
	}
}
