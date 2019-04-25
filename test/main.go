package main

import "fmt"

func main() {
	s := make(chan bool, 2)
	s <- true
	fmt.Println(1)
	s <- true
	fmt.Println(2)

	select {
	case s <- true:
		fmt.Println(3)
	default:
		fmt.Println("default")
	}
}
