package main


var GlobalInt int = 1
var GlobalPtr *int = &GlobalInt

var pp1 *Point = &Point{x: 1, y: 2}
var pp2 *Point = &Point{x: 3, y: 4}

type Point struct {
	x int
	y int
}

func f1() {
	fmtPrintf(S("%d\n"), *GlobalPtr) // 1
}

func f2() {
	fmtPrintf(S("%d\n"), pp1.y) // 2
	fmtPrintf(S("%d\n"), pp2.x) // 3
}

type Gtype struct {
	typ  int
	size int
}

var gIntE = Gtype{typ: 7, size: 8}
var gInt = &gIntE

type DeclFunc struct {
	tok      bytes
	rettypes []*Gtype
}

var builtinLenGlobal = &DeclFunc{
	tok:      bytes("tok"),
	rettypes: []*Gtype{&gIntE, &gIntE},
}

func f3() {
	retTypes := builtinLenGlobal.rettypes
	fmtPrintf(S("%d\n"), len(retTypes)+2) // 4
	var gi *Gtype = retTypes[0]
	fmtPrintf(S("%d\n"), gi.typ-2) // 5
}

/*
func f4() {
	var builtinLenLocal = &DeclFunc{
		tok:"tok",
		rettypes: []*Gtype{&gIntE,&gIntE},
	}

	retTypes := builtinLenLocal.rettypes
	fmtPrintf(S("%d\n"), len(retTypes) + 4) // 6
	var gi *Gtype = retTypes[0]
	fmtPrintf(S("%d\n"), gi.size - 1) // 7
}
*/

func main() {
	f1()
	f2()
	f3()
	//f4()
}
