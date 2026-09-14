package main

import (
	"unsafe"

	"./before"
	"./proper"
	"./types"

	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/fmt/fmt_"
	"github.com/anton2920/gofa/ints"
	"github.com/anton2920/gofa/trace/trace_"
)

var Movies = [...]types.Movie{
	{Title: "Hamlet", PriceCode: types.MovieTypeNewRelease},
	{Title: "As You Like It", PriceCode: types.MovieTypeChildrens},
	{Title: "Othello", PriceCode: types.MovieTypeRegular},

	{Title: "Terminator", PriceCode: types.MovieTypeNewRelease},
	{Title: "Saw", PriceCode: types.MovieTypeChildrens},
	{Title: "Kolobok", PriceCode: types.MovieTypeRegular},
}

var TestData = [...]types.Customer{
	{
		Name: "BigCo",
		Rentals: []types.Rental{
			{Movie: Movies[0], DaysRented: 55},
			{Movie: Movies[1], DaysRented: 35},
			{Movie: Movies[2], DaysRented: 40},
		},
	},
	{
		Name: "Ivan",
		Rentals: []types.Rental{
			{Movie: Movies[3], DaysRented: 14},
			{Movie: Movies[4], DaysRented: 5},
			{Movie: Movies[5], DaysRented: 23},
		},
	},
}

func Main(ctx *context.Context) {
	{
		fmt_.Println(ctx, ctx.Fmt.Reset().S("Before refactoring..."))

		for i := 0; i < len(TestData); i++ {
			customer := &TestData[i]
			statement := before.Statement(ctx, customer)
			fmt_.Println(ctx, ctx.Fmt.Reset().S(statement))
		}
	}
	if false {
		fmt_.Println(ctx, ctx.Fmt.Reset().S("My initial version..."))
		save := ctx.Arena
		{
			statements := StatementsISPC(ctx, TestData[:])
			for i := 0; i < len(statements); i++ {
				statement := statements[i]
				fmt_.Println(ctx, ctx.Fmt.Reset().S(statement))
			}
		}
		ctx.Arena = save
	}
	if false {
		fmt_.Println(ctx, ctx.Fmt.Reset().S("My initial version (with prepared data)..."))
		save := ctx.Arena
		{
			var data PreparedData
			PrepareDataForStatementsISPC(ctx, unsafe.Pointer(&data), TestData[:])

			statements := StatementsISPCWithPreparedData(ctx, unsafe.Pointer(&data), TestData[:])
			for i := 0; i < len(statements); i++ {
				statement := statements[i]
				fmt_.Println(ctx, ctx.Fmt.Reset().S(statement))
			}
		}
		ctx.Arena = save
	}
	if false {
		fmt_.Println(ctx, ctx.Fmt.Reset().S("My new version..."))
		save := ctx.Arena
		{
			statements := StatementsISPC2(ctx, TestData[:])
			for i := 0; i < len(statements); i++ {
				statement := statements[i]
				fmt_.Println(ctx, ctx.Fmt.Reset().S(statement))
			}
		}
		ctx.Arena = save
	}
	if false {
		fmt_.Println(ctx, ctx.Fmt.Reset().S("My new version (with prepared data)..."))
		save := ctx.Arena
		{
			var data PreparedData2
			PrepareDataForStatementsISPC2(ctx, unsafe.Pointer(&data), TestData[:])

			statements := StatementsISPCWithPreparedData2(ctx, unsafe.Pointer(&data), TestData[:])
			for i := 0; i < len(statements); i++ {
				statement := statements[i]
				fmt_.Println(ctx, ctx.Fmt.Reset().S(statement))
			}
		}
		ctx.Arena = save
	}
	{
		fmt_.Println(ctx, ctx.Fmt.Reset().S("My proper version..."))
		save := ctx.Arena
		{
			statements := proper.Statements(ctx, TestData[:])
			for i := 0; i < len(statements); i++ {
				statement := statements[i]
				fmt_.Println(ctx, ctx.Fmt.Reset().S(statement))
			}
		}
		ctx.Arena = save
	}
}

func main() {
	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithEvenlySplitSize(&ctx, ints.GiB(1)), "Failed to allocate required amount of memory")

	trace_.BeginProfile()
	Main(ctx.Noescape())
	trace_.EndAndPrintProfile()
}
