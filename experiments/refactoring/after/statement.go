package main

import "fmt"

func Statement(invoice Invoice, plays Plays) string {
	return RenderPlainText(CreateStatementData(invoice, plays))
}

func RenderPlainText(data StatementData) string {
	var result = fmt.Sprintf("Statement for %s\n", data.Customer)
	for _, perf := range data.Performances {
		result += fmt.Sprintf("\t%s: $%.2f (%d seats)\n", perf.Play.Name, perf.Amount, perf.Audience)
	}
	result += fmt.Sprintf("Amount owed is $%.2f\n", data.TotalAmount)
	result += fmt.Sprintf("You earned %d credits\n", data.TotalVolumeCredits)
	return result
}

func HTMLStatement(invoice Invoice, plays Plays) string {
	return RenderHTML(CreateStatementData(invoice, plays))
}

func RenderHTML(data StatementData) string {
	return ""
}
