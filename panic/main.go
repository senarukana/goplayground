package main

import (
	"fmt"
	"math"
)

func main() {
	var i int64 = 2 << 34
	if r, err := IntFromInt64(i); err != nil {
		fmt.Println(r)
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func IntFromInt64(x int64) (i int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	i = ConvertInt64ToInt(x)
	return i, nil
}

func ConvertInt64ToInt(x int64) int {
	if math.MinInt32 <= x && x <= math.MaxInt32 {
		return int(x)
	}
	panic(fmt.Sprintf("%d is out of the int32 range", x))
}
