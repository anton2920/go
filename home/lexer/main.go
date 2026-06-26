package main

/*
#cgo LDFLAGS: -L. -Wl,-rpath,. -llexerispc
#include "liblexerispc.h"
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/anton2920/gofa/bytes"
	"github.com/anton2920/gofa/cgo"
	_ "github.com/anton2920/gofa/debug"
	"github.com/anton2920/gofa/os/posix/freebsd"
)

type Lexer struct {
	Filename string
	Contents []byte

	ErrorCount int
	CallCount  int

	rdOfft   int
	offt     int
	lineOfft int

	_ [32]byte
}

type Token uint8

func (l *Lexer) Init(filename string, contents []byte) {
	l.Filename = filename
	l.Contents = contents
}

func orexit(msg string, err error) {
	if err != nil {
		println("ERROR:", msg, err.Error())
		freebsd.Exit(1)
	}
}

func main() {
	runtime.AllocationsAreDisabled = true

	const filename = "main.go"
	fd, err := freebsd.Open(filename, freebsd.O_RDONLY, 0); orexit("failed to open lexer input file:", err)

	var st freebsd.Stat_t
	err = freebsd.Fstat(fd, &st); orexit("failed to stat lexer input file:", err)

	ptr, err := freebsd.Mmap(nil, uint(st.Size), freebsd.PROT_READ, freebsd.MAP_PRIVATE, fd, 0); orexit("failed to memory-map lexer input file:", err)
	contents := bytes.SliceFromUnsafePointer(ptr, int(st.Size))
	
	err = freebsd.Close(fd); orexit("failed to close lexer input file:", err)

	var l Lexer
	l.Init(filename, contents)

	for {
		tok := Token(cgo.Call1(C.LexerNextToken, uintptr(unsafe.Pointer(&l))))
		if tok == 128 {
			break
		}
	}

	println("Was in 'Next'", l.CallCount, "times")
	println("STOP")
}
