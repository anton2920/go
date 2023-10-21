package main

import "fmt"

type PlayType int

const (
	Tragedy PlayType = iota
	Comedy
)

type Play struct {
	Name string
	Type PlayType
}

type Plays map[string]Play

type Performance struct {
	PlayID   string
	Audience int
}

type Invoice struct {
	Customer     string
	Performances []Performance
}

func Statement(invoice Invoice, plays Plays) string {
	var totalAmount float64
	var volumeCredits int

	var result = fmt.Sprintf("Statement for %s\n", invoice.Customer)

	for _, perf := range invoice.Performances {
		play := plays[perf.PlayID]
		var thisAmount int

		switch play.Type {
		case Tragedy:
			thisAmount = 40000
			if perf.Audience > 30 {
				thisAmount += 1000 * (perf.Audience - 30)
			}
		case Comedy:
			thisAmount = 30000
			if perf.Audience > 20 {
				thisAmount += 10000 + 500*(perf.Audience-20)
			}
			thisAmount += 300 * perf.Audience
		default:
			/* TODO(anton2920): ideally it's not a programmer's error, but in this example this will suffice. */
			panic("unknown play type")
		}

		volumeCredits += max(perf.Audience-30, 0)
		if play.Type == Comedy {
			volumeCredits += perf.Audience / 5
		}

		result += fmt.Sprintf("\t%s: $%.2f (%d seats)\n", play.Name, float64(thisAmount)/100, perf.Audience)
		totalAmount += float64(thisAmount)
	}

	result += fmt.Sprintf("Amount owed is $%.2f\n", totalAmount/100)
	result += fmt.Sprintf("You earned %d credits\n", volumeCredits)

	return result
}

func main() {
}
