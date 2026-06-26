/* --- */
#define uni uniform
#define var varying

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
uni Token
#pragma ignore warning(all)
LexerNextToken(Lexer *uni l)
{
	lexerNext(l);

	lexerSkipSpaces(l);
	switch (l->ch) {
	default:
		lexerNext(l);
		break;
	}

	return extract(l->ch, 0);
}
