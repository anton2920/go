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

func StatementsISPC2(ctx *context.Context, customers []types.Customer) []string {
	t := trace_.Begin("")

	var (
		rentalCosts   []float32
		rentalCredits []uint16
		rentalDays    []uint16
		counts        []uint16

		offsets []uint16
		currs   []uint16
	)
	var (
		countsByCustomer []uint16
		totalCosts       []float32
		totalCredits     []uint16
	)

	{
		t := trace_.Begin("StatementsISPC2/PreparingData")

		counts = ctx.Arena.PushUint16Array(int(types.MovieTypeCount))
		for i := 0; i < len(customers); i++ {
			customer := &customers[i]
			rentals := customer.Rentals
			for j := 0; j < len(rentals); j++ {
				rental := &rentals[j]
				movie := &rental.Movie
				typ := movie.PriceCode
				counts[typ]++
			}
		}
		offsets = ctx.Arena.PushUint16Array(int(types.MovieTypeCount) + 1)
		currs = ctx.Arena.PushUint16Array(len(counts))

		var total uint16
		for i := 0; i < len(counts); i++ {
			total += counts[i]
			offsets[i+1] = total
		}

		rentalCosts = ctx.Arena.PushFloat32Array(int(total))
		rentalCredits = ctx.Arena.PushUint16Array(int(total))
		rentalDays = ctx.Arena.PushUint16Array(int(total))

		countsByCustomer = ctx.Arena.PushUint16Array(int(types.MovieTypeCount) * len(customers))
		for i := 0; i < len(customers); i++ {
			customer := &customers[i]
			rentals := customer.Rentals
			for j := 0; j < len(rentals); j++ {
				rental := &rentals[j]
				movie := &rental.Movie
				days := rental.DaysRented

				typ := movie.PriceCode
				offset := offsets[typ]
				curr := currs[typ]

				rentalDays[offset+curr] = uint16(days)

				currs[typ] = curr + 1
				countsByCustomer[i*int(types.MovieTypeCount)+int(typ)]++
			}
		}

		totalCosts = ctx.Arena.PushFloat32Array(len(customers))
		totalCredits = ctx.Arena.PushUint16Array(len(customers))

		trace_.End(t)
	}

	{
		t := trace_.Begin("StatementsISPC2/Calculation")

		cgo.Call1(C.CalculateStatements2, uintptr(unsafe.Pointer(&struct {
			OffsetsPtr *uint16
			CurrsPtr   *uint16

			RentalCostsPtr   *float32
			RentalCreditsPtr *uint16
			RentalDaysPtr    *uint16
			CountsPtr        *uint16

			TotalCostsPtr    *float32
			TotalCreditsPtr  *uint16
			CountsByCustomer *uint16
			CustomersCount   uint32
		}{&offsets[0], &currs[0], &rentalCosts[0], &rentalCredits[0], &rentalDays[0], &counts[0], &totalCosts[0], &totalCredits[0], &countsByCustomer[0], uint32(len(customers))})))

		trace_.End(t)
	}

	var statements []string
	{
		t := trace_.Begin("StatementsISPC2/Formatting")

		var f fmt.Formatter
		for i := 0; i < len(currs); i++ {
			currs[i] = 0
		}

		statements = ctx.Arena.PushStringArray(len(customers))
		for i := 0; i < len(customers); i++ {
			customer := &customers[i]

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), ctx.Arena.Size-int(ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(customer.Name).Ln()

			rentals := customer.Rentals
			for j := 0; j < len(rentals); j++ {
				rental := &rentals[j]
				movie := &rental.Movie

				typ := movie.PriceCode
				offset := offsets[typ]
				curr := currs[typ]

				cost := rentalCosts[offset+curr]

				f.S("\t").S(movie.Title).S("\t").G32(cost).Ln()

				currs[typ] = curr + 1
			}

			totalCost := totalCosts[i]
			totalCredits := totalCredits[i]

			f.S("Amount owed is ").G32(totalCost).Ln()
			f.S("You earned ").D(int(totalCredits)).S(" frequent renter points").Ln()

			statements[i] = f.String()
			ctx.Arena.PrevOfft = ctx.Arena.CurrOfft
			ctx.Arena.CurrOfft += uintptr(len(statements[i]))
		}

		trace_.End(t)
	}

	trace_.End(t)
	return statements
}

func StatementISPC2(ctx *context.Context, customer *types.Customer) string {
	customers := [...]types.Customer{*customer}
	statements := StatementsISPC2(ctx, customers[:])
	return statements[0]
}

type PreparedData2 struct {
	Offsets []uint16
	Currs   []uint16

	RentalCosts   []float32
	RentalCredits []uint16
	RentalDays    []uint16
	Counts        []uint16

	TotalCosts       []float32
	TotalCredits     []uint16
	CountsByCustomer []uint16
}

