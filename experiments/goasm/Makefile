GO = gccgo
AS = as

all: main

main: main.go number.o
	$(GO) main.go number.o -o main

number.o: number.s
	$(AS) number.s -o number.o

clean:
	rm -f *.o main
