package main

import (
	"testing"
	"time"
)

// time.RFC1123Z:    "%a, %d %b %Y %T %z", // "Mon, 02 Jan 2006 15:04:05 -0700",
var when = time.Date(2006, time.January, 2, 3, 4, 5, 12345678, time.UTC)

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
	buf := make([]byte, 0, 512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = appendTime(buf, when)
		buf = buf[:0] // clear contents of byte slice
	}
	_ = buf
}

func BenchmarkWriteTime(b *testing.B) {
	var silly []byte
	buf := make([]byte, 512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		silly = writeTime(buf, when)
		// buf = buf[:512] // reslice byte slice
	}
	_ = silly
}

func BenchmarkStandardLibrary(b *testing.B) {
	// b.Skip("invalid measurement unless using same time format")
	var foo string
	for i := 0; i < b.N; i++ {
		foo = when.Format(time.RFC1123Z)
	}
	_ = foo
}
