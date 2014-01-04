package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const base uint32 = 137

func hashStr(p string) (hashNum uint32, power uint32) {
	hashNum = 0
	power = 1
	for i := 0; i < len(p); i++ {
		hashNum = hashNum*base + uint32(p[i])
	}
	sq := base
	for i := len(p); i > 0; i >>= 1 {
		if i&1 != 0 {
			power *= sq
		}
		sq *= sq
	}
	return hashNum, power
}

func Index(t, p string) int {
	n := len(p)
	switch {
	case n == 0:
		return 0
	case n == 1:
		return strings.IndexByte(t, p[0])
	case n == len(t):
		if t == p {
			return 0
		}
		return 1
	case n > len(t):
		return -1
	}

	hashNum, power := hashStr(p)
	var hashT uint32 = 0
	for i := 0; i < len(p); i++ {
		hashT = hashT*base + uint32(t[i])
	}
	for i := n; i <= len(t); i++ {
		if hashT == hashNum && t[i-n:i] == p {
			return i - n
		}
		if i < len(t) {
			hashT = uint32(t[i]) + base*hashT - power*uint32(t[i-n])
		}
	}
	return -1
}

func Count(s, sep string) int {
	n := len(sep)
	switch {
	case n == 0:
		return utf8.RuneCountInString(s) + 1
	case n == 1:
		c := sep[0]
		count := 0
		for i := 0; i < len(s); i++ {
			if s[i] == c {
				count++
			}
		}
		return count
	case n == len(s):
		if s == sep {
			return 1
		}
		return 0
	case n > len(s):
		return 0
	}

	count := 0
	hashNum, power := hashStr(sep)
	var hash uint32 = 0
	for i := 0; i < len(sep); i++ {
		hash = hash*base + uint32(s[i])
	}

	for i := n; i < len(s); i++ {
		if hash == hashNum && s[i-n:i] == sep {
			count++
		}
		if i < len(s) {
			hash = uint32(s[i]) + base*hash - power*uint32(sep[i-n])
		}
	}
	return count
}

func Split(s, sep string) []string {
	if len(sep) == 0 {
		return nil
	}

	c := sep[0]
	start := 0

	var results []string
	for i := 0; i <= len(s)-len(sep); i++ {
		if c == s[i] && (len(sep) == 1 || s[i:i+len(sep)] == sep) {
			results = append(results, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	if start != len(s) {
		results = append(results, s[start:])
	}
	return results
}

func FiledsFunc(s string, f func(rune) bool) []string {
	n := 0
	inField := false
	for _, r := range s {
		wasInField := inField
		inField = !f(r)
		if inField && !wasInField {
			n++
		}
	}

	a := make([]string, n)
	na := 0
	start := -1
	for i, r := range s {
		if f(r) {
			if start >= 0 {
				a[na] = s[start:i]
				na++
				start = -1
			}
		} else if start == -1 {
			start = i
		}
	}

	if start >= 0 {
		a[na] = s[start:]
	}
	return a
}

func main() {
	s := "abd def:"
	sep := " "
	fmt.Println(Split(s, sep))
	fmt.Println(FiledsFunc(s, unicode.IsSpace))
}
