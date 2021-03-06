package main

import (
	"testing"
	"time"
)

var when = time.Date(2006, time.January, 2, 3, 4, 5, 12345678, time.UTC)

// var format = time.RFC1123Z      // "%a, %d %b %Y %T %z", // "Mon, 02 Jan 2006 15:04:05 -0700"
var format = time.RFC3339Nano // "%Y-%m-%dT%T.%N%1",   // "2006-01-02T15:04:05.999999999Z07:00"

func TestAppendTime(t *testing.T) {
	got := string(appendTime(nil, when))
	want := when.Format(format)
	if got != want {
		t.Errorf("GOT: %q; WANT: %q", got, want)
	}
}

func TestCopyTime(t *testing.T) {
	got := string(copyTime(nil, when))
	want := when.Format(format)
	if got != want {
		t.Errorf("GOT: %q; WANT: %q", got, want)
	}
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

func BenchmarkCopyTime(b *testing.B) {
	var silly []byte
	buf := make([]byte, 512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		silly = copyTime(buf, when)
		// buf = buf[:512] // reslice byte slice
	}
	_ = silly
}

func BenchmarkStandardLibrary(b *testing.B) {
	// b.Skip("invalid measurement unless using same time format")
	var foo string
	for i := 0; i < b.N; i++ {
		foo = when.Format(format)
	}
	_ = foo
}
