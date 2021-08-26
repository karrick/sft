build: sft

sft:
	rm -f output.go
	go build

bench: sft
	./$< -o output.go '%a, %d %b %Y %T %z'
	go run output.go
	go test -bench=. -benchmem output*.go
