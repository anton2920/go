package main

import (
	"./after"
	"./before"
	"./json"
	"./types"
	"./utils"

	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/fmt/fmt_"
	"github.com/anton2920/gofa/ints"
	"github.com/anton2920/gofa/trace/trace_"
)

func Main(ctx *context.Context) {
	var random types.Invoice
	random.Customer = "Random"
	random.Performances = make([]types.Performance, 20)
	utils.FillRandomPerformances(random.Performances, 6585)
	json.Invoices = append(json.Invoices, random)

	{
		fmt_.Println(ctx, ctx.Fmt.Reset().S("Before refactoring:"))
		for i := 0; i < len(json.Invoices); i++ {
			invoice := &json.Invoices[i]
			fmt_.Println(ctx, ctx.Fmt.Reset().S(before.Statement(invoice, json.Plays)))
		}
	}

	{
		fmt_.Println(ctx, ctx.Fmt.Reset().S("After refactoring:"))
		for i := 0; i < len(json.Invoices); i++ {
			invoice := &json.Invoices[i]
			fmt_.Println(ctx, ctx.Fmt.Reset().S(after.Statement(invoice, json.Plays)))
		}
	}
}

func main() {
	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithEvenlySplitSize(&ctx, ints.GiB(1)), "Failed to allocate required amount of memory")

	trace_.BeginProfile()
	Main(ctx.Noescape())
	trace_.EndAndPrintProfile()
}
