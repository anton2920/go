package before

import (
	"fmt"

	"../types"

	"github.com/anton2920/gofa/ints"
)

func Statement(invoice *types.Invoice, plays types.Plays) string {
	var totalAmount float64
	var volumeCredits int

	var result = fmt.Sprintf("Statement for %s\n", invoice.Customer)

	for _, perf := range invoice.Performances {
		play := plays[perf.PlayID]
		var thisAmount int

		switch play.Type {
		case "tragedy":
			thisAmount = 40000
			if perf.Audience > 30 {
				thisAmount += 1000 * (perf.Audience - 30)
			}
		case "comedy":
			thisAmount = 30000
			if perf.Audience > 20 {
				thisAmount += 10000 + 500*(perf.Audience-20)
			}
			thisAmount += 300 * perf.Audience
		default:
			/* TODO(anton2920): ideally it's not a programmer's error, but in this example this will suffice. */
			panic("unknown play type " + play.Type)
		}

		volumeCredits += ints.Max(perf.Audience-30, 0)
		if play.Type == "comedy" {
			volumeCredits += perf.Audience / 5
		}

		result += fmt.Sprintf("\t%s: $%.2f (%d seats)\n", play.Name, float64(thisAmount)/100, perf.Audience)
		totalAmount += float64(thisAmount)
	}

	result += fmt.Sprintf("Amount owed is $%.2f\n", totalAmount/100)
	result += fmt.Sprintf("You earned %d credits\n", volumeCredits)

	return result
}
