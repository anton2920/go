#!/bin/sh

set -e

PATH=$HOME/go14/bin:$PATH; export PATH
GOPATH=$HOME/go; export GOPATH
CGO_ENABLED=0; export CGO_ENABLED

go build -o main main.go
