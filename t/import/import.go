package main

import (
	"./mylib"
	"fmt"
)

func f1() {
	mylib.MyFunc()
	fmt.Printf("%d\n", mylib.MyNumber)
	fmt.Printf("%s\n", mylib.MyString)
}

func main() {
	f1()
}
