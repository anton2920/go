package main

/*
int getNumber()
{
	__asm__ __volatile__ ("movq $42, %rax");
}
*/
import "C"
import "fmt"

func main() {
	fmt.Printf("Returned from C: %d\n", int(C.getNumber()))
}
