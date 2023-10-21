package main

import (
	"math/rand"
	"testing"
)

var testKeys = [...]string{"hamlet", "as-like", "othello"}

var testInvoice = Invoice{"BigCo", []Performance{
	{testKeys[0], 55},
	{testKeys[1], 35},
	{testKeys[2], 40},
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

func BenchmarkStatementFixed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Statement(testInvoice, testPlays)
	}
}

func BenchmarkStatementRandom(b *testing.B) {
	const seed = 0x6585
	const nperfs = 1024
	invoice := Invoice{
		Customer:     "Random",
		Performances: make([]Performance, nperfs),
	}

	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < nperfs; i++ {
		idx := rng.Intn(len(testPlays))
		invoice.Performances = append(invoice.Performances, Performance{testKeys[idx], rng.Int()})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Statement(invoice, testPlays)
	}
}
