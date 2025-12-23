package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anton2920/gofa/bytes"
)

func Fatalf(format string, args ...interface{}) {
	if format[len(format)-1] != '\n' {
		format = format + "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func Usage() {
	Fatalf("usage: adiff2 file1.csv file2.csv\n")
}

type Line struct {
	Units string
	Value [2]string
}

func ParseFile(filename string) (map[string][]Line, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %v", filename, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read entire file %q: %v", filename, err)
	}

	contents := bytes.AsString(data)
	m := make(map[string][]Line)

	var done bool
	for i := 0; !done; i++ {
		line, rest, ok := strings.Cut(contents, "\r\n")
		if !ok {
			done = true
		}
		if len(line) == 0 {
			break
		}

		var l Line
		{
			var name string
			var rest string

			if line[0] == '"' {
				name, rest, ok = strings.Cut(line[1:], `";`)
				name = `"` + name + `"`
			} else {
				name, rest, ok = strings.Cut(line, ";")
			}
			if !ok {
				return nil, fmt.Errorf("%s:%d: unexpected EOL, when parsing name", filename, i)
			}

			l.Units, rest, ok = strings.Cut(rest, ";")
			if !ok {
				return nil, fmt.Errorf("%s:%d: unexpected EOL, when parsing units", filename, i)
			}

			l.Value[0], rest, ok = strings.Cut(rest, ";")
			if !ok {
				return nil, fmt.Errorf("%s:%d: unexpected EOL, when parsing value[0]", filename, i)
			}

			l.Value[1], rest, ok = strings.Cut(rest, ";")

			m[name] = append(m[name], l)
		}

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
			var i int

			for i = 0; i < min(len(v1), len(v2)); i++ {
				fmt.Printf("%s;%s;%s;%s;;%s;%s;%s;%s\n", k, v1[i].Units, v1[i].Value[0], v1[i].Value[1], k, v2[i].Units, v2[i].Value[0], v2[i].Value[1])
			}
			for ; i < len(v1); i++ {
				fmt.Printf("%s;%s;%s;%s\n", k, v1[i].Units, v1[i].Value[0], v1[i].Value[1])
			}
			for ; i < len(v2); i++ {
				fmt.Printf(";;;;;%s;%s;%s;%s\n", k, v2[i].Units, v2[i].Value[0], v2[i].Value[1])
			}

			delete(m1, k)
			delete(m2, k)
		}
	}

	for k, v := range m1 {
		for i := 0; i < len(v); i++ {
			fmt.Printf("%s;%s;%s;%s\n", k, v[i].Units, v[i].Value[0], v[i].Value[1])
		}
	}

	for k, v := range m2 {
		for i := 0; i < len(v); i++ {
			fmt.Printf(";;;;;%s;%s;%s;%s\n", k, v[i].Units, v[i].Value[0], v[i].Value[1])
		}
	}
}
