package main

import "fmt"

func getNumber() int

func main() {
	asmRet := getNumber()
	fmt.Printf("Assembly returned: %d\n", asmRet)
}
