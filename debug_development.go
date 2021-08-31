// +build sft_debug

package main

import (
	"fmt"
	"os"
)

// debug formats and prints arguments to stderr for development builds
func debug(f string, a ...interface{}) {
	os.Stderr.Write([]byte("sft: " + fmt.Sprintf(f, a...)))
}
