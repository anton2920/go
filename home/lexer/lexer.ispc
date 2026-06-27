//ispc:flags: --target=avx2-i8x32 -g

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

/* TODO(anton2920): define them properly on i386. */
typedef int64  goint;
typedef uint64 gouint;

struct gostring {
	uint8* Data;
	goint  Len;
};

struct goslice {
	uint8* Data;
	goint  Len;
	goint  Cap;
};

export void Nop() {}

static
void
CrashUnless(uni bool cond)
{
	u64 *uni p = NULL;
	if (!cond) {
		*p = 0x0;
	}
}
/* --- */

typedef uint8 Token;

struct Lexer {
	gostring Filename;
	goslice  Contents;

	goint ErrorCount;
	goint CallCount;

	goint rdOfft;
	goint offt;
	goint lineOfft;

	var uint8 ch;
};


inline
bool
isspace(uint8 c)
{
	return (c == 32) || (c == 9) || (c == 10) || (c == 13);
}

#define LEXER_DEBUG

#ifdef LEXER_DEBUG
#define lexerError(l, msg) { l->ErrorCount++; print(msg); }
#else
#define lexerError(l, msg) l->ErrorCount++;
#endif


static
void
lexerNext(Lexer *uni l)
{
	l->CallCount++;

	if (l->rdOfft >= l->Contents.Len) {
		l->ch = 128;
		return;
	}

	l->offt = l->rdOfft;
	uint8 r = l->Contents.Data[l->rdOfft];
	uni goint w = sizeof r;

	if (r == 0) {
		lexerError(l, "invalid character NUL");
	} else if (r >= 0x80) {
		lexerError(l, "multibyte characters currently are not supported");
	}

	l->rdOfft += w;
	l->ch = r;
}


static
void
lexerSkipSpaces(Lexer *uni l)
{
	while (isspace(l->ch)) {
		l->ch = l->Contents.Data[l->rdOfft];
		if (isspace(l->ch)) {
			l->rdOfft++;
		}
	}
}


export
uni bool
#pragma ignore warning(all)
LexerFindTokens(Lexer *uni l, goslice *uni begins, goslice *uni ends)
{
	uni i32 begin = 0, end = 0;
	uni u64 mask;
	var u8 ch;

	/* TODO(anton2920): this goes out of bounds. */
	while (begin < l->Contents.Len) {
		ch = l->Contents.Data[begin+programIndex];

		while (!(mask = packmask(isspace(ch)))) {
			end += programCount;
			ch = l->Contents.Data[end+programIndex];
		}

		maskWhile: while (mask) {
			uni i32 tz = count_trailing_zeros(mask);
			end += tz;
			mask >>= tz + 1;
			if (end <= begin) {
				begin = ++end;
				goto maskWhile;
			}

			CrashUnless((begins->Len + 1 < begins->Cap) && (ends->Len + 1 < ends->Cap));
			((uni i32 *uni)begins->Data)[begins->Len++] = begin;
			((uni i32 *uni)ends->Data)[ends->Len++] = end;

			begin = ++end;
		}
	}

	return false;
}
