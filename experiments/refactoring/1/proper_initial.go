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

func Statements(ctx *context.Context, customers []types.Customer) []string {
	t := trace_.Begin("")

	n := len(customers)
	var statements []string
	var (
		customersBeginIndicies []uint32
		amountsOwed            []float32
		rentersPoints          []uint16
	)
	var (
		rentalsDays  []int32
		rentalsTypes []uint8
	)
	var (
		rentalsCostsX100 []uint32
		rentalsCredits   []uint16
	)

	{
		t := trace_.Begin("Statements/PreparingData")

		amountsOwed = ctx.Arena.PushFloat32Array(n) /* TODO(anton2920): get rid of decimals alltogether... */
		rentersPoints = ctx.Arena.PushUint16Array(n)
		customersBeginIndicies = ctx.Arena.PushUint32Array(n + 1)

		var nrentals int
		for i := 0; i < len(customers); i++ {
			customersBeginIndicies[i] = uint32(nrentals)
			nrentals += len(customers[i].Rentals)
		}
		customersBeginIndicies[n] = uint32(nrentals)

		rentalsDays = ctx.Arena.PushInt32Array(nrentals)
		rentalsTypes = ctx.Arena.PushUint8Array(nrentals) /* TODO(anton2920): 255 movies types should be enough, right...? */

		var curr uint32
		for i := 0; i < len(customers); i++ {
			customer := &customers[i]
			for j := 0; j < len(customer.Rentals); j++ {
				rental := &customer.Rentals[j]
				rentalsTypes[curr] = uint8(rental.Movie.PriceCode)
				rentalsDays[curr] = int32(rental.DaysRented)
				curr++
			}
		}

		rentalsCostsX100 = ctx.Arena.PushUint32Array(nrentals)
		rentalsCredits = ctx.Arena.PushUint16Array(nrentals)

		trace_.End(t)
	}

	{
		t := trace_.Begin("Statements/Calculation")

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
		}{&customersBeginIndicies[0], &amountsOwed[0], &rentersPoints[0], &rentalsDays[0], &rentalsTypes[0], &rentalsCostsX100[0], &rentalsCredits[0], uint32(len(customers)), uint32(len(rentalsTypes))})))

		trace_.End(t)
	}

	var f fmt.Formatter
	{
		t := trace_.Begin("Statements/Formatting")

		statements = ctx.Arena.PushStringArray(n)
		for i := 0; i < n; i++ {
			customer := &customers[i]

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), int(uintptr(ctx.Arena.Size)-ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(customer.Name).Ln()

			begin := int(customersBeginIndicies[i])
			for j := 0; j < len(customer.Rentals); j++ {
				rental := &customer.Rentals[j]
				f.S("\t").S(rental.Movie.Title).S("\t").G32(float32(rentalsCostsX100[begin+j]) / 100).Ln()
			}

			f.S("Amount owed is ").G32(amountsOwed[i]).Ln()
			f.S("You earned ").D(int(rentersPoints[i])).S(" frequent renter points").Ln()

			statements[i] = f.String()
			ctx.Arena.PrevOfft = ctx.Arena.CurrOfft
			ctx.Arena.CurrOfft += uintptr(len(statements[i]))
		}

		trace_.End(t)
	}

	trace_.End(t)
	return statements
}

func Statement(ctx *context.Context, customer *types.Customer) string {
	customers := [...]types.Customer{*customer}
	statements := Statements(ctx, customers[:])
	return statements[0]
}
