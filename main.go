package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

func main() {
	optFuncname := flag.String("f", "appendTime", "name of append function")
	optOutput := flag.String("o", "", "name of file to output")
	optPackage := flag.String("p", "main", "name of package to use")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "USAGE: %s [-f FUNCNAME] [-p PACKAGE] FORMAT_SPEC\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	cg, err := NewCodeGenerator(*optPackage, *optFuncname)
	if err != nil {
		bail(err)
	}

	if err = cg.Parse(flag.Arg(0)); err != nil {
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

	if err = cg.Emit(iow); err != nil {
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

type returnValues struct {
	values []string
}

type codeGenerator struct {
	libraries                     map[string]struct{}
	valuesFromInit                map[string]*returnValues
	initFromSymbol                map[string]string
	orderedSymbols                []string
	blobs                         []string
	packageName                   string
	functionName                  string
	gensymCounter                 int
	isDigit, isWeekdays, isMonths bool
	isU, isW, isMC, isM           bool
}

func NewCodeGenerator(packageName, functionName string) (*codeGenerator, error) {
	cg := &codeGenerator{
		valuesFromInit: make(map[string]*returnValues),
		initFromSymbol: make(map[string]string),
		libraries:      make(map[string]struct{}),
		packageName:    packageName,
		functionName:   functionName,
	}

	return cg, nil
}

func (cg *codeGenerator) Emit(iow io.Writer) error {
	var buf bytes.Buffer
	_, err := buf.WriteString(fmt.Sprintf("package %s\n\n", cg.packageName))
	if err != nil {
		return err
	}

	//
	// Library imports
	//
	cg.libraries["fmt"] = struct{}{}  // for main
	cg.libraries["time"] = struct{}{} // for main

	sortedLibraries := make([]string, 0, len(cg.libraries))
	for p := range cg.libraries {
		sortedLibraries = append(sortedLibraries, p)
	}
	sort.Strings(sortedLibraries)

	if _, err := buf.WriteString("import (\n"); err != nil {
		return err
	}

	for _, p := range sortedLibraries {
		if _, err = buf.WriteString(fmt.Sprintf("    \"%s\"\n", p)); err != nil {
			return err
		}
	}

	if _, err = buf.WriteString(")\n"); err != nil {
		return err
	}

	//
	// Main and specified function prefix
	//

	stub := fmt.Sprintf(`
func main() {
    // fmt.Println(string(appendTime(nil, time.Now())))
	fmt.Println(string(appendTime(nil, time.Date(2021, time.August, 22, 0, 0, 0, 1, time.UTC))))
}

func %s(buf []byte, t time.Time) []byte {
    // situational constant initializations
`, cg.functionName)

	if cg.isM {
		stub += "    const ampm = \"ampm\"\n"
		stub += "    var ampmIndex = []int{0,0,0,0,0,0,0,0,0,0,0,0,2,2,2,2,2,2,2,2,2,2,2,2}\n"
	}
	if cg.isMC {
		stub += "    const ampmc = \"AMPM\"\n"
		if !cg.isM {
			stub += "    var ampmIndex = []int{0,0,0,0,0,0,0,0,0,0,0,0,2,2,2,2,2,2,2,2,2,2,2,2}\n"
		}
	}
	if cg.isDigit {
		stub += "    const digits = \"0123456789 123456789\"\n"
		stub += "    var quotient, remainder int\n"
	}
	if cg.isWeekdays {
		stub += "    const weekdaysLong = \"SundayMondayTuesdayWednesdayThursdayFridaySaturday\"\n"
		stub += "    var weekdaysLongIndices = []int{0, 6, 12, 19, 28, 36, 42, 50}\n"
	}
	if cg.isMonths {
		stub += "    const monthsLong = \"JanuaryFebruaryMarchAprilMayJuneJulyAugustSeptemberOctoberNovemberDecember\"\n"
		stub += "    var monthsLongIndices = []int{0, 7, 15, 20, 25, 28, 32, 36, 42, 51, 58, 66, 74}\n"
	}
	if cg.isU {
		stub += "    var uFromWeekday = []byte{'7', '1', '2', '3', '4', '5', '6'}\n"
	}
	if cg.isW {
		stub += "    var wFromWeekday = []byte{'0', '1', '2', '3', '4', '5', '6'}\n"
	}

	stub += "    // dynamically generated variable initializations\n"
	if _, err = buf.WriteString(stub); err != nil {
		return err
	}

	completedInits := make(map[string]struct{})
	for _, symbol := range cg.orderedSymbols {
		init, ok := cg.initFromSymbol[symbol]
		if !ok {
			return fmt.Errorf("cannot find initialization for %q", symbol)
		}

		if _, ok := completedInits[init]; ok {
			continue // this init and its symbols already written
		}
		completedInits[init] = struct{}{}

		values, ok := cg.valuesFromInit[init]
		if !ok {
			return fmt.Errorf("cannot find values for %q, for %q", init, symbol)
		}
		foo := fmt.Sprintf("    %s := %s\n", strings.Join(values.values, ", "), init)
		if _, err = buf.WriteString(foo); err != nil {
			return err
		}
	}

	for _, blob := range cg.blobs {
		if _, err = buf.WriteString(blob); err != nil {
			return err
		}
	}

	if _, err = buf.WriteString("\n    return buf\n}\n"); err != nil {
		return err
	}

	_, err = iow.Write(buf.Bytes())
	return err
}

func (cg *codeGenerator) Parse(format string) error {
	var buf []byte
	var foundPercent bool

	for ri, rune := range format {
		if !foundPercent {
			if rune == '%' {
				foundPercent = true
				switch len(buf) {
				case 0:
					// no-op
				case 1:
					cg.blobs = append(cg.blobs, fmt.Sprintf("    buf = append(buf, %q)\n", buf[0]))
					buf = buf[:0]
				default:
					cg.blobs = append(cg.blobs, fmt.Sprintf("    buf = append(buf, %q...)\n", buf))
					buf = buf[:0]
				}
			} else {
				appendRune(&buf, rune)
			}
			continue
		}
		switch rune {
		case 'a':
			cg.blobs = append(cg.blobs, cg.writeWeekdayShort())
		case 'A':
			cg.blobs = append(cg.blobs, cg.writeWeekdayLong())
		case 'b':
			cg.blobs = append(cg.blobs, cg.writeMonthShort())
		case 'B':
			cg.blobs = append(cg.blobs, cg.writeMonthLong())
		case 'c':
			cg.blobs = append(cg.blobs, cg.writeC())
		case 'C':
			cg.blobs = append(cg.blobs, cg.writeCC())
		case 'd':
			cg.blobs = append(cg.blobs, cg.writeD())
		case 'D':
			cg.blobs = append(cg.blobs, cg.writeDC())
		case 'e':
			cg.blobs = append(cg.blobs, cg.writeE())
		case 'F':
			cg.blobs = append(cg.blobs, cg.writeFC())
		case 'g':
			cg.blobs = append(cg.blobs, cg.writeG())
		case 'G':
			cg.blobs = append(cg.blobs, cg.writeGC())
		case 'h':
			cg.blobs = append(cg.blobs, cg.writeMonthShort())
		case 'H':
			cg.blobs = append(cg.blobs, cg.writeHC())
		case 'I':
			cg.blobs = append(cg.blobs, cg.writeIC())
		case 'j':
			cg.blobs = append(cg.blobs, cg.writeJ())
		case 'k':
			cg.blobs = append(cg.blobs, cg.writeK())
		case 'l':
			cg.blobs = append(cg.blobs, cg.writeL())
		case 'm':
			cg.blobs = append(cg.blobs, cg.writeM())
		case 'M':
			cg.blobs = append(cg.blobs, cg.writeMC())
		case 'n':
			cg.blobs = append(cg.blobs, cg.writeN())
		case 'N':
			cg.blobs = append(cg.blobs, cg.writeNC())
		case 'p':
			cg.blobs = append(cg.blobs, cg.writeP())
		case 'P':
			cg.blobs = append(cg.blobs, cg.writePC())
		case 'r':
			cg.blobs = append(cg.blobs, cg.writeR())
		case 'R':
			cg.blobs = append(cg.blobs, cg.writeRC())
		case 's':
			cg.blobs = append(cg.blobs, cg.writeS())
		case 'S':
			cg.blobs = append(cg.blobs, cg.writeSC())
		case 't':
			cg.blobs = append(cg.blobs, cg.writeT())
		case 'T':
			cg.blobs = append(cg.blobs, cg.writeTC())
		case 'u':
			cg.blobs = append(cg.blobs, cg.writeU())
		case 'w':
			cg.blobs = append(cg.blobs, cg.writeW())
		case 'x':
			cg.blobs = append(cg.blobs, cg.writeDC())
		case 'X':
			cg.blobs = append(cg.blobs, cg.writeTC())
		case 'y':
			cg.blobs = append(cg.blobs, cg.writeY())
		case 'Y':
			cg.blobs = append(cg.blobs, cg.writeYC())
		case 'z':
			cg.blobs = append(cg.blobs, cg.writeZ())
		case 'Z':
			cg.blobs = append(cg.blobs, cg.writeZC())
		case '%':
			cg.blobs = append(cg.blobs, cg.writePercent())
		case '+':
			cg.blobs = append(cg.blobs, cg.writePlus())
		case '1':
			cg.blobs = append(cg.blobs, cg.writeTZ())
		case '2':
			cg.blobs = append(cg.blobs, cg.writeLMin())
		case '3':
			cg.blobs = append(cg.blobs, cg.writeMilli())
		case '4':
			cg.blobs = append(cg.blobs, cg.writeMicro())
		default:
			return fmt.Errorf("cannot recognize format verb %q at index %d", rune, ri)
		}
		foundPercent = false
	}

	if foundPercent {
		return errors.New("cannot find closing format verb")
	}
	switch len(buf) {
	case 0:
		// no-op
	case 1:
		cg.blobs = append(cg.blobs, fmt.Sprintf("    buf = append(buf, %q)\n", buf[0]))
		buf = buf[:0]
	default:
		cg.blobs = append(cg.blobs, fmt.Sprintf("    buf = append(buf, %q...)\n", buf))
		buf = buf[:0]
	}

	return nil
}

func appendRune(buf *[]byte, r rune) {
	if r < utf8.RuneSelf {
		*buf = append(*buf, byte(r))
		return
	}
	olen := len(*buf)
	*buf = append(*buf, 0, 0, 0, 0)              // grow buf large enough to accommodate largest possible UTF8 sequence
	n := utf8.EncodeRune((*buf)[olen:olen+4], r) // encode rune into newly allocated buf space
	*buf = (*buf)[:olen+n]                       // trim buf to actual size used by rune addition
}

func (cg *codeGenerator) gensym(x, y int, format string, a ...interface{}) string {
	x-- // convert x from 1..y to 0..(y-1)
	var symbol string

	init := fmt.Sprintf(format, a...)

	if values, ok := cg.valuesFromInit[init]; ok {
		if got, want := len(values.values), y; got != want {
			// TODO panic should return error
			panic(fmt.Errorf("found %d return values; expected %d", got, want))
		}
		symbol = values.values[x]
		if symbol == "_" {
			symbol = cg.symbol()
			values.values[x] = symbol
			cg.initFromSymbol[symbol] = init
		}
		return symbol
	}

	symbol = cg.symbol()
	values := make([]string, y)
	for i := 0; i < y; i++ {
		if i == x {
			values[i] = symbol
		} else {
			values[i] = "_"
		}
	}

	cg.initFromSymbol[symbol] = init
	cg.valuesFromInit[init] = &returnValues{values: values}
	return symbol
}

func (cg *codeGenerator) symbol() string {
	// Meant to be called by gensym method, but could be called from
	// elsewhere.
	symbol := fmt.Sprintf("gs%d", cg.gensymCounter)
	cg.orderedSymbols = append(cg.orderedSymbols, symbol)
	cg.gensymCounter++
	return symbol
}

func (cg *codeGenerator) write2DigitsMin(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write2DigitsMin
    quotient = %s / 10
	remainder = %s %% 10
	if quotient > 0 {
		buf = append(buf, digits[quotient])
	}
	buf = append(buf, digits[remainder])
`, value, value)
}

func (cg *codeGenerator) write2DigitsSpace(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write2DigitsSpace
    quotient = %s / 10
	remainder = %s %% 10
	buf = append(buf, digits[10+quotient])
	buf = append(buf, digits[remainder])
`, value, value)
}

func (cg *codeGenerator) write2DigitsZero(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write2DigitsZero
    quotient = %s / 10
	remainder = %s %% 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
`, value, value)
}

func (cg *codeGenerator) write3DigitsZero(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write3DigitsZero
	// hundreds
	quotient = %s / 100
	remainder = %s %% 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
`, value, value)
}

func (cg *codeGenerator) write4DigitsZero(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write4DigitsZero
    // thousands
	quotient = %s / 1000
	remainder = %s %% 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
`, value, value)
}

func (cg *codeGenerator) write6DigitsZero(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write6DigitsZero
	// hundred-thousands
	quotient = %s / 100000
	remainder = %s %% 100000
	buf = append(buf, digits[quotient])
	// ten-thousands
	quotient = remainder / 10000
	remainder = %s %% 10000
	buf = append(buf, digits[quotient])
	// thousands
	quotient = remainder / 1000
	remainder = %s %% 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
`, value, value, value, value)
}

