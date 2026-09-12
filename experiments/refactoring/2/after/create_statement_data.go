package after

import "../types"

type EnrichedPerformance struct {
	types.Performance
	Play          types.Play
	Amount        float64
	VolumeCredits int
}

type StatementData struct {
	Customer           string
	Performances       []EnrichedPerformance
	TotalAmount        float64
	TotalVolumeCredits int
}

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func Reduce[T, U any](ts []T, f func(acc U, cur T) U, z U) U {
	result := z
	for _, t := range ts {
		result = f(result, t)
	}
	return result
}

func CreateStatementData(invoice *types.Invoice, plays types.Plays) StatementData {
	var result StatementData

	playFor := func(performance types.Performance) types.Play {
		return plays[performance.PlayID]
	}

	enrichPerformance := func(performance types.Performance) EnrichedPerformance {
		calculator := CreatePerformanceCalculator(performance, playFor(performance))
		return EnrichedPerformance{
			Performance:   performance,
			Play:          calculator.Play(),
			Amount:        calculator.Amount() / 100,
			VolumeCredits: calculator.VolumeCredits(),
		}
	}

	totalAmount := func(data StatementData) float64 {
		return Reduce(data.Performances, func(total float64, p EnrichedPerformance) float64 {
			return total + p.Amount
		}, 0)
	}

	totalVolumeCredits := func(data StatementData) int {
		return Reduce(data.Performances, func(total int, p EnrichedPerformance) int {
			return total + p.VolumeCredits
		}, 0)
	}

	result.Customer = invoice.Customer
	result.Performances = Map(invoice.Performances, enrichPerformance)
	result.TotalAmount = totalAmount(result)
	result.TotalVolumeCredits = totalVolumeCredits(result)

	return result
}

type PerformanceCalculator interface {
	Play() types.Play
	Amount() float64
	VolumeCredits() int
}

type GenericCalculator struct {
	performance types.Performance
	play        types.Play
}

type TragedyCalculator GenericCalculator

type ComedyCalculator GenericCalculator

/* NOTE(anton2920): sanity checks. */
var (
	_ PerformanceCalculator = &TragedyCalculator{}
	_ PerformanceCalculator = &ComedyCalculator{}
)

func CreatePerformanceCalculator(performance types.Performance, play types.Play) PerformanceCalculator {
	switch play.Type {
	case "tragedy":
		return &TragedyCalculator{performance, play}
	case "comedy":
		return &ComedyCalculator{performance, play}
	default:
		/* TODO(anton2920): ideally it's not a programmer's error, but in this example this will suffice. */
		panic("unknown play type")
	}
}

func (tc *TragedyCalculator) Play() types.Play {
	return tc.play
}

func (tc *TragedyCalculator) Amount() float64 {
	result := 40000
	if tc.performance.Audience > 30 {
		result += 1000 * (tc.performance.Audience - 30)
	}
	return float64(result)
}

func (tc *TragedyCalculator) VolumeCredits() int {
	return max(tc.performance.Audience-30, 0)
}

func (cc *ComedyCalculator) Play() types.Play {
	return cc.play
}

func (cc *ComedyCalculator) Amount() float64 {
	result := 30000
	if cc.performance.Audience > 20 {
		result += 10000 + 500*(cc.performance.Audience-20)
	}
	result += 300 * cc.performance.Audience
	return float64(result)
}

func (cc *ComedyCalculator) VolumeCredits() int {
	return max(cc.performance.Audience-30, 0) + cc.performance.Audience/5
}
