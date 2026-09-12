package main

/*
#cgo LDFLAGS: -L. -Wl,-rpath,. -llexerispc
#include "liblexerispc.h"
*/
import "C"

import (
	"fmt"
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

func MapFileToMemory(filename string) ([]byte, error) {
	fd, err := freebsd.Open(filename, freebsd.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open lexer input file: %v", err)
	}

	var st freebsd.Stat_t
	if err := freebsd.Fstat(fd, &st); err != nil {
		return nil, fmt.Errorf("failed to stat lexer input file: %v", err)
	}

	ptr, err := freebsd.Mmap(nil, uint(st.Size), freebsd.PROT_READ, freebsd.MAP_PRIVATE, fd, 0);
	if err != nil {
		return nil, fmt.Errorf("failed to memory-map lexer input file: %v", err)
	}
	contents := bytes.SliceFromUnsafePointer(ptr, int(st.Size))

	if err := freebsd.Close(fd); err != nil {
		return nil, fmt.Errorf("failed to close lexer input file: %v", err)
	}

	return contents, nil
}

func LexerFindTokens(l *Lexer, begins *[]int32, ends *[]int32) {
	//C.LexerFindTokens((*C.struct_Lexer)(pointers.UnsafeNoescape(unsafe.Pointer(l))), (*C.struct_goslice)(pointers.UnsafeNoescape(unsafe.Pointer(begins))), (*C.struct_goslice)(pointers.UnsafeNoescape(unsafe.Pointer(ends))))
	cgo.Call3(C.LexerFindTokens, uintptr(unsafe.Pointer(l)), uintptr(unsafe.Pointer(begins)), uintptr(unsafe.Pointer(ends)))
}

func LexerFindTokensGo(l *Lexer, begins *[]int32, ends *[]int32) {
	var begin, end int32

	isspace := func(ch byte) bool {
		return (ch == ' ') || (ch == '\t') || (ch == '\n') || (ch == '\r')
	}

	for int(end) < len(l.Contents) {
		ch := l.Contents[end]
		if !isspace(ch) {
			end++
			continue
		}

		if begin < end {
			*begins = append(*begins, begin)
			*ends = append(*ends, end)
		}

		begin = end + 1
		end++
	}
}

func main() {
	runtime.AllocationsAreDisabled = true

	const filename = "main.go"
	contents, err := MapFileToMemory(filename); orexit("failed to map file to memory:", err)

	var l Lexer
	l.Init(filename, contents)

	begins := make([]int32, 512)[:0]
	ends := make([]int32, 512)[:0]

	LexerFindTokens(&l, &begins, &ends)
	//LexerFindTokensGo(&l, (*[]int32)(pointers.UnsafeNoescape(unsafe.Pointer(&begins))), (*[]int32)(pointers.UnsafeNoescape(unsafe.Pointer(&ends))))

	for i := 0; i < len(begins); i++ {
		println(bytes.AsString(l.Contents[begins[i]:ends[i]]))
	}

	println("STOP")
}