func (cg *codeGenerator) write9DigitsZero(value string) string {
	cg.isDigit = true
	return fmt.Sprintf(`    // write9DigitsZero
	// hundred-millions
	quotient = %s / 100000000
	remainder = %s %% 100000000
	buf = append(buf, digits[quotient])
	// ten-millions
	quotient = remainder / 10000000
	remainder = %s %% 10000000
	buf = append(buf, digits[quotient])
	// millions
	quotient = remainder / 1000000
	remainder = %s %% 1000000
	buf = append(buf, digits[quotient])
	// hundred-thousands
	quotient = remainder / 100000
	remainder = %s %% 100000
	buf = append(buf, digits[quotient])
	// ten-thousands
	quotient = remainder / 10000
	remainder = %s %% 10000
	buf = append(buf, digits[quotient])
	// thousands
	quotient = remainder / 1000
	remainder = %s %% 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
`, value, value, value, value, value, value, value)
}

func (cg *codeGenerator) writeWeekdayShort() string {
	cg.isWeekdays = true
	wd := cg.gensym(1, 1, "t.Weekday()")
	index := cg.gensym(1, 1, "weekdaysLongIndices[%s]", wd)
	index3 := cg.gensym(1, 1, "%s + 3", index)
	return fmt.Sprintf(`
    // Append Weekday Short
	buf = append(buf, weekdaysLong[%s:%s]...)
`, index, index3)
}

