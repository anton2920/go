#!/bin/sh

PROJECT=lab_04

VERBOSITY=0
VERBOSITYFLAGS=""
while test "$1" = "-v"; do
	VERBOSITY=$((VERBOSITY+1))
	VERBOSITYFLAGS="$VERBOSITYFLAGS -v"
	shift
done

run()
{
	if test $VERBOSITY -gt 1; then echo "$@"; fi
	"$@" || exit 1
}

printv()
{
	if test $VERBOSITY -gt 0; then echo "$@"; fi
}

# NOTE(anton2920): disable Go 1.11+ package management.
GO111MODULE=off; export GO111MODULE

CGO_ENABLED=0; export CGO_ENABLED

STARTTIME=`date +%s`

case $1 in
	'' | debug)
		CGO_ENABLED=1; export CGO_ENABLED
		run go build -o $PROJECT -race -pgo off -gcflags='all=-N -l -d=checkptr=0'
		;;
	check)
		CGO_ENABLED=1; export CGO_ENABLED
		run go test -race .
		;;
	check-bench)
		run go test -bench=. -run=^Benchmark .
		;;
	clean)
		run rm -f $PROJECT $PROJECT.s $PROJECT.esc $PROJECT.test c.out cpu.pprof cpu.png mem.pprof mem.png
		run go clean -cache -modcache -testcache
		run rm -rf `go env GOCACHE`
		run rm -rf /tmp/cover*
		;;
	disas | disasm | disassembly)
		printv go build -pgo off -gcflags="-S"
		go build -gcflags="-S" >$PROJECT.disas 2>&1
		;;
	esc | escape | escape-analysis)
		printv go build -pgo off -gcflags="-m -m"
		go build -gcflags="-m -m" >$PROJECT.m 2>&1
		;;
	fmt)
		if which goimports >/dev/null; then
			run goimports -l -w *.go
		else
			run gofmt -l -s -w *.go
		fi
		;;
	objdump)
		go build -o $PROJECT -pgo off
		printvv go tool objdump -S -s ^main\. $PROJECT
		go tool objdump -S -s ^main\. $PROJECT >$PROJECT.s
		;;
	release)
		run go build -o $PROJECT -ldflags="-s -w"
		;;
	vet)
		run go vet -asmdecl -assign -atomic -bools -buildtag -cgocall -composites -copylocks -directive -errorsas -framepointer -httpresponse -ifaceassert -loopclosure -lostcancel -nilfunc -printf -shift -sigchanyzer -slog -stdmethods -stringintconv -structtag -testinggoroutine -tests -timeformat -unmarshal -unreachable -unusedresult
		;;
	*)
		echo "Target $1 is not supported"
		exit 1
		;;
esac

ENDTIME=`date +%s`

echo Done $1 in $((ENDTIME-STARTTIME))s
