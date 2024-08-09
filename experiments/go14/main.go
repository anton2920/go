package main

func Write(fd int32, buf string) int

func Exit(status int32)

//go:nosplit
func main() {
	Write(1, "Hello, world!\n")
	Exit(0)
}
