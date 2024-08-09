#!/bin/sh

go tool 6g main.go
go tool 6a write.s
go tool pack c main.a main.6 write.6
go tool 6l -E main.main -s -w -o main main.a

rm *.6 main.a
