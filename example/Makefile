build: example

clean:
	rm -f example formatTime.go

example: main.go formatTime.go
	go build

formatTime.go:
	go generate
