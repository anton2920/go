package main

type PlayType int

const (
	Tragedy PlayType = iota
	Comedy
)

type Play struct {
	Name string
	Type PlayType
}

type Plays map[string]Play

type Performance struct {
	PlayID   string
	Audience int
}

type Invoice struct {
	Customer     string
	Performances []Performance
}

func main() {

}
