package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anton2920/gofa/bytes"
)

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func Usage() {
	Fatalf("usage: adiff file1.csv file2.csv\n")
}

func ParseFile(filename string) (map[string]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %v", filename, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read entire file %q: %v", filename, err)
	}

	contents := bytes.AsString(data)
	m := make(map[string]string)

	var done bool
	for !done {
		line, rest, ok := strings.Cut(contents, "\r\n")
		if !ok {
			done = true
		}

		uuid, rem, ok := strings.Cut(line, ";")
		title, _, _ := strings.Cut(rem, ";")
		m[uuid] = title

		contents = rest
	}

	return m, nil
}

func main() {
	if len(os.Args) != 3 {
		Usage()
	}

	m1, err := ParseFile(os.Args[1])
	if err != nil {
		Fatalf("Failed to parse file %q: %v", os.Args[1], err)
	}

	m2, err := ParseFile(os.Args[2])
	if err != nil {
		Fatalf("Failed to parse file %q: %v", os.Args[2], err)
	}

	for k, v1 := range m1 {
		if v2, ok := m2[k]; ok {
			if v1 != v2 {
				fmt.Printf("For UUID %q first file has '%s', second has '%s'\n", k, v1, v2)
			}
			delete(m1, k)
			delete(m2, k)
		}
	}

	for k, v := range m1 {
		fmt.Printf("UUID %q is present only in first file with value '%s'\n", k, v)
	}

	for k, v := range m2 {
		fmt.Printf("UUID %q is present only in second file with value '%s'\n", k, v)
	}
}
