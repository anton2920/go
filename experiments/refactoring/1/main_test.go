package main

import (
	"testing"
	"unsafe"

	"./before"
	"./proper"
	"./types"
	"./utils"

	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/ints"
)

type (
	StatementFunction                 func(*context.Context, *types.Customer) string
	PrepareDataForStatementFunction   func(*context.Context, unsafe.Pointer, *types.Customer)
	StatementFunctionWithPreparedData func(*context.Context, unsafe.Pointer, *types.Customer) string

	StatementsFunction                 func(*context.Context, []types.Customer) []string
	PrepareDataForStatementsFunction   func(*context.Context, unsafe.Pointer, []types.Customer)
	StatementsFunctionWithPreparedData func(*context.Context, unsafe.Pointer, []types.Customer) []string
)

const (
	Seed = 1234

	MaxCustomers = 100
	MaxRentals   = 100
	MaxDays      = 100
)

var (
	StatementFunctions = [...]StatementFunction{
		before.Statement, before.StatementBytesBuffer, before.StatementMyFmt,
		proper.Statement,
		StatementISPC, StatementISPC2,
	}
	StatementFunctionsWithPreparedData = [...]struct {
		PrepareDataForStatement PrepareDataForStatementFunction
		Statement               StatementFunctionWithPreparedData
		DataSize                uintptr
		DataAlignment           uintptr
	}{
		{PrepareDataForStatementISPC, StatementISPCWithPreparedData, unsafe.Sizeof(PreparedData{}), unsafe.Alignof(PreparedData{})},
		{PrepareDataForStatementISPC2, StatementISPCWithPreparedData2, unsafe.Sizeof(PreparedData2{}), unsafe.Alignof(PreparedData2{})},
	}

	StatementsFunctions = [...]StatementsFunction{
		before.Statements, before.StatementsBytesBuffer, before.StatementsMyFmt,
		proper.Statements,
		StatementsISPC, StatementsISPC2,
	}
	StatementsFunctionsWithPreparedData = [...]struct {
		PrepareDataForStatements PrepareDataForStatementsFunction
		Statements               StatementsFunctionWithPreparedData
		DataSize                 uintptr
		DataAlignment            uintptr
	}{
		{PrepareDataForStatementsISPC, StatementsISPCWithPreparedData, unsafe.Sizeof(PreparedData{}), unsafe.Alignof(PreparedData{})},
		{PrepareDataForStatementsISPC2, StatementsISPCWithPreparedData2, unsafe.Sizeof(PreparedData2{}), unsafe.Alignof(PreparedData2{})},
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
	customer := &TestData[0]

	for _, fn := range StatementFunctions {
		t.Run(utils.FunctionName(fn), func(t *testing.T) {
			statement := fn(&ctx, customer)
			if statement != expected {
				t.Errorf("expected %q, got %q", expected, statement)
			}
			ctx.Arena.Reset()
		})
	}
}

func TestStatementWithPreparedData(t *testing.T) {
	const expected = `Rental Record for BigCo
	Hamlet	165
	As You Like It	49.5
	Othello	59
Amount owed is 273.5
You earned 4 frequent renter points
`
	var ctx context.Context
	ctx.InitWithEvenlySplitByteSlice(make([]byte, 4096))
	customer := &TestData[0]

	for _, sample := range StatementFunctionsWithPreparedData {
		t.Run(utils.FunctionName(sample.Statement), func(t *testing.T) {
			ctx.Arena.Reset()
			data := ctx.Arena.PushSizeWithAlignment(sample.DataSize, sample.DataAlignment)
			sample.PrepareDataForStatement(&ctx, data, customer)
			save := ctx.Arena

			statement := sample.Statement(&ctx, data, customer)
			if statement != expected {
				t.Errorf("expected %q, got %q", expected, statement)
			}
			ctx.Arena = save
		})
	}
}

func BenchmarkStatement(b *testing.B) {
	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithFourSizes(&ctx, ints.MiB(1), ints.KiB(1), ints.KiB(1), ints.KiB(1)), "Failed to allocate required amount of memory")

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
	context_.Must(&ctx, context_.BootstrapWithFourSizes(&ctx, ints.MiB(1), ints.KiB(1), ints.KiB(1), ints.KiB(1)), "Failed to allocate required amount of memory")

	b.ResetTimer()
	for _, customer := range [...]*types.Customer{&TestData[0], utils.GenerateRandomCustomer(Seed, MaxRentals, MaxDays)} {
		b.Run(customer.Name, func(b *testing.B) {
			for _, sample := range StatementFunctionsWithPreparedData {
				b.Run(utils.FunctionName(sample.Statement), func(b *testing.B) {
					ctx.Arena.Reset()
					data := ctx.Arena.PushSizeWithAlignment(sample.DataSize, sample.DataAlignment)
					sample.PrepareDataForStatement(&ctx, data, customer)
					save := ctx.Arena

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						sample.Statement(&ctx, data, customer)
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
			for _, sample := range StatementsFunctionsWithPreparedData {
				b.Run(utils.FunctionName(sample.Statements), func(b *testing.B) {
					ctx.Arena.Reset()
					data := ctx.Arena.PushSizeWithAlignment(sample.DataSize, sample.DataAlignment)
					sample.PrepareDataForStatements(&ctx, data, customers)
					save := ctx.Arena

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						sample.Statements(&ctx, data, customers)
						ctx.Arena = save
					}
				})
			}
		})
	}
}