func (cg *codeGenerator) writeWeekdayLong() string {
	cg.isWeekdays = true
	wd1 := cg.gensym(1, 1, "t.Weekday()")
	wd2 := cg.gensym(1, 1, "%s + 1", wd1)
	wdli1 := cg.gensym(1, 1, "weekdaysLongIndices[%s]", wd1)
	wdli2 := cg.gensym(1, 1, "weekdaysLongIndices[%s]", wd2)
	return fmt.Sprintf(`
    // Append Weekday Long
	buf = append(buf, weekdaysLong[%s:%s]...)
`, wdli1, wdli2)
}

func (cg *codeGenerator) writeMonthShort() string {
	cg.isMonths = true
	month := cg.gensym(2, 3, "t.Date()")
	monthMinusOne := cg.gensym(1, 1, "%s - 1", month)
	indexL := cg.gensym(1, 1, "monthsLongIndices[%s]", monthMinusOne)
	indexR := cg.gensym(1, 1, "%s + 3", indexL)
	return fmt.Sprintf(`
    // Append Month Short
    buf = append(buf, monthsLong[%s:%s]...)
`, indexL, indexR)
}

func (cg *codeGenerator) writeMonthLong() string {
	cg.isMonths = true
	month := cg.gensym(2, 3, "t.Date()")
	monthMinusOne := cg.gensym(1, 1, "%s - 1", month)
	indexL := cg.gensym(1, 1, "monthsLongIndices[%s]", monthMinusOne)
	indexR := cg.gensym(1, 1, "monthsLongIndices[%s]", month)
	return fmt.Sprintf(`
    // Append Month Long
	buf = append(buf, monthsLong[%s:%s]...)
`, indexL, indexR)
}

