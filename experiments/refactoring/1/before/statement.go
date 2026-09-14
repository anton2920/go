package before

import (
	"bytes"
	"strconv"

	"../types"

	"github.com/anton2920/gofa/context"
)

func Ftoa(x float64) string {
	return strconv.FormatFloat(x, 'g', -1, 64)
}

func Statement(_ *context.Context, customer *types.Customer) string {
	var frequentRenterPoints int
	var totalAmount float64

	result := "Rental Record for " + customer.Name + "\n"

	rentals := customer.Rentals
	for _, each := range rentals {
		var thisAmount float64

		switch each.Movie.PriceCode {
		case types.MovieTypeRegular:
			thisAmount += 2
			if each.DaysRented > 2 {
				thisAmount += float64(each.DaysRented-2) * 1.5
			}
		case types.MovieTypeNewRelease:
			thisAmount += float64(each.DaysRented * 3)
		case types.MovieTypeChildrens:
			thisAmount += 1.5
			if each.DaysRented > 3 {
				thisAmount += float64(each.DaysRented-3) * 1.5
			}
		}

		frequentRenterPoints++
		if (each.Movie.PriceCode == types.MovieTypeNewRelease) && (each.DaysRented > 1) {
			frequentRenterPoints++
		}

		result += "\t" + each.Movie.Title + "\t" + Ftoa(thisAmount) + "\n"
		totalAmount += thisAmount
	}

	result += "Amount owed is " + Ftoa(totalAmount) + "\n"
	result += "You earned " + strconv.Itoa(frequentRenterPoints) + " frequent renter points\n"
	return result
}

func StatementBytesBuffer(_ *context.Context, customer *types.Customer) string {
	var frequentRenterPoints int
	var totalAmount float64

	var result bytes.Buffer
	result.WriteString("Rental Record for ")
	result.WriteString(customer.Name)
	result.WriteByte('\n')

	rentals := customer.Rentals
	for _, each := range rentals {
		var thisAmount float64

		switch each.Movie.PriceCode {
		case types.MovieTypeRegular:
			thisAmount += 2
			if each.DaysRented > 2 {
				thisAmount += float64(each.DaysRented-2) * 1.5
			}
		case types.MovieTypeNewRelease:
			thisAmount += float64(each.DaysRented * 3)
		case types.MovieTypeChildrens:
			thisAmount += 1.5
			if each.DaysRented > 3 {
				thisAmount += float64(each.DaysRented-3) * 1.5
			}
		}

		frequentRenterPoints++
		if (each.Movie.PriceCode == types.MovieTypeNewRelease) && (each.DaysRented > 1) {
			frequentRenterPoints++
		}

		result.WriteByte('\t')
		result.WriteString(each.Movie.Title)
		result.WriteByte('\t')
		result.WriteString(Ftoa(thisAmount))
		result.WriteByte('\n')

		totalAmount += thisAmount
	}

	result.WriteString("Amount owed is ")
	result.WriteString(Ftoa(totalAmount))
	result.WriteByte('\n')

	result.WriteString("You earned ")
	result.WriteString(strconv.Itoa(frequentRenterPoints))
	result.WriteString(" frequent renter points\n")

	return result.String()
}

func StatementMyFmt(ctx *context.Context, customer *types.Customer) string {
	var frequentRenterPoints int
	var totalAmount float64

	f := &ctx.Fmt
	f.Reset().S("Rental Record for ").S(customer.Name).Ln()

	rentals := customer.Rentals
	for _, each := range rentals {
		var thisAmount float64

		switch each.Movie.PriceCode {
		case types.MovieTypeRegular:
			thisAmount += 2
			if each.DaysRented > 2 {
				thisAmount += float64(each.DaysRented-2) * 1.5
			}
		case types.MovieTypeNewRelease:
			thisAmount += float64(each.DaysRented * 3)
		case types.MovieTypeChildrens:
			thisAmount += 1.5
			if each.DaysRented > 3 {
				thisAmount += float64(each.DaysRented-3) * 1.5
			}
		}

		frequentRenterPoints++
		if (each.Movie.PriceCode == types.MovieTypeNewRelease) && (each.DaysRented > 1) {
			frequentRenterPoints++
		}

		f.S("\t").S(each.Movie.Title).S("\t").G(thisAmount).Ln()
		totalAmount += thisAmount
	}

	f.S("Amount owed is ").G(totalAmount).Ln()
	f.S("You earned ").D(frequentRenterPoints).S(" frequent renter points").Ln()
	return f.String()
}

func Statements(ctx *context.Context, customers []types.Customer) []string {
	statements := ctx.Arena.PushStringArray(len(customers))
	for i := 0; i < len(customers); i++ {
		statements[i] = Statement(ctx, &customers[i])
	}
	return statements
}

func StatementsBytesBuffer(ctx *context.Context, customers []types.Customer) []string {
	statements := ctx.Arena.PushStringArray(len(customers))
	for i := 0; i < len(customers); i++ {
		statements[i] = StatementBytesBuffer(ctx, &customers[i])
	}
	return statements
}

func StatementsMyFmt(ctx *context.Context, customers []types.Customer) []string {
	statements := ctx.Arena.PushStringArray(len(customers))
	for i := 0; i < len(customers); i++ {
		statements[i] = StatementMyFmt(ctx, &customers[i])
	}
	return statements
}
