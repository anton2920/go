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

func StatementsISPC(ctx *context.Context, customers []types.Customer) []string {
	t := trace_.Begin("")

	n := len(customers)
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
		rentalsCosts   []float32
		rentalsCredits []uint16
	)

	{
		t := trace_.Begin("StatementsISPC/PreparingData")

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

		rentalsCosts = ctx.Arena.PushFloat32Array(nrentals)
		rentalsCredits = ctx.Arena.PushUint16Array(nrentals)

		trace_.End(t)
	}

	{
		t := trace_.Begin("StatementsISPC/Calculation")

		cgo.Call1(C.CalculateStatements, uintptr(unsafe.Pointer(&struct {
			CustomersBeginIndiciesPtr *uint32
			AmountsOwedPtr            *float32
			RentersPointsPtr          *uint16

			RentalsDaysPtr  *int32
			RentalsTypesPtr *uint8

			RentalsCosts      *float32
			RentalsCreditsPtr *uint16

			NCustomers uint32
			NRentals   uint32
		}{&customersBeginIndicies[0], &amountsOwed[0], &rentersPoints[0], &rentalsDays[0], &rentalsTypes[0], &rentalsCosts[0], &rentalsCredits[0], uint32(len(customers)), uint32(len(rentalsTypes))})))

		trace_.End(t)
	}

	var statements []string
	{
		t := trace_.Begin("StatementsISPC/Formatting")

		var f fmt.Formatter
		statements = ctx.Arena.PushStringArray(n)
		for i := 0; i < n; i++ {
			customer := &customers[i]

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), int(uintptr(ctx.Arena.Size)-ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(customer.Name).Ln()

			begin := int(customersBeginIndicies[i])
			for j := 0; j < len(customer.Rentals); j++ {
				rental := &customer.Rentals[j]
				f.S("\t").S(rental.Movie.Title).S("\t").G32(rentalsCosts[begin+j]).Ln()
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

func StatementISPC(ctx *context.Context, customer *types.Customer) string {
	customers := [...]types.Customer{*customer}
	statements := StatementsISPC(ctx, customers[:])
	return statements[0]
}

type PreparedData struct {
	CustomersBeginIndicies []uint32
	AmountsOwed            []float32
	RentersPoints          []uint16

	RentalsDays  []int32
	RentalsTypes []uint8

	RentalsCosts   []float32
	RentalsCredits []uint16
}

func PrepareDataForStatementsISPC(ctx *context.Context, _data unsafe.Pointer, customers []types.Customer) {
	t := trace_.Begin("")

	data := (*PreparedData)(_data)

	n := len(customers)
	data.AmountsOwed = ctx.Arena.PushFloat32Array(n)
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

	data.RentalsCosts = ctx.Arena.PushFloat32Array(nrentals)
	data.RentalsCredits = ctx.Arena.PushUint16Array(nrentals)

	trace_.End(t)
}

func StatementsISPCWithPreparedData(ctx *context.Context, _data unsafe.Pointer, customers []types.Customer) []string {
	t := trace_.Begin("")

	data := (*PreparedData)(_data)

	n := len(customers)
	{
		t := trace_.Begin("StatementsISPCWithPreparedData/Calculation")

		cgo.Call1(C.CalculateStatements, uintptr(unsafe.Pointer(&struct {
			CustomersBeginIndiciesPtr *uint32
			AmountsOwedPtr            *float32
			RentersPointsPtr          *uint16

			RentalsDaysPtr  *int32
			RentalsTypesPtr *uint8

			RentalsCosts      *float32
			RentalsCreditsPtr *uint16

			NCustomers uint32
			NRentals   uint32
		}{&data.CustomersBeginIndicies[0], &data.AmountsOwed[0], &data.RentersPoints[0], &data.RentalsDays[0], &data.RentalsTypes[0], &data.RentalsCosts[0], &data.RentalsCredits[0], uint32(len(customers)), uint32(len(data.RentalsTypes))})))

		trace_.End(t)
	}

	var statements []string
	{
		t := trace_.Begin("StatementsISPCWithPreparedData/Formatting")

		var f fmt.Formatter
		statements = ctx.Arena.PushStringArray(n)
		for i := 0; i < n; i++ {
			customer := &customers[i]

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), int(uintptr(ctx.Arena.Size)-ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(customer.Name).Ln()

			begin := int(data.CustomersBeginIndicies[i])
			for j := 0; j < len(customer.Rentals); j++ {
				rental := &customer.Rentals[j]
				f.S("\t").S(rental.Movie.Title).S("\t").G32(data.RentalsCosts[begin+j]).Ln()
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

func PrepareDataForStatementISPC(ctx *context.Context, data unsafe.Pointer, customer *types.Customer) {
	customers := [...]types.Customer{*customer}
	PrepareDataForStatementsISPC(ctx, data, customers[:])
}

func StatementISPCWithPreparedData(ctx *context.Context, data unsafe.Pointer, customer *types.Customer) string {
	customers := [...]types.Customer{*customer}
	statements := StatementsISPCWithPreparedData(ctx, data, customers[:])
	return statements[0]
}
