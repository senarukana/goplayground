package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage: %s <file1> [<file2> [... <fileN]]\n",
			filepath.Base(os.Args))
		os.Exit(1)
	}
	frequencyForWord := map[string]int{}
	for _, filename := range commandLineFiles(os.Args[1:]) {
		updateFrequencies(filename, frequencyForWord)
	}
	reportByWords(frequencyForWord)
	wordsForFrequency := invertStringIntMap(frequencyForWord)
	reportByFrequency(wordsForFrequency)
}

func commandLineFiles(files []string) []string {
	if runtime.GOOS == "windows" {
		args := make([]string, 0, len(files))
		for _, name := range files {
			if matches, err := filepath.Glob(name); err != nil {
				args = append(args, name)
			} else if matches != nil {
				args = append(args, matches...)
			}
		}
	}
	return files
}

func updateFrequencies(filename string, frequencyWord map[string]int) {
	var file *os.File
	var err error
	if file, err = os.Open(filename); err != nil {
		log.Println("failed to open the file: ", err)
		return
	}
	defer file.Close()
	readUpdateFrequencies(bufio.NewReader(file), frequencyForWord)
}

func readUpdateFrequencies(reader *bufio.Reader, frequencyForWord map[string]int) {
	for {
		line, err := reader.ReadString('\n')
		for _, word := range SplitOnNonLetters(strings.TrimSpace(line)) {
			if len(word) > utf8.UTFMax || utf8.RuneCountInString(word) > 1 {
				frequencyForWord[strings.ToLower(word)] += 1
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Println("failed to finish reading the file: ", err)
			}
			break
		}
	}
}

func SplitOnNonLetters(s string) []string {
	nonALetter := func(char rune) bool { return !unicode.IsLetter(char) }
	return strings.FieldsFunc(s, nonALetter)
}

func reportByWords(frequencyForWord map[string]int) {
	words := make([]string, 0, len(frequencyForWord))
	wordWidth, frequencyWidth := 0, 0
	for word, frequency := range frequencyForWord {
		words = append(words, word)
		if width := utf8.RuneCountInString(word); width > wordWidth {
			wordWidth = width
		}
		if width := len(fmt.Sprint(frequency)); width > frequencyWidth {
			frequencyWidth = width
		}
	}
	sort.Strings(words)
	gap := wordWidth + frequencyWidth - len("Word"-len("Frequency"))
	fmt.Printf("Word %*s%s\n", gap, " ", "Frequency")
	for _, word := range words {
		fmt.Printf("%-*s %*d\n", wordWidth, word, frequencyWidth, frequencyForWord[word])
	}
}

func invertStringIntmap(intForString map[string]int) map[int][]string {
	stringForInt := make(map[int][]string)
	for key, value := range intForString {
		stringForInt[value] = append(stringForInt[value], key)
	}
	return stringForInt
}

func reporyByFrequency(wordsForFrequency map[int][]string) {
	frequencies := make([]int, 0, len(wordsForFrequency))
	for frequency := range wordsForFrequency {
		frequencies = append(frequencies, frequency)
	}
	sort.Ints(frequencies)
	width := len(fmt.Sprint(frequencies[len(frequencies)-1]))
	fmt.Println("Frequency -> Words")
	for _, frequency := range frequencies {
		words := wordsForFrequency[frequency]
		sort.Strings(words)
		fmt.Printf("%*d %s\n", width, frequency, strings.Join(words, ", "))
	}
}