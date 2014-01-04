package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var workers = runtime.NumCPU()

var imageTagFormatter = "<img src=\"%s\" width=\"%d\" height=\"%d\" />"

func main() {
	runtime.GOMAXPROCS(workers)
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage: %s <file.log>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	imagesChan := make(chan string, workers*16)
	done := make(chan struct{}, workers)
	resultsChan := make(chan string)

	go dispatch(imagesChan, os.Args[1:])
	for i := 0; i < workers; i++ {
		go imageProcessor(imagesChan, done, resultsChan)
	}
	waitAndProcessResult(resultsChan, done)
}

func dispatch(imagesChan chan string, imageNames []string) {
	for _, name := range imageNames {
		imagesChan <- name
	}
	close(imagesChan)
}

func imageProcessor(imagesChan <-chan string, done chan<- struct{}, resultsChan chan<- string) {
	for name := range imagesChan {
		if tag, ok := imageReader(name); ok {
			resultsChan <- tag
		}
	}
	done <- struct{}{}
}

func imageReader(name string) (string, bool) {
	file, err := os.Open(name)
	if err != nil {
		log.Println("Open image %v error, %v", name, err)
		return "", false
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Println("Decode image %v config error, %v", name, err)
		return "", false
	}
	return fmt.Sprintf(imageTagFormatter, filepath.Base(name), config.Width, config.Height), true
}

func waitAndProcessResult(resultsChan <-chan string, done <-chan struct{}) {
	for workers > 0 {
		select {
		case result := <-resultsChan:
			fmt.Println(result)
		case _ = <-done:
			workers--
		}
	}
DONE:
	for {
		select {
		case result := <-resultsChan:
			fmt.Println(result)
		default:
			break DONE
		}
	}
}
