package json

import "../types"

var Invoices = []types.Invoice{
	{
		Customer: "BigCo",
		Performances: []types.Performance{
			{PlayID: "hamlet", Audience: 55},
			{PlayID: "as-like", Audience: 35},
			{PlayID: "othello", Audience: 40},
		},
	},
}
