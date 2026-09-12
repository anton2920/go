package main

/*
#cgo LDFLAGS: -l1ispc -L. -Wl,-rpath=.
#include "lib1ispc.h"
*/
import "C"

import (
	"unsafe"

	"./types"

	"github.com/anton2920/gofa/cgo"
	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/fmt"
	"github.com/anton2920/gofa/pointers"
	"github.com/anton2920/gofa/trace/trace_"
)

type PreparedData struct {
	CustomersBeginIndicies []uint32
	AmountsOwed            []float32
	RentersPoints          []uint16

	RentalsDays  []int32
	RentalsTypes []uint8

	RentalsCostsX100 []uint32
	RentalsCredits   []uint16
}

func PrepareDataForStatements(ctx *context.Context, data *PreparedData, customers []types.Customer) {
	t := trace_.Begin("")

	n := len(customers)
	data.AmountsOwed = ctx.Arena.PushFloat32Array(n) /* TODO(anton2920): get rid of decimals alltogether... */
	data.RentersPoints = ctx.Arena.PushUint16Array(n)
	data.CustomersBeginIndicies = ctx.Arena.PushUint32Array(n + 1)

	var nrentals int
	for i := 0; i < len(customers); i++ {
		data.CustomersBeginIndicies[i] = uint32(nrentals)
		nrentals += len(customers[i].Rentals)
	}
	data.CustomersBeginIndicies[n] = uint32(nrentals)

	data.RentalsDays = ctx.Arena.PushInt32Array(nrentals)
	data.RentalsTypes = ctx.Arena.PushUint8Array(nrentals) /* TODO(anton2920): 255 movies types should be enough, right...? */

	var curr uint32
	for i := 0; i < len(customers); i++ {
		customer := &customers[i]
		for j := 0; j < len(customer.Rentals); j++ {
			rental := &customer.Rentals[j]
			data.RentalsTypes[curr] = uint8(rental.Movie.PriceCode)
			data.RentalsDays[curr] = int32(rental.DaysRented)
			curr++
		}
	}

	data.RentalsCostsX100 = ctx.Arena.PushUint32Array(nrentals)
	data.RentalsCredits = ctx.Arena.PushUint16Array(nrentals)

	trace_.End(t)
}

func StatementsWithPreparedData(ctx *context.Context, data *PreparedData, customers []types.Customer) []string {
	t := trace_.Begin("")

	n := len(customers)
	{
		t := trace_.Begin("StatementsWithPreparedData/Calculation")

		cgo.Call1(C.CalculateStatements, uintptr(unsafe.Pointer(&struct {
			CustomersBeginIndiciesPtr *uint32
			AmountsOwedPtr            *float32
			RentersPointsPtr          *uint16

			RentalsDaysPtr  *int32
			RentalsTypesPtr *uint8

			RentalsCostsX100Ptr *uint32
			RentalsCreditsPtr   *uint16

			NCustomers uint32
			NRentals   uint32
		}{&data.CustomersBeginIndicies[0], &data.AmountsOwed[0], &data.RentersPoints[0], &data.RentalsDays[0], &data.RentalsTypes[0], &data.RentalsCostsX100[0], &data.RentalsCredits[0], uint32(len(customers)), uint32(len(data.RentalsTypes))})))

		trace_.End(t)
	}

	var statements []string
	var f fmt.Formatter
	{
		t := trace_.Begin("StatementsWithPreparedData/Formatting")

		statements = ctx.Arena.PushStringArray(n)
		for i := 0; i < n; i++ {
			customer := &customers[i]

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), int(uintptr(ctx.Arena.Size)-ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(customer.Name).Ln()

			begin := int(data.CustomersBeginIndicies[i])
			for j := 0; j < len(customer.Rentals); j++ {
				rental := &customer.Rentals[j]
				f.S("\t").S(rental.Movie.Title).S("\t").G32(float32(data.RentalsCostsX100[begin+j]) / 100).Ln()
			}

			f.S("Amount owed is ").G32(data.AmountsOwed[i]).Ln()
			f.S("You earned ").D(int(data.RentersPoints[i])).S(" frequent renter points").Ln()

			statements[i] = f.String()
			ctx.Arena.PrevOfft = ctx.Arena.CurrOfft
			ctx.Arena.CurrOfft += uintptr(len(statements[i]))
		}

		trace_.End(t)
	}

	trace_.End(t)
	return statements
}

func PrepareDataForStatement(ctx *context.Context, data *PreparedData, customer *types.Customer) {
	customers := [...]types.Customer{*customer}
	PrepareDataForStatements(ctx, data, customers[:])
}

func StatementWithPreparedData(ctx *context.Context, data *PreparedData, customer *types.Customer) string {
	customers := [...]types.Customer{*customer}
	statements := StatementsWithPreparedData(ctx, data, customers[:])
	return statements[0]
}
