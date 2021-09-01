# sft

Go code generator for converting a `time.Time` value to a string
representation.

## Description

This program is designed to be run from the command line with a time
formatting string, and it will emit Go source code that formats
`time.Time` values according to the specified formatting string.

```Bash
$ sft -m -o formatTime.go -f formatTime '%F %T'
$ go run formatTime.go
$ less formatTime.go
```

The resultant code can be copied and pasted into another Go source
file, or the file can simply be compiled into another project. The
name of the function it creates and the name of the package it uses
can be changed on the command line. Removing the `-m` command line
skips creating a `main` function, and outputs a single function with
the specified function and package name.

The program could also be invoked from a Go generate statement in
other Go source code.

```Go
//go:generate sft -f formatTime -o formatTime.go '%F %T'
```

Then when the time format spec changes, simply type `go generate` at
the command line to regenerate the time formatting function.

## Performance

It is a bit faster than the Go standard library time formatting
functionality.

```Bash
$ karrick@promethium sft % make clean bench
rm -f append copy sft append.go copy.go append_test.go copy_test.go
go build -o sft main.go
./sft -extra -f appendTime -append -o append_test.go RFC3339Nano
./sft -extra -f copyTime -o copy_test.go RFC3339Nano
go test -bench=. -benchmem main_test.go append_test.go copy_test.go
goos: darwin
goarch: amd64
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
BenchmarkAppendTime-16         	19457082	        55.87 ns/op	       0 B/op	       0 allocs/op
BenchmarkCopyTime-16           	21264758	        54.54 ns/op	       0 B/op	       0 allocs/op
BenchmarkStandardLibrary-16    	 5772063	       203.4 ns/op	      32 B/op	       1 allocs/op
PASS
ok  	command-line-arguments	3.764s
```
