#!/bin/sh

go tool compile -p main main.go
go tool asm -p main write.s
go tool pack c main.a main.o write.o
go tool link -E main.main -linkmode external -s -w -o main main.a

# rm *.o main.a
