package main

import "fmt"

func f1() {
	// C style
	for i := 0; i < 10; i = i + 1 {
		fmt.Printf("%d\n", i)
	}
}

func f2() {
	for i := 9; i < 20; i = i + 1 {
		if i == 9 {
			continue
		}
		if i == 16 {
			break
		}
		fmt.Printf("%d\n", i)
	}
}

func main() {
	f1()
	f2()
}