func (cg *codeGenerator) writeRune(r rune) string {
	return fmt.Sprintf("    buf = append(buf, %q)\n", r)
}

func (cg *codeGenerator) writeC() string {
	foo := "\n    // writeC\n"
	foo += cg.writeWeekdayShort()
	foo += cg.writeRune(' ')
	foo += cg.writeMonthShort()
	foo += cg.writeRune(' ')
	foo += cg.writeE()
	foo += cg.writeRune(' ')
	foo += cg.writeTC()
	foo += cg.writeRune(' ')
	foo += cg.writeYC()
	return foo
}

func (cg *codeGenerator) writeCC() string {
	year := cg.gensym(1, 3, "t.Date()")
	century := cg.gensym(1, 1, "%s / 100", year)
	return "\n    // writeCC\n" + cg.write2DigitsZero(century)
}

func (cg *codeGenerator) writeD() string {
	date := cg.gensym(3, 3, "t.Date()")
	return "\n    // writeD\n" + cg.write2DigitsZero(date)
}

func (cg *codeGenerator) writeDC() string {
	foo := "\n    // writeDC\n"
	foo += cg.write2DigitsZero(cg.gensym(1, 1, "int(%s)", cg.gensym(2, 3, "t.Date()")))
	foo += cg.writeRune('/')
	foo += cg.write2DigitsZero(cg.gensym(3, 3, "t.Date()"))
	foo += cg.writeRune('/')
	foo += cg.write2DigitsZero(cg.gensym(1, 1, "%s %% 100", cg.gensym(1, 3, "t.Date()")))
	return foo
}

