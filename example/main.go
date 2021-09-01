package main

//go:generate ../sft -f formatTime -o formatTime.go "%F %T"

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(string(formatTime(nil, time.Now().UTC())))
}
