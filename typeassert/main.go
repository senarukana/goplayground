package main

import (
	"fmt"
	"json"
)

type Test struct {
	a int
}

func main() {
	var a int
	var b float64
	classifier(a, b, Test{})
}
func classifier(items ...interface{}) {
	for i, x := range items {
		switch x.(type) {
		case bool:
			fmt.Printf("param #%d is a bool\n", i)
		case float64:
			fmt.Printf("param #%d is a float64\n", i)
		case int, int8, int16, int32, int64:
			fmt.Printf("param #%d is an int\n", i)
		case nil:
			fmt.Printf("param #%d is nil\n", i)
		case string:
			fmt.Printf("param #%d is a string\n", i)
		default:
			fmt.Printf("param #%d's type is unknown\n", i)
		}
	}
}
