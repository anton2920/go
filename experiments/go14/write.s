#define	SYS_exit	1
#define	SYS_write	4

TEXT ·Exit(SB), $0-4
	MOVQ	$SYS_exit, AX
	MOVL	status+0(FP), DI
	SYSCALL
	RET

TEXT ·Write(SB), $0-32
	MOVQ	$SYS_write, AX
	MOVL	fd+0(FP), DI
	MOVQ	buf+8(FP), SI
	MOVQ	buf_len+16(FP), DX
	SYSCALL
	MOVQ	AX, ret+24(FP)
	RET
