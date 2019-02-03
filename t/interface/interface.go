package main

import "fmt"

func f1() {
	var p *Point
	p = &Point{
		x: 1,
		y: 2,
	}
	sum := p.sum()
	fmt.Printf("%d\n", sum - 2) // 1
}

func f2() {
	var myInterface MyInterface
	ptr := &Point{
		x: 2,
		y: 3,
	}
	myInterface = ptr
	sum := myInterface.sum()
	fmt.Printf("%d\n", sum - 3) // 2

	diff := myInterface.diff()
	fmt.Printf("%d\n", diff + 2) // 3
}

func f3() {
	var myInterface MyInterface
	ptr := &Asset{
		money: 2,
		stock: 3,
	}
	myInterface = ptr
	sum := myInterface.sum()
	fmt.Printf("%d\n", sum - 1) // 4

	diff := myInterface.diff()
	fmt.Printf("%d\n", diff + 4) // 5
}

func f4(bol bool) {
	var myInterface MyInterface
	point := &Point{
		x: 2,
		y: 4,
	}

	asset := &Asset{
		money: 2,
		stock: 6,
	}

	if bol {
		myInterface = point
	} else {
		myInterface = asset
	}

	sum := myInterface.sum()

	fmt.Printf("%d\n", sum) // 6, 8

	diff := myInterface.diff()
	fmt.Printf("%d\n", diff + 5) // 7, 9
}

func main() {
	f1()
	f2()
	f3()
	f4(true)
	f4(false)
}

type MyInterface interface {
	sum() int
	diff() int
}

type Point struct {
	x int
	y int
}

func (p *Point) sum() int {
	return p.x + p.y
}

func (p *Point) diff() int {
	return p.y - p.x
}

type Asset struct {
	money int
	stock int
}

func (a *Asset) sum() int {
	return a.money + a.stock
}

func (a *Asset) diff() int {
	return a.stock - a.money
}