func (cg *codeGenerator) writeE() string {
	date := cg.gensym(3, 3, "t.Date()")
	return "\n    // writeE\n" + cg.write2DigitsSpace(date)
}

func (cg *codeGenerator) writeFC() string {
	year := cg.gensym(1, 3, "t.Date()")
	month := cg.gensym(2, 3, "t.Date()")
	date := cg.gensym(3, 3, "t.Date()")
	foo := "\n    // writeFC\n"
	foo += cg.write4DigitsZero(year)
	foo += cg.writeRune('-')
	monthInt := cg.gensym(1, 1, "int(%s)", month)
	foo += cg.write2DigitsZero(monthInt)
	foo += cg.writeRune('-')
	foo += cg.write2DigitsZero(date)
	return foo
}

func (cg *codeGenerator) writeG() string {
	year := cg.gensym(1, 2, "t.ISOWeek()")
	year2 := cg.gensym(1, 1, "%s %% 100", year)
	return "\n    // writeG\n" + cg.write2DigitsZero(year2)
}

func (cg *codeGenerator) writeGC() string {
	year := cg.gensym(1, 2, "t.ISOWeek()")
	return "\n    // writeGC\n" + cg.write4DigitsZero(year)
}

func (cg *codeGenerator) writeHC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	return "\n    // writeHC\n" + cg.write2DigitsZero(hour)
}

func (cg *codeGenerator) writeIC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)
	return "\n    // writeIC\n" + cg.write2DigitsZero(hour12)
}

func (cg *codeGenerator) writeJ() string {
	return "\n    // writeJ\n" + cg.write3DigitsZero(cg.gensym(1, 1, "t.YearDay()"))
}

func (cg *codeGenerator) writeK() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	return "\n    // writeK\n" + cg.write2DigitsZero(hour)
}

func (cg *codeGenerator) writeL() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)
	return "\n    // writeL\n" + cg.write2DigitsSpace(hour12)
}

func (cg *codeGenerator) writeLMin() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)
	return "\n    // writeLMin\n" + cg.write2DigitsMin(hour12)
}

func (cg *codeGenerator) writeM() string {
	month := cg.gensym(2, 3, "t.Date()")
	monthInt := cg.gensym(1, 1, "int(%s)", month)
	return "\n    // writeM\n" + cg.write2DigitsZero(monthInt)
}

