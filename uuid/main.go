package main

import (
	"fmt"
	"github.com/nu7hatch/gouuid"
)

func main() {
	u, err := uuid.NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(u)
}
