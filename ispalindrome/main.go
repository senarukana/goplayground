package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("usage: %s [-a|--ascii] word1 [word2 [... wordN]]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	words := os.Args[1:]
	for _, word := range words {
		fmt.Printf("%5t %q\n", IsPalindrome(word), word)
	}
}

func init() {
	if len(os.Args) > 1 &&
		(os.Args[1] == "-a" || os.Args[1] == "--ascii") {
		os.Args = append(os.Args[:1], os.Args[2:]...)
		IsPalindrome = IsPalindromeASCii
	} else {
		IsPalindrome = IsPalindromeUTF8
	}
}

var IsPalindrome func(string) bool

func IsPalindromeUTF8(word string) bool {
	for len(word) > 0 {
		first, sizeOfFirst := utf8.DecodeRuneInString(word)
		if sizeOfFirst == len(word) {
			break
		}
		last, sizeOfLast := utf8.DecodeLastRuneInString(word)
		if first != last {
			return false
		}
		word = word[sizeOfFirst : len(word)-sizeOfLast]
	}
	return true
}

func IsPalindromeASCii(word string) bool {
	if len(word) <= 1 {
		return true
	}
	j := len(word) - 1
	for i := 0; i < len(word)/2; i++ {
		if word[i] != word[j] {
			return false
		}
		j--
	}
	return true
}
