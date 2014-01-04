package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

const maxGoroutines = 100

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage %s <path>\n, ", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	infoChan := make(chan fileInfo, maxGoroutines*2)
	go findDuplicateFiles(infoChan, os.Args[1])
	pathsInfo := mergeResults(infoChan)
	outputResults(pathsInfo)

}

type fileInfo struct {
	name string
	size int64
	hash []byte
}

func findDuplicateFiles(infoChan chan fileInfo, dirName string) {
	waiter := &sync.WaitGroup{}
	filepath.Walk(dirName, makeWalkFunc(infoChan, waiter))
	waiter.Wait()
	close(infoChan)
}

const maxSizeOfSmallFile = 1024 * 32

func makeWalkFunc(infoChan chan fileInfo, waiter *sync.WaitGroup) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Size() > 0 && (info.Mode()&os.ModeType == 0) {
			if info.Size() < maxSizeOfSmallFile || runtime.NumGoroutine() > maxGoroutines {
				processFile(infoChan, path, info, nil)
			} else {
				waiter.Add(1)
				go processFile(infoChan, path, info, func() { waiter.Done() })
			}
		}
		return nil
	}
}

func processFile(infoChan chan<- fileInfo, fileName string, info os.FileInfo, done func()) {
	if done != nil {
		defer done()
	}
	file, err := os.Open(fileName)
	if err != nil {
		log.Println("error open file %s, %v", fileName, err)
		return
	}
	defer file.Close()
	hash := sha1.New()
	if size, err := io.Copy(hash, file); size != info.Size() || err != nil {
		if err != nil {
			log.Println("error read file %s, %v", fileName, err)
			return
		} else {
			log.Println("failed to read the whole file %s", fileName)
		}
	}
	infoChan <- fileInfo{fileName, info.Size(), hash.Sum(nil)}
}

type pathsInfo struct {
	size  int64
	paths []string
}

func mergeResults(infoChan <-chan fileInfo) map[string]*pathsInfo {
	pathData := make(map[string]*pathsInfo)
	format := fmt.Sprintf("%%016X:%%%dX", sha1.Size*2)
	for info := range infoChan {
		key := fmt.Sprintf(format, info.size, info.hash)
		value, found := pathData[key]
		if !found {
			value = &pathsInfo{size: info.size}
			pathData[key] = value
		}
		value.paths = append(value.paths, info.name)
	}
	return pathData
}

func outputResults(pathsData map[string]*pathsInfo) {
	keys := make([]string, 0, len(pathsData))
	for k := range pathsData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		data := pathsData[key]
		if len(data.paths) > 1 {
			fmt.Printf("%d duplicate files (%s bytes):\n", len(data.paths), commas(data.size))
			for _, name := range data.paths {
				fmt.Printf("\t%s\n", name)
			}
		}
	}
}

func commas(x int64) string {
	value := fmt.Sprint(x)
	for i := len(value) - 3; i > 0; i -= 3 {
		value = value[:i] + "," + value[i:]
	}
	return value
}
