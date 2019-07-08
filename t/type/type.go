package main


var myarray [2]myint = [2]myint{3, 2}

func anytype(x interface{}) {
	fmtPrintf(S("%d\n"), x)
}

func f1() {
	var a myint = '1'
	fmtPrintf(S("%c\n"), a)
}

func f2() {
	fmtPrintf(S("%d\n"), myarray[1])
	anytype(3)
	anytype(4)
}

func main() {
	f1()
	f2()
}

type myint int

type int byte
