package main

// https://golang.org/ref/spec#Predeclared_identifiers

// Functions:
//	append cap close complex copy delete imag len
//	make new panic print println real recover

var builtinCode = `
package builtin

const MiniGo int = 1

func len(x interface{}) int {
}

func make(x interface{}) interface{} {
}

func println(s string) {
	puts(s)
}

func recover() interface{} {
}

type error interface {
	Error() string
}
`
