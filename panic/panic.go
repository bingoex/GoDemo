package main

import (
	"fmt"
)

func rrr() {
	if v := recover(); v != nil {
		fmt.Println("paniced with:", v)
	}
}

func main() {
	defer rrr() //出现了painc则调用这个函数

	panic(33) // (*(int *)(0)) = 5

	fmt.Println("behind panic never go here")
}
