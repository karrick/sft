package main

import (
	"testing"
	"time"
)

// time.RFC1123Z:    "%a, %d %b %Y %T %z", // "Mon, 02 Jan 2006 15:04:05 -0700",

func foo() {
	// fmt.Println(string(appendTime(nil, time.Now().Add(8*time.Hour))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 0, 0, 0, 1, time.UTC))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 11, 59, 59, 0, time.UTC))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 12, 00, 00, 0, time.UTC))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 12, 00, 01, 0, time.UTC))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 13, 00, 00, 0, time.UTC))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 23, 59, 59, 0, time.UTC))))
	// fmt.Println(string(appendTime(nil, time.Date(2006, time.January, 2, 24, 00, 00, 0, time.UTC))))
}

func BenchmarkAppendTime(b *testing.B) {
	foo := make([]byte, 0, 64)
	now := time.Now()
	for i := 0; i < b.N; i++ {
		foo = appendTime(nil, now)
		foo = foo[:0]
	}
	_ = foo
}

func BenchmarkStandardLibrary(b *testing.B) {
	var foo string
	now := time.Now()
	for i := 0; i < b.N; i++ {
		foo = now.Format(time.RFC1123Z)
	}
	_ = foo
}
