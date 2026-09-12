package types

type MovieType int

const (
	MovieTypeRegular = MovieType(iota)
	MovieTypeNewRelease
	MovieTypeChildrens
	MovieTypeCount
)

type Movie struct {
	Title     string
	PriceCode MovieType
}
