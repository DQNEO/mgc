package main

import (
	"fmt"
	"strconv"
)

func f1() {
	var a string = "10485760"
	var i int
	i, _ = strconv.Atoi(a)
	fmt.Printf("%d\n", i-10485760) // 0

	a = "1"
	i, _ = strconv.Atoi(a)
	fmt.Printf("%d\n", i) // 1
}

func f2() {
	var s string
	s = strconv.Itoa(0)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(7)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(10)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(100)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(1234567890)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(-1)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(-7)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(-10)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(-100)
	fmt.Printf("%s\n", s)

	s = strconv.Itoa(-1234567890)
	fmt.Printf("%s\n", s)

}

func main() {
	f1()
	f2()
}