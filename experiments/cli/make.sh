#!/bin/sh

PATH=$HOME/go14/bin:$PATH
export PATH

GOPATH=$HOME/go
export GOPATH

go build -o main main_raw.go converter.go
