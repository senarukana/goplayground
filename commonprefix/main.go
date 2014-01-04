package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

func CommonPrefix(strs []string) string {
	var prefix string
	for _, str := range strs {
		if len(prefix) == 0 || len(prefix) > len(str) {
			prefix = str
		}
	}

	var i int
	for _, str := range strs {
		for i = 0; i < len(str) && i < len(prefix); i++ {
			if prefix[i] == str[i] {
				continue
			}
			break
		}
		if i < len(prefix) {
			prefix = str[0:i]
		}
	}
	return prefix
}

func commonPrefix(strs []string) string {
	components := make([][]rune, len(strs))
	for i, str := range strs {
		components[i] = []rune(str)
	}
	if len(components) == 0 || len(components[0]) == 0 {
		return ""
	}
	var common bytes.Buffer
FINISH:
	for i := 0; i < len(components[0]); i++ {
		c := components[0][i]
		for j := 1; j < len(components); j++ {
			if i > len(components[j]) || components[j][i] != c {
				break FINISH
			}
		}
		common.WriteRune(c)
	}
	return common.String()
}

func commonPathPrefix(paths []string) string {
	const separator = string(filepath.Separator)
	components := make([][]string, len(paths))
	for i, path := range paths {
		components[i] = strings.Split(path, separator)
	}
	var common []string
FINISH:
	for i := 0; i < len(components[0]); i++ {
		part := components[0][i]
		for j := 1; j < len(components); j++ {
			if i > len(components[j]) || components[j][i] != part {
				break FINISH
			}
		}
		common = append(common, part)
	}
	return filepath.Join(common...)
}

func main() {
	strs := []string{"home/ted/lizhe", "home/ted/li/lizhe", "home/ted/li"}
	fmt.Println(commonPathPrefix(strs))
}
