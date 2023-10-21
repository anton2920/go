package main

import "testing"

var testInvoice = Invoice{"BigCo", []Performance{
	{"hamlet", 55},
	{"as-like", 35},
	{"othello", 40},
}}

var testPlays = Plays{
	"hamlet":  Play{"Hamlet", Tragedy},
	"as-like": Play{"As You Like It", Comedy},
	"othello": Play{"Otherllo", Tragedy},
}

func TestStatement(t *testing.T) {
	const expected = `Statement for BigCo
	Hamlet: $650.00 (55 seats)
	As You Like It: $580.00 (35 seats)
	Otherllo: $500.00 (40 seats)
Amount owed is $1730.00
You earned 47 credits
`
	statement := Statement(testInvoice, testPlays)
	if statement != expected {
		t.Errorf("expected %s, got %s", expected, statement)
	}
}

func BenchmarkStatement(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Statement(testInvoice, testPlays)
	}
}
