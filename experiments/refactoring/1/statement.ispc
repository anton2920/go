/* --- */
#define uni uniform
#define var varying

typedef int8 i8;
typedef uint8 u8;
typedef int16 i16;
typedef uint16 u16;
typedef int32 i32;
typedef uint32 u32;
typedef int64 i64;
typedef uint64 u64;
typedef float f32;
typedef double f64;
/* --- */

struct Params {
	u32 *CustomersBeginIndicies;
	f32 *AmountsOwed;
	u16 *RentersPoints;

	i32 *RentalsDays;
	u8 *RentalsTypes;

	u32 *RentalsCostsX100;
	u16 *RentalsCredits;

	u32 NCustomers;
	u32 NRentals;
};

export
void
CalculateStatements(Params *uni params)
{
	const uni u32 inits[programCount] = {200, 0, 150};
	const uni i32 cutoffs[programCount] = {20, 0, 30};
	const uni u32 scales[programCount] = {15, 30, 15};

	foreach (i = 0 ... params->NRentals) {
		i32 days = params->RentalsDays[i];
		u8 typ = params->RentalsTypes[i];

		#pragma ignore warning(perf)
		u32 init = inits[typ];
		#pragma ignore warning(perf)
		i32 cutoff = cutoffs[typ];
		#pragma ignore warning(perf)
		u32 scale = scales[typ];

		params->RentalsCostsX100[i] = init + (u32)max((var i32)0, days*10 - cutoff) * scale;
		params->RentalsCredits[i] = 1 + (u16)((params->RentalsTypes[i] == 1) && (days > 1));
	}

	for (uni i32 cidx = 0; cidx < params->NCustomers; cidx++) {
		u32 amount = 0;
		u16 credits = 0;
		foreach (i = params->CustomersBeginIndicies[cidx] ... params->CustomersBeginIndicies[cidx+1]) {
			amount += params->RentalsCostsX100[i];
			credits += params->RentalsCredits[i];
		}
		params->AmountsOwed[cidx] = (uni f32)reduce_add(amount) / 100.f;
		params->RentersPoints[cidx] = reduce_add(credits);
	}
}
