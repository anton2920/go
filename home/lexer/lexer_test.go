package main

import (
	"testing"

	"github.com/anton2920/gofa/ints"
)

const filename = "main.go"
const N = 512

func TestLexerFindTokens(t *testing.T) {
	contents, err := MapFileToMemory(filename)
	if err != nil {
		t.Fatalf("Failed to map file to memory: %v", err)
	}

	var l Lexer
	l.Init(filename, contents)

	beginsExpected := make([]int32, N)[:0]
	endsExpected := make([]int32, N)[:0]
	LexerFindTokensGo(&l, &beginsExpected, &endsExpected)

	beginsActual := make([]int32, N)[:0]
	endsActual := make([]int32, N)[:0]
	LexerFindTokens(&l, &beginsActual, &endsActual)

	if len(beginsExpected) != len(beginsActual) {
		t.Errorf("Expected %d begins, got %d", len(beginsExpected), len(beginsActual))
	}
	for i := 0; i < ints.Min(len(beginsExpected), len(beginsActual)); i++ {
		if (beginsExpected[i] != beginsActual[i]) || (endsExpected[i] != endsActual[i]){
			t.Errorf("For index %d expected %q, got %q", i, l.Contents[beginsExpected[i]:endsExpected[i]], l.Contents[beginsActual[i]:endsActual[i]])
		}
	}
}

func BenchmarkLexerFindTokens(b *testing.B) {
	contents, err := MapFileToMemory(filename)
	if err != nil {
		b.Fatalf("Failed to map file to memory: %v", err)
	}

	var l Lexer
	l.Init(filename, contents)

	begins := make([]int32, N)
	ends := make([]int32, N)

	b.SetBytes(int64(len(contents)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		begins = begins[:0]
		ends = ends[:0]
		LexerFindTokens(&l, &begins, &ends)
	}
}

func BenchmarkLexerFindTokensGo(b *testing.B) {
	contents, err := MapFileToMemory(filename)
	if err != nil {
		b.Fatalf("Failed to map file to memory: %v", err)
	}

	var l Lexer
	l.Init(filename, contents)

	begins := make([]int32, N)
	ends := make([]int32, N)

	b.SetBytes(int64(len(contents)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		begins = begins[:0]
		ends = ends[:0]
		LexerFindTokensGo(&l, &begins, &ends)
	}
}
