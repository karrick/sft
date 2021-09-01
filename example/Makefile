build: example

clean:
	rm -f example formatTime.go

example: main.go formatTime.go
	go build -o $@ $^

formatTime.go:
	go generate
