package main

type Type int32

const (
	None = Type(iota)
	Has
	Insert
)

type Command struct {
	Type  Type
	Value int32
}

var s [100]chan Command
var s0 chan bool

func S(i int) {
	if i >= len(s) {
		return
	}
	for n := range s[i] {
		switch n.Type {
		case Has:
			s0 <- false
		case Insert:
			for m := range s[i] {
				switch m.Type {
				case Has:
					if m.Value <= n.Value {
						s0 <- (m.Value == n.Value)
					} else if i < len(s)-1 {
						s[i+1] <- Command{Has, m.Value}
					} else {
						s0 <- false
					}
				case Insert:
					if i < len(s)-1 {
						if m.Value < n.Value {
							s[i+1] <- Command{Insert, n.Value}
							n.Value = m.Value
						} else if m.Value > n.Value {
							s[i+1] <- Command{Insert, m.Value}
						}
					}
				}
			}
		}
	}
}

func main() {
	s0 = make(chan bool)
	for i := 0; i < len(s); i++ {
		s[i] = make(chan Command)
		go S(i)
	}

	var numbers [len(s)*4 + 1]int32
	for i := 0; i < len(numbers); i++ {
		// numbers[i] = int32(rand.Int() % 1000)
		numbers[i] = int32(i / 4)
	}

	for i := 0; i < len(numbers); i++ {
		s[0] <- Command{Insert, numbers[i]}
		println("Inserting", numbers[i])
	}

	for i := 0; i < len(numbers); i++ {
		s[0] <- Command{Has, numbers[i]}
		println("Search for", numbers[i], <-s0)
	}
}
