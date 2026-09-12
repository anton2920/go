package utils

import (
	"math/rand"

	"../types"

	"github.com/anton2920/gofa/ints"
)

func GenerateRandomCustomer(seed int64, maxRentals int, maxDays int) *types.Customer {
	var random types.Customer

	rng := rand.New(rand.NewSource(seed))
	random.Name = "Random"

	random.Rentals = make([]types.Rental, ints.Max(1, rng.Intn(maxRentals)))
	for i := 0; i < len(random.Rentals); i++ {
		random.Rentals[i] = types.Rental{
			Movie:      types.Movie{Title: "Random", PriceCode: types.MovieType(rng.Intn(int(types.MovieTypeCount)))},
			DaysRented: ints.Max(1, rng.Intn(maxDays)),
		}
	}

	return &random
}

func GenerateRandomCustomers(seed int64, maxCustomers int, maxRentals int, maxDays int) []types.Customer {
	rng := rand.New(rand.NewSource(seed))

	randoms := make([]types.Customer, ints.Max(1, rng.Intn(maxCustomers)))
	for i := 0; i < len(randoms); i++ {
		random := &randoms[i]
		random.Name = "Random"
		random.Rentals = make([]types.Rental, ints.Max(1, maxRentals))
		for j := 0; j < len(random.Rentals); j++ {
			random.Rentals[j] = types.Rental{
				Movie:      types.Movie{Title: "Random", PriceCode: types.MovieType(rng.Intn(int(types.MovieTypeCount)))},
				DaysRented: rng.Intn(ints.Max(1, maxDays)),
			}
		}
	}

	return randoms
}
