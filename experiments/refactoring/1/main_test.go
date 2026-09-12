package main

import (
	"testing"

	"./before"
	"./types"
	"./utils"

	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/ints"
)

type (
	StatementFunction                 func(*context.Context, *types.Customer) string
	StatementFunctionWithPreparedData func(*context.Context, *PreparedData, *types.Customer) string

	StatementsFunction                 func(*context.Context, []types.Customer) []string
	StatementsFunctionWithPreparedData func(*context.Context, *PreparedData, []types.Customer) []string
)

const (
	Seed = 1234

	MaxCustomers = 100
	MaxRentals   = 100
	MaxDays      = 100
)

var (
	StatementFunctions = [...]StatementFunction{
		before.StatementOriginal, before.StatementBytesBuffer, before.StatementMyFmt,
		Statement,
	}
	StatementFunctionsWithPreparedData = [...]StatementFunctionWithPreparedData{
		StatementWithPreparedData,
	}

	StatementsFunctions = [...]StatementsFunction{
		before.StatementsOriginal, before.StatementsBytesBuffer, before.StatementsMyFmt,
		Statements,
	}
	StatementsFunctionsWithPreparedData = [...]StatementsFunctionWithPreparedData{
		StatementsWithPreparedData,
	}
)

func TestStatement(t *testing.T) {
	const expected = `Rental Record for BigCo
	Hamlet	165
	As You Like It	49.5
	Othello	59
Amount owed is 273.5
You earned 4 frequent renter points
`

	var ctx context.Context
	ctx.InitWithEvenlySplitByteSlice(make([]byte, 4096))

	for _, fn := range StatementFunctions {
		t.Run(utils.FunctionName(fn), func(t *testing.T) {
			statement := fn(&ctx, &TestData[0])
			if statement != expected {
				t.Errorf("expected %q, got %q", expected, statement)
			}
			ctx.Arena.Reset()
		})
	}

	customers := TestData[:1]

	var data PreparedData
	PrepareDataForStatements(&ctx, &data, customers)
	save := ctx.Arena

	for _, fn := range StatementFunctionsWithPreparedData {
		t.Run(utils.FunctionName(fn), func(t *testing.T) {
			statement := fn(&ctx, &data, &customers[0])
			if statement != expected {
				t.Errorf("expected %q, got %q", expected, statement)
			}
			ctx.Arena = save
		})
	}
}

func BenchmarkStatement(b *testing.B) {
	var ctx context.Context
	ctx.InitWithEvenlySplitByteSlice(make([]byte, 4096))

	b.ResetTimer()
	for _, customer := range [...]*types.Customer{&TestData[0], utils.GenerateRandomCustomer(Seed, MaxRentals, MaxDays)} {
		b.Run(customer.Name, func(b *testing.B) {
			for _, fn := range StatementFunctions {
				b.Run(utils.FunctionName(fn), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn(&ctx, customer)
						ctx.Arena.Reset()
					}
				})
			}
		})
	}
}

func BenchmarkStatementWithPreparedData(b *testing.B) {
	var ctx context.Context
	ctx.InitWithEvenlySplitByteSlice(make([]byte, 4096))

	b.ResetTimer()
	for _, customer := range [...]*types.Customer{&TestData[0], utils.GenerateRandomCustomer(Seed, MaxRentals, MaxDays)} {
		b.Run(customer.Name, func(b *testing.B) {
			ctx.Arena.Reset()

			var data PreparedData
			PrepareDataForStatement(&ctx, &data, customer)
			save := ctx.Arena

			for _, fn := range StatementFunctionsWithPreparedData {
				b.Run(utils.FunctionName(fn), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn(&ctx, &data, customer)
						ctx.Arena = save
					}
				})
			}
		})
	}
}

func BenchmarkStatements(b *testing.B) {
	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithFourSizes(&ctx, ints.MiB(1), ints.KiB(1), ints.KiB(1), ints.KiB(1)), "Failed to allocate required amount of memory")

	b.ResetTimer()
	for _, customers := range [...][]types.Customer{TestData[:], utils.GenerateRandomCustomers(Seed, MaxCustomers, MaxRentals, MaxDays)} {
		b.Run(customers[0].Name, func(b *testing.B) {
			for _, fn := range StatementsFunctions {
				b.Run(utils.FunctionName(fn), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn(&ctx, customers)
						ctx.Arena.Reset()
					}
				})
			}
		})
	}
}

func BenchmarkStatementsWithPreparedData(b *testing.B) {
	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithFourSizes(&ctx, ints.MiB(1), ints.KiB(1), ints.KiB(1), ints.KiB(1)), "Failed to allocate required amount of memory")

	b.ResetTimer()
	for _, customers := range [...][]types.Customer{TestData[:], utils.GenerateRandomCustomers(Seed, MaxCustomers, MaxRentals, MaxDays)} {
		b.Run(customers[0].Name, func(b *testing.B) {
			ctx.Arena.Reset()

			var data PreparedData
			PrepareDataForStatements(&ctx, &data, customers)
			save := ctx.Arena

			for _, fn := range StatementsFunctionsWithPreparedData {
				b.Run(utils.FunctionName(fn), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn(&ctx, &data, customers)
						ctx.Arena = save
					}
				})
			}
		})
	}
}