func (cg *codeGenerator) writeMC() string {
	minute := cg.gensym(2, 3, "t.Clock()")
	return "\n    // writeMC\n" + cg.write2DigitsZero(minute)
}

func (cg *codeGenerator) writeN() string {
	return "\n    // writeN\n" + cg.writeRune('\n')
}

func (cg *codeGenerator) writeNC() string {
	return "\n    // writeNC\n" + cg.write9DigitsZero(cg.gensym(1, 1, "t.Nanosecond()"))
}

func (cg *codeGenerator) writeMicro() string {
	nanos := cg.gensym(1, 1, "t.Nanosecond()")
	micros := cg.gensym(1, 1, "%s / 1000", nanos)
	return "\n    // writeMicro\n" + cg.write6DigitsZero(micros)
}

func (cg *codeGenerator) writeMilli() string {
	nanos := cg.gensym(1, 1, "t.Nanosecond()")
	millis := cg.gensym(1, 1, "%s / 1000000", nanos)
	return "\n    // writeMillis\n" + cg.write3DigitsZero(millis)
}

func (cg *codeGenerator) writeP() string {
	cg.isMC = true
	hour := cg.gensym(1, 3, "t.Clock()")
	hourIndex := cg.gensym(1, 1, "ampmIndex[%s]", hour)
	hourIndex2 := cg.gensym(1, 1, "%s + 2", hourIndex)
	return fmt.Sprintf(`
    // writeP
    buf = append(buf, ampmc[%s:%s]...)
`, hourIndex, hourIndex2)
}

func (cg *codeGenerator) writePC() string {
	cg.isM = true
	hour := cg.gensym(1, 3, "t.Clock()")
	hourIndex := cg.gensym(1, 1, "ampmIndex[%s]", hour)
	hourIndex2 := cg.gensym(1, 1, "%s + 2", hourIndex)
	return fmt.Sprintf(`
    // writePC
    buf = append(buf, ampm[%s:%s]...)
`, hourIndex, hourIndex2)
}

func (cg *codeGenerator) writeR() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	minute := cg.gensym(2, 3, "t.Clock()")
	second := cg.gensym(3, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)

	foo := "\n    // writeR\n"
	foo += cg.write2DigitsZero(hour12)
	foo += cg.writeRune(':')
	foo += cg.write2DigitsZero(minute)
	foo += cg.writeRune(':')
	foo += cg.write2DigitsZero(second)
	foo += cg.writeRune(' ')
	foo += cg.writeP()
	return foo
}

func (cg *codeGenerator) writeRC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	minute := cg.gensym(2, 3, "t.Clock()")
	foo := "\n    // writeRC\n"
	foo += cg.write2DigitsZero(hour)
	foo += cg.writeRune(':')
	foo += cg.write2DigitsZero(minute)
	return foo
}

func (cg *codeGenerator) writeS() string {
	epoch := cg.gensym(1, 1, "t.Unix()")
	cg.libraries["strconv"] = struct{}{}
	return fmt.Sprintf("    buf = strconv.AppendInt(buf, %s, 10) // writeS\n", epoch)
}

func (cg *codeGenerator) writeSC() string {
	second := cg.gensym(3, 3, "t.Clock()")
	return "\n    // writeSC\n" + cg.write2DigitsZero(second)
}

func (cg *codeGenerator) writeT() string {
	return "\n    // writeT\n" + cg.writeRune('\t')
}

func (cg *codeGenerator) writeTC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	minute := cg.gensym(2, 3, "t.Clock()")
	second := cg.gensym(3, 3, "t.Clock()")
	foo := "\n    // writeTC\n"
	foo += cg.write2DigitsZero(hour)
	foo += cg.writeRune(':')
	foo += cg.write2DigitsZero(minute)
	foo += cg.writeRune(':')
	foo += cg.write2DigitsZero(second)
	return foo
}

func (cg *codeGenerator) writeU() string {
	cg.isU = true
	wd := cg.gensym(1, 1, "t.Weekday()")
	u := cg.gensym(1, 1, "uFromWeekday[%s]", wd)
	return fmt.Sprintf("    // writeU\n    buf = append(buf, %s)\n", u)
}

