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


/* NOTE(anton2920): must be in sync with './types/movie.go'. */
enum MovieType {
	MovieTypeRegular,
	MovieTypeNewRelease,
	MovieTypeChildrens,
	MovieTypeCount,
};

struct Params {
	u32 *CustomersBeginIndicies;
	f32 *AmountsOwed;
	u16 *RentersPoints;

	i32 *RentalDays;
	u8 *RentalTypes;

	f32 *RentalCosts;
	u16 *RentalCredits;

	u32 NCustomers;
	u32 NRentals;
};


static const uni f32 Inits[MovieTypeCount]   = {2.0f, 0.0f, 1.5f};
static const uni f32 Cutoffs[MovieTypeCount] = {2.0f, 0.0f, 3.0f};
static const uni f32 Scales[MovieTypeCount]  = {1.5f, 3.0f, 1.5f};


export
void
CalculateStatements(Params *uni params)
{
	foreach (i = 0 ... params->NRentals) {
		i32 days = params->RentalDays[i];
		u8  typ  = params->RentalTypes[i];

		#pragma ignore warning(perf)
		f32 init   = Inits[typ];
		#pragma ignore warning(perf)
		f32 cutoff = Cutoffs[typ];
		#pragma ignore warning(perf)
		f32 scale  = Scales[typ];

		params->RentalCosts[i] = init + max(0.f, days - cutoff) * scale;
		params->RentalCredits[i] = 1 + (u16)((typ == MovieTypeNewRelease) && (days > 1));
	}

	for (uni i32 cidx = 0; cidx < params->NCustomers; cidx++) {
		uni u32 begin = params->CustomersBeginIndicies[cidx];
		uni u32 end = params->CustomersBeginIndicies[cidx+1];

		f32 amount = 0;
		u16 credits = 0;

		foreach (i = begin ... end) {
			amount += params->RentalCosts[i];
			credits += params->RentalCredits[i];
		}

		params->AmountsOwed[cidx] = reduce_add(amount);
		params->RentersPoints[cidx] = reduce_add(credits);
	}
}


struct Params2 {
	u16 *Offsets;
	u16 *Currs;

	f32 *RentalCosts;
	u16 *RentalCredits;
	u16 *RentalDays;
	u16 *Counts;

	f32 *TotalCosts;
	u16 *TotalCredits;
	u16 *CountsByCustomer;
	u32  CustomersCount;
};


template <u8 typ>
static inline
void
CalculateMovieRentalCostAndCredits(uni f32 costs[], uni u16 credits[], uni u16 rentalDays[], uni u16 n)
{
	uni f32 init   = Inits[typ];
	uni f32 cutoff = Cutoffs[typ];
	uni f32 scale  = Scales[typ];

	foreach (i = 0 ... n) {
		u16 days = rentalDays[i];

		f32 cost = init;
		if (days > cutoff) {
			cost += (days - cutoff) * scale;
		}
		costs[i] = cost;

		credits[i] = 1 + (u16)((typ == MovieTypeNewRelease) && (days > 1));
	}
}


export
void
CalculateStatements2(Params2 *uni params)
{
	{
		f32 *uni costs   = params->RentalCosts;
		u16 *uni credits = params->RentalCredits;
		u16 *uni days    = params->RentalDays;
		u16 *uni counts  = params->Counts;
		uni  u16 start, count;

		start = 0; count = counts[MovieTypeRegular];
		CalculateMovieRentalCostAndCredits<MovieTypeRegular>(&costs[start], &credits[start], &days[start], count);

		start += count; count = counts[MovieTypeNewRelease];
		CalculateMovieRentalCostAndCredits<MovieTypeNewRelease>(&costs[start], &credits[start], &days[start], count);

		start += count; count = counts[MovieTypeChildrens];
		CalculateMovieRentalCostAndCredits<MovieTypeChildrens>(&costs[start], &credits[start], &days[start], count);
	}

	{
		u16 *uni offsets = params->Offsets;

		f32 *uni totalCostsArray = params->TotalCosts;
		f32 *uni costs   = params->RentalCosts;

		u16 *uni totalCreditsArray = params->TotalCredits;
		u16 *uni credits = params->RentalCredits;

		u16 *uni currs = params->Currs;
		memset(currs, 0, sizeof(uni u16) * MovieTypeCount);

		uni u16 n = params->CustomersCount;
		for (uni int i = 0; i < n; i++) {
			f32 totalCost = 0;
			u16 totalCredits = 0;

			for (uni int typ = 0; typ < MovieTypeCount; typ++) {
				uni u16 offset = offsets[typ];
				uni u16 count  = params->CountsByCustomer[i*MovieTypeCount + typ];
				uni u16 curr   = currs[typ];

				foreach (k = 0 ... count) {
					totalCost += costs[offset+curr+k];
					totalCredits += credits[offset+curr+k];
				}

				currs[typ] += count;
			}

			totalCostsArray[i] = reduce_add(totalCost);
			totalCreditsArray[i] = reduce_add(totalCredits);
		}
	}
}


