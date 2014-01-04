package main

import (
	"fmt"
	"reflect"
)

func IndexReflectX(xs interface{}, x interface{}) int {
	if slice := reflect.ValueOf(xs); slice.Kind() == reflect.Slice {
		for i := 0; i < slice.Len(); i++ {
			if reflect.DeepEqual(x, slice.Index(i)) {
				return i
			}
		}
	}
	return -1
}

func main() {
	i := IndexReflectX(a, a1)
	fmt.Println(i)
}
