package main

import (
	"testing"

	"./after"
	"./before"
	"./json"
	"./types"
	"./utils"
)

const (
	Nperfs = 1024
	Seed   = 6585
)

var StatementFunctions = [...]types.StatementFunction{before.Statement, after.Statement}

func TestStatement(t *testing.T) {
	const expected = `Statement for BigCo
	Hamlet: $650.00 (55 seats)
	As You Like It: $580.00 (35 seats)
	Otherllo: $500.00 (40 seats)
Amount owed is $1730.00
You earned 47 credits
`

	invoice := &json.Invoices[0]
	for _, fn := range StatementFunctions {
		t.Run(utils.FunctionName(fn), func(t *testing.T) {
			statement := fn(invoice, json.Plays)
			if statement != expected {
				t.Errorf("expected %q, got %q", expected, statement)
			}
		})
	}
}

func BenchmarkStatement(b *testing.B) {
	var random types.Invoice
	random.Customer = "Random"
	random.Performances = make([]types.Performance, Nperfs)
	utils.FillRandomPerformances(random.Performances, Seed)

	b.ResetTimer()
	for _, invoice := range [...]*types.Invoice{&json.Invoices[0], &random} {
		b.Run(invoice.Customer, func(b *testing.B) {
			for _, fn := range StatementFunctions {
				b.Run(utils.FunctionName(fn), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn(invoice, json.Plays)
					}
				})
			}
		})
	}
}
