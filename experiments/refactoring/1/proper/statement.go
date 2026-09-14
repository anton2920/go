package proper

import (
	"unsafe"

	"../types"

	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/fmt"
	gotypes "github.com/anton2920/gofa/go/types"
	"github.com/anton2920/gofa/mem"
	"github.com/anton2920/gofa/pointers"
	"github.com/anton2920/gofa/trace/trace_"
)

var (
	Inits = [...]float32{
		types.MovieTypeRegular:    2,
		types.MovieTypeNewRelease: 0,
		types.MovieTypeChildrens:  1.5,
	}
	Cutoffs = [...]float32{
		types.MovieTypeRegular:    2,
		types.MovieTypeNewRelease: 0,
		types.MovieTypeChildrens:  3,
	}
	Scales = [...]float32{
		types.MovieTypeRegular:    1.5,
		types.MovieTypeNewRelease: 3,
		types.MovieTypeChildrens:  1.5,
	}
)

type Total struct {
	Cost    float32
	Credits uint32
}

func ArenaPushTotalArray(a *mem.Arena, n int) []Total {
	ptr := a.PushSizeWithAlignment(unsafe.Sizeof(Total{})*uintptr(n), unsafe.Alignof(Total{}))
	return *(*[]Total)(unsafe.Pointer(&gotypes.SliceHeader{Data: uintptr(ptr), Len: n, Cap: n}))
}

func Statements(ctx *context.Context, customers []types.Customer) []string {
	t := trace_.Begin("")

	var (
		rentalCosts []float32
		totals      []Total
	)

	{
		t := trace_.Begin("proper.Statements/PreparingData")

		var totalRentals int
		for i := 0; i < len(customers); i++ {
			customer := &customers[i]
			rentals := customer.Rentals
			totalRentals += len(rentals)
		}

		rentalCosts = ctx.Arena.PushFloat32Array(totalRentals)
		totals = ArenaPushTotalArray(&ctx.Arena, len(customers))

		trace_.End(t)
	}

	{
		t := trace_.Begin("proper.Statements/Calculating")

		var curr int
		for i := 0; i < len(customers); i++ {
			var totalCost float32
			var totalCredits uint32

			customer := &customers[i]
			rentals := customer.Rentals

			for j := 0; j < len(rentals); j++ {
				rental := &rentals[j]
				movie := &rental.Movie

				days := rental.DaysRented
				typ := movie.PriceCode

				init := Inits[typ]
				cutoff := Cutoffs[typ]
				scale := Scales[typ]

				cost := init
				fdays := float32(days)
				if fdays > cutoff {
					cost += (fdays - cutoff) * scale
				}

				credits := 1
				if (typ == types.MovieTypeNewRelease) && (days > 1) {
					credits++
				}

				totalCost += cost
				totalCredits += uint32(credits)

				rentalCosts[curr] = cost
				curr++
			}

			totals[i].Cost = totalCost
			totals[i].Credits = totalCredits
		}

		trace_.End(t)
	}

	var statements []string
	{
		t := trace_.Begin("proper.Statements/Formatting")

		statements = ctx.Arena.PushStringArray(len(customers))

		var f fmt.Formatter
		var curr int

		for i := 0; i < len(customers); i++ {
			customer := &customers[i]
			rentals := customer.Rentals
			name := customer.Name

			f.InitWithUnsafePointer(pointers.Add(ctx.Arena.Base, ctx.Arena.CurrOfft), int(uintptr(ctx.Arena.Size)-ctx.Arena.CurrOfft))
			f.S("Rental Record for ").S(name).Ln()

			for j := 0; j < len(rentals); j++ {
				rental := &rentals[j]
				movie := &rental.Movie
				title := movie.Title

				cost := rentalCosts[curr]
				curr++

				f.S("\t").S(title).S("\t").G32(cost).Ln()
			}

			total := &totals[i]
			totalCost := total.Cost
			totalCredit := total.Credits

			f.S("Amount owed is ").G32(totalCost).Ln()
			f.S("You earned ").D(int(totalCredit)).S(" frequent renter points").Ln()

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
