package main

import (
	"fmt"
)

func Filter(limit int, predicate func(int) bool, appender func(int)) {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			appender(i)
		}
	}
}

func main() {
	readings := []int{4, -2, 3, 1, 5, 0}
	even := make([]int, 0, len(readings))
	Filter(len(readings), func(i int) bool { return i%2 == 0 }, func(i int) { even = append(even, i) })
	fmt.Printf("%#v\n", even)
}