func PrepareDataForStatementsISPC2(ctx *context.Context, _data unsafe.Pointer, customers []types.Customer) {
	t := trace_.Begin("")

	data := (*PreparedData2)(_data)

	data.Counts = ctx.Arena.PushUint16Array(int(types.MovieTypeCount))
	for i := 0; i < len(customers); i++ {
		customer := &customers[i]
		rentals := customer.Rentals
		for j := 0; j < len(rentals); j++ {
			rental := &rentals[j]
			movie := &rental.Movie
			typ := movie.PriceCode
			data.Counts[typ]++
		}
	}
	data.Offsets = ctx.Arena.PushUint16Array(len(data.Counts) + 1)
	data.Currs = ctx.Arena.PushUint16Array(len(data.Counts))

	var total uint16
	for i := 0; i < len(data.Counts); i++ {
		count := data.Counts[i]
		total += count
		data.Offsets[i+1] = total
	}

	data.RentalCosts = ctx.Arena.PushFloat32Array(int(total))
	data.RentalCredits = ctx.Arena.PushUint16Array(int(total))
	data.RentalDays = ctx.Arena.PushUint16Array(int(total))

	data.CountsByCustomer = ctx.Arena.PushUint16Array(int(types.MovieTypeCount) * len(customers))
	for i := 0; i < len(customers); i++ {
		customer := &customers[i]
		rentals := customer.Rentals
		for j := 0; j < len(rentals); j++ {
			rental := &rentals[j]
			movie := &rental.Movie
			days := rental.DaysRented

			typ := movie.PriceCode
			offset := data.Offsets[typ]
			curr := data.Currs[typ]

			data.RentalDays[offset+curr] = uint16(days)

			data.Currs[typ]++
			data.CountsByCustomer[i*int(types.MovieTypeCount)+int(typ)]++
		}
	}

	data.TotalCosts = ctx.Arena.PushFloat32Array(len(customers))
	data.TotalCredits = ctx.Arena.PushUint16Array(len(customers))

	trace_.End(t)
}

func StatementsISPCWithPreparedData2(ctx *context.Context, _data unsafe.Pointer, customers []types.Customer) []string {
	t := trace_.Begin("")

	data := (*PreparedData2)(_data)

	{
		t := trace_.Begin("StatementsISPCWithPreparedData2/Calculation")

		cgo.Call1(C.CalculateStatements2, uintptr(unsafe.Pointer(&struct {
			OffsetsPtr *uint16
			CurrsPtr   *uint16

			RentalCostsPtr   *float32
			RentalCreditsPtr *uint16
			RentalDaysPtr    *uint16
			CountsPtr        *uint16

			TotalCostsPtr    *float32
			TotalCreditsPtr  *uint16
			CountsByCustomer *uint16
			CustomersCount   uint32
		}{&data.Offsets[0], &data.Currs[0], &data.RentalCosts[0], &data.RentalCredits[0], &data.RentalDays[0], &data.Counts[0], &data.TotalCosts[0], &data.TotalCredits[0], &data.CountsByCustomer[0], uint32(len(customers))})))

		trace_.End(t)
	}

	var statements []string
	{
		t := trace_.Begin("StatementsISPCWithPreparedData2/Formatting")

		var f fmt.Formatter
		for i := 0; i < len(data.Currs); i++ {
			data.Currs[i] = 0
		}

		statements = ctx.Arena.PushStringArray(len(customers))
		for i := 0; i < len(customers); i++ {
			customer := &customers[i]

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), ctx.Arena.Size-int(ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(customer.Name).Ln()

			rentals := customer.Rentals
			for j := 0; j < len(rentals); j++ {
				rental := &rentals[j]
				movie := &rental.Movie

				typ := movie.PriceCode
				offset := data.Offsets[typ]
				curr := data.Currs[typ]

				cost := data.RentalCosts[offset+curr]

				f.S("\t").S(movie.Title).S("\t").G32(cost).Ln()

				data.Currs[typ]++
			}

			totalCost := data.TotalCosts[i]
			totalCredits := data.TotalCredits[i]

			f.S("Amount owed is ").G32(totalCost).Ln()
			f.S("You earned ").D(int(totalCredits)).S(" frequent renter points").Ln()

			statements[i] = f.String()
			ctx.Arena.PrevOfft = ctx.Arena.CurrOfft
			ctx.Arena.CurrOfft += uintptr(len(statements[i]))
		}

		trace_.End(t)
	}

	trace_.End(t)
	return statements
}

func PrepareDataForStatementISPC2(ctx *context.Context, data unsafe.Pointer, customer *types.Customer) {
	customers := [...]types.Customer{*customer}
	PrepareDataForStatementsISPC2(ctx, data, customers[:])
}

func StatementISPCWithPreparedData2(ctx *context.Context, data unsafe.Pointer, customer *types.Customer) string {
	customers := [...]types.Customer{*customer}
	statements := StatementsISPCWithPreparedData2(ctx, data, customers[:])
	return statements[0]
}
