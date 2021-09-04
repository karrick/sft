package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	optAppend := flag.Bool("append", false, "use append")
	optExtra := flag.Bool("extra", false, "allow non-standard formatting verbs")
	optFuncname := flag.String("f", "appendTime", "name of append function")
	optMain := flag.Bool("m", false, "emit a main function")
	optOutput := flag.String("o", "", "name of file to output")
	optPackage := flag.String("p", "main", "name of package to use")
	optReformat := flag.Bool("reformat", false, "reformat like gofmt")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "USAGE: %s [-f FUNCNAME] [-o OUTPUT_FILE] [-p PACKAGE] FORMAT_SPEC\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	extra := *optExtra
	spec := flag.Arg(0)

	if a, ok := formatMap[spec]; ok {
		spec = a
		extra = true
	}

	args := make([]string, len(os.Args))
	for i, a := range os.Args {
		if i < len(args)-1 {
			args[i] = a
		} else {
			args[i] = "\"" + a + "\""
		}
	}
	cmd := strings.Join(args, " ")

	cg, err := NewCodeGenerator(spec, cmd, &Config{
		Package:    *optPackage,
		FuncName:   *optFuncname,
		AllowExtra: extra,
		UseAppend:  *optAppend,
		EmitMain:   *optMain,
		Reformat:   *optReformat,
	})
	if err != nil {
		bail(err)
	}

	var iow io.Writer = os.Stdout
	var fh *os.File

	if *optOutput != "" {
		fh, err = os.Create(*optOutput)
		if err != nil {
			bail(err)
		}
		iow = fh
	}

	if _, err = cg.WriteTo(iow); err != nil {
		bail(err)
	}

	if fh != nil {
		if err = fh.Close(); err != nil {
			bail(err)
		}
	}
}

func bail(err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", filepath.Base(os.Args[0]), err)
	os.Exit(1)
}

var formatMap map[string]string

func init() {
	formatMap = map[string]string{
		time.ANSIC:       "%c",
		time.UnixDate:    "%a %b %e %T %Z %Y",  // "Mon Jan _2 15:04:05 MST 2006",
		time.RubyDate:    "%a %b %d %T %z %Y",  // "Mon Jan 02 15:04:05 -0700 2006",
		time.RFC822:      "%d %b %y %R %Z",     // "02 Jan 06 15:04 MST",
		time.RFC822Z:     "%d %b %y %R %z",     // "02 Jan 06 15:04 -0700",
		time.RFC850:      "%A, %d-%b-%y %T %Z", // "Monday, 02-Jan-06 15:04:05 MST",
		time.RFC1123:     "%a, %d %b %Y %T %Z", // "Mon, 02 Jan 2006 15:04:05 MST",
		time.RFC1123Z:    "%a, %d %b %Y %T %z", // "Mon, 02 Jan 2006 15:04:05 -0700",
		time.RFC3339:     "%Y-%m-%dT%T%1",      // "2006-01-02T15:04:05Z07:00", // TODO: %1 not standard
		time.RFC3339Nano: "%Y-%m-%dT%T.%N%1",   // "2006-01-02T15:04:05.999999999Z07:00", // TODO: %1 not standard
		time.Kitchen:     "%2:%M%p",            // "3:04PM", // TODO: %2 not standard
		time.Stamp:       "%b %e %T",           // "Jan _2 15:04:05"
		time.StampMilli:  "%b %e %T.%3",        // "Jan _2 15:04:05.000" // TODO: %3 not standard
		time.StampMicro:  "%b %e %T.%4",        // "Jan _2 15:04:05.000000" // TODO: %4 not standard
		time.StampNano:   "%b %e %T.%N",        // "Jan _2 15:04:05.000000000"

		"ANSIC":       "%c",
		"UnixDate":    "%a %b %e %T %Z %Y",  // "Mon Jan _2 15:04:05 MST 2006",
		"RubyDate":    "%a %b %d %T %z %Y",  // "Mon Jan 02 15:04:05 -0700 2006",
		"RFC822":      "%d %b %y %R %Z",     // "02 Jan 06 15:04 MST",
		"RFC822Z":     "%d %b %y %R %z",     // "02 Jan 06 15:04 -0700",
		"RFC850":      "%A, %d-%b-%y %T %Z", // "Monday, 02-Jan-06 15:04:05 MST",
		"RFC1123":     "%a, %d %b %Y %T %Z", // "Mon, 02 Jan 2006 15:04:05 MST",
		"RFC1123Z":    "%a, %d %b %Y %T %z", // "Mon, 02 Jan 2006 15:04:05 -0700",
		"RFC3339":     "%Y-%m-%dT%T%1",      // "2006-01-02T15:04:05Z07:00", // TODO: %1 not standard
		"RFC3339Nano": "%Y-%m-%dT%T.%N%1",   // "2006-01-02T15:04:05.999999999Z07:00", // TODO: %1 not standard
		"Kitchen":     "%2:%M%p",            // "3:04PM", // TODO: %2 not standard
		"Stamp":       "%b %e %T",           // "Jan _2 15:04:05"
		"StampMilli":  "%b %e %T.%3",        // "Jan _2 15:04:05.000" // TODO: %3 not standard
		"StampMicro":  "%b %e %T.%4",        // "Jan _2 15:04:05.000000" // TODO: %4 not standard
		"StampNano":   "%b %e %T.%N",        // "Jan _2 15:04:05.000000000"
	}
}