func (cg *codeGenerator) writeW() string {
	cg.isW = true
	wd := cg.gensym(1, 1, "t.Weekday()")
	w := cg.gensym(1, 1, "wFromWeekday[%s]", wd)
	return fmt.Sprintf("    // writeW\n    buf = append(buf, %s)\n", w)
}

func (cg *codeGenerator) writeY() string {
	year := cg.gensym(1, 3, "t.Date()")
	year2 := cg.gensym(1, 1, "%s %% 100", year)
	return "\n    // writeY\n" + cg.write2DigitsZero(year2)
}

func (cg *codeGenerator) writeYC() string {
	year := cg.gensym(1, 3, "t.Date()")
	return "\n    // writeYC\n" + cg.write4DigitsZero(year)
}

func (cg *codeGenerator) writeZ() string {
	zoneSeconds := cg.gensym(2, 2, "t.Zone()")

	zoneHourPositive := cg.gensym(1, 1, "%s / 3600", zoneSeconds)
	zoneMinutePositive := cg.gensym(1, 1, "%s %% 3600 / 60", zoneSeconds)

	zoneNegative := cg.gensym(1, 1, "-"+zoneSeconds)
	zoneHourNegative := cg.gensym(1, 1, "%s / 3600", zoneNegative)
	zoneMinuteNegative := cg.gensym(1, 1, "%s %% 3600 / 60", zoneNegative)

	// TODO: table lookup?
	return fmt.Sprintf(`
    // writeZ
    if %s >= 0 {
        buf = append(buf, '+')
        %s
        %s
    } else {
        buf = append(buf, '-')
        %s
        %s
    }
`,
		zoneSeconds,
		cg.write2DigitsZero(zoneHourPositive), cg.write2DigitsZero(zoneMinutePositive),
		cg.write2DigitsZero(zoneHourNegative), cg.write2DigitsZero(zoneMinuteNegative))
}

func (cg *codeGenerator) writeZC() string {
	zoneName := cg.gensym(1, 2, "t.Zone()")
	return fmt.Sprintf("    buf = append(buf, %s...)\n    // writeZC\n", zoneName)
}

func (cg *codeGenerator) writeTZ() string {
	zoneSeconds := cg.gensym(2, 2, "t.Zone()")

	zoneHourPositive := cg.gensym(1, 1, "%s / 3600", zoneSeconds)
	zoneMinutePositive := cg.gensym(1, 1, "%s %% 3600 / 60", zoneSeconds)

	zoneNegative := cg.gensym(1, 1, "-"+zoneSeconds)
	zoneHourNegative := cg.gensym(1, 1, "%s / 3600", zoneNegative)
	zoneMinuteNegative := cg.gensym(1, 1, "%s %% 3600 / 60", zoneNegative)

	// TODO: table lookup?
	return fmt.Sprintf(`
    // writeTZ
    if %s == 0 {
        buf = append(buf, 'Z')
    } else if %s > 0 {
        buf = append(buf, '+')
        %s
        buf = append(buf, ':')
        %s
    } else {
        buf = append(buf, '-')
        %s
        buf = append(buf, ':')
        %s
    }
`,
		zoneSeconds,
		zoneSeconds,
		cg.write2DigitsZero(zoneHourPositive), cg.write2DigitsZero(zoneMinutePositive),
		cg.write2DigitsZero(zoneHourNegative), cg.write2DigitsZero(zoneMinuteNegative))
}

func (cg *codeGenerator) writePercent() string {
	return cg.writeRune('%')
}

func (cg *codeGenerator) writePlus() string {
	foo := "\n    // writePlus\n"
	foo += cg.writeWeekdayShort()
	foo += cg.writeRune(' ')
	foo += cg.writeMonthShort()
	foo += cg.writeRune(' ')
	foo += cg.writeE()
	foo += cg.writeRune(' ')
	foo += cg.writeTC()
	foo += cg.writeRune(' ')
	foo += cg.writeP()
	foo += cg.writeRune(' ')
	foo += cg.writeZC()
	foo += cg.writeRune(' ')
	foo += cg.writeYC()
	return foo
}
