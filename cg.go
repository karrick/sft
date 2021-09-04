package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

type Config struct {
	Package    string
	FuncName   string
	AllowExtra bool
	EmitMain   bool
	Reformat   bool
	UseAppend  bool
}

type returnValues struct {
	values []string
}

type CodeGenerator struct {
	libraries      map[string]struct{}
	valuesFromInit map[string]*returnValues
	initFromSymbol map[string]string
	orderedSymbols []string

	// operations stores all of the formatting operations needed to be performed
	// in the order in which they are required to run. We have to build this out
	// of band with emitting the final output because before we have to identify
	// which imports to include at the top of the output before we emit these
	// formatting operations.
	buf []byte

	cmd                             string
	spec                            string
	packageName                     string
	functionName                    string
	gensymCounter                   int
	offset                          int // While >= 0, use this for offset; when -1 use runtime offset
	maxLength                       int
	isDigit, isWeekdays, isMonths   bool
	isU, isW, isMC, isM             bool
	reformat                        bool
	allowExtra, emitMain, useAppend bool
}

func NewCodeGenerator(spec, cmd string, config *Config) (*CodeGenerator, error) {
	var err error

	if spec == "" {
		return nil, errors.New("cannot create code generator without time format spec")
	}
	if config == nil {
		config = &Config{}
	}
	if config.Package == "" {
		config.Package = "main"
	}
	if config.FuncName == "" {
		config.FuncName = "formatTime"
	}
	cg := &CodeGenerator{
		valuesFromInit: make(map[string]*returnValues),
		initFromSymbol: make(map[string]string),
		libraries:      make(map[string]struct{}),
		spec:           spec,
		cmd:            cmd,
		packageName:    config.Package,
		functionName:   config.FuncName,
		allowExtra:     config.AllowExtra,
		emitMain:       config.EmitMain,
		useAppend:      config.UseAppend,
		reformat:       config.Reformat,
	}

	buf, err := cg.scan()
	if err != nil {
		return nil, err
	}
	// fmt.Fprintf(os.Stderr, "BEGIN:\n%s\nEND\n", buf)

	if err = cg.prepare(buf); err != nil {
		return nil, err
	}

	return cg, nil
}

// Scan the spec string and build the output for the required operations.
func (cg *CodeGenerator) scan() ([]byte, error) {
	buf := make([]byte, 0, 65536)

	var stringConstant []byte
	var foundPercent bool

	for ri, rune := range cg.spec {
		if !foundPercent {
			if rune == '%' {
				foundPercent = true
				if len(stringConstant) > 0 {
					buf = append(buf, cg.writeStringConstant(string(stringConstant))...)
					stringConstant = stringConstant[:0]
				}
			} else {
				appendRune(&stringConstant, rune)
			}
			continue
		}
		switch rune {
		case 'a':
			buf = append(buf, cg.writeWeekdayShort()...)
		case 'A':
			buf = append(buf, cg.writeWeekdayLong()...)
		case 'b':
			buf = append(buf, cg.writeMonthShort()...)
		case 'B':
			buf = append(buf, cg.writeMonthLong()...)
		case 'c':
			buf = append(buf, cg.writeC()...)
		case 'C':
			buf = append(buf, cg.writeCC()...)
		case 'd':
			buf = append(buf, cg.writeD()...)
		case 'D':
			buf = append(buf, cg.writeDC()...)
		case 'e':
			buf = append(buf, cg.writeE()...)
		case 'F':
			buf = append(buf, cg.writeFC()...)
		case 'g':
			buf = append(buf, cg.writeG()...)
		case 'G':
			buf = append(buf, cg.writeGC()...)
		case 'h':
			buf = append(buf, cg.writeMonthShort()...)
		case 'H':
			buf = append(buf, cg.writeHC()...)
		case 'I':
			buf = append(buf, cg.writeIC()...)
		case 'j':
			buf = append(buf, cg.writeJ()...)
		case 'k':
			buf = append(buf, cg.writeK()...)
		case 'l':
			buf = append(buf, cg.writeL()...)
		case 'm':
			buf = append(buf, cg.writeM()...)
		case 'M':
			buf = append(buf, cg.writeMC()...)
		case 'n':
			buf = append(buf, cg.writeN()...)
		case 'N':
			buf = append(buf, cg.writeNC()...)
		case 'p':
			buf = append(buf, cg.writeP()...)
		case 'P':
			buf = append(buf, cg.writePC()...)
		case 'r':
			buf = append(buf, cg.writeR()...)
		case 'R':
			buf = append(buf, cg.writeRC()...)
		case 's':
			buf = append(buf, cg.writeS()...)
		case 'S':
			buf = append(buf, cg.writeSC()...)
		case 't':
			buf = append(buf, cg.writeT()...)
		case 'T':
			buf = append(buf, cg.writeTC()...)
		case 'u':
			buf = append(buf, cg.writeU()...)
		case 'w':
			buf = append(buf, cg.writeW()...)
		case 'x':
			buf = append(buf, cg.writeDC()...)
		case 'X':
			buf = append(buf, cg.writeTC()...)
		case 'y':
			buf = append(buf, cg.writeY()...)
		case 'Y':
			buf = append(buf, cg.writeYC()...)
		case 'z':
			buf = append(buf, cg.writeZ()...)
		case 'Z':
			buf = append(buf, cg.writeZC()...)
		case '%':
			buf = append(buf, cg.writePercent()...)
		case '+':
			buf = append(buf, cg.writePlus()...)
		case '1':
			if !cg.allowExtra {
				return nil, fmt.Errorf("cannot recognize format verb %q at index %d", rune, ri)
			}
			buf = append(buf, cg.writeTZ()...)
		case '2':
			if !cg.allowExtra {
				return nil, fmt.Errorf("cannot recognize format verb %q at index %d", rune, ri)
			}
			buf = append(buf, cg.writeLMin()...)
		case '3':
			if !cg.allowExtra {
				return nil, fmt.Errorf("cannot recognize format verb %q at index %d", rune, ri)
			}
			buf = append(buf, cg.writeMilli()...)
		case '4':
			if !cg.allowExtra {
				return nil, fmt.Errorf("cannot recognize format verb %q at index %d", rune, ri)
			}
			buf = append(buf, cg.writeMicro()...)
		default:
			return nil, fmt.Errorf("cannot recognize format verb %q at index %d", rune, ri)
		}
		foundPercent = false
	}

	if foundPercent {
		return nil, errors.New("cannot find closing format verb")
	}
	if len(stringConstant) > 0 {
		buf = append(buf, cg.writeStringConstant(string(stringConstant))...)
	}

	return buf, nil
}

// prepare final output
//
// sbuf -> buf -> cg.buf
func (cg *CodeGenerator) prepare(sbuf []byte) error {
	buf := make([]byte, 0, 1024+len(sbuf))
	var err error

	appendString(&buf, "package %s\n\n", cg.packageName)

	//
	// Library imports
	//
	if cg.emitMain {
		cg.libraries["fmt"] = struct{}{} // for main
	}
	cg.libraries["time"] = struct{}{}

	sortedLibraries := make([]string, 0, len(cg.libraries))
	for p := range cg.libraries {
		sortedLibraries = append(sortedLibraries, p)
	}
	sort.Strings(sortedLibraries)

	appendString(&buf, "import (\n")
	for _, p := range sortedLibraries {
		appendString(&buf, "    \"%s\"\n", p)
	}
	appendString(&buf, ")\n\n")

	//
	// Main and specified function prefix
	//
	if cg.emitMain {
		appendString(&buf, `func main() {
    when := time.Date(2006, time.January, 2, 3, 4, 5, 123456789, time.UTC)
    fmt.Println(string(%s(make([]byte, 128), when)))
}

`, cg.functionName)
	}

	appendString(&buf, "func %s(buf []byte, t time.Time) []byte {\n", cg.functionName)

	if cg.isM {
		appendString(&buf, "    const ampm = \"ampm\"\n")
		appendString(&buf, "    var ampmIndex = []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}\n")
	}
	if cg.isMC {
		appendString(&buf, "    const ampmc = \"AMPM\"\n")
		if !cg.isM {
			appendString(&buf, "    var ampmIndex = []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}\n")
		}
	}
	if cg.isDigit {
		appendString(&buf, "    const digits = \"0123456789 123456789\"\n")
		appendString(&buf, "    var quotient, remainder int\n")
	}
	if cg.isWeekdays {
		appendString(&buf, "    const weekdaysLong = \"SundayMondayTuesdayWednesdayThursdayFridaySaturday\"\n")
		appendString(&buf, "    var weekdaysLongIndices = []int{0, 6, 12, 19, 28, 36, 42, 50}\n")
	}
	if cg.isMonths {
		appendString(&buf, "    const monthsLong = \"JanuaryFebruaryMarchAprilMayJuneJulyAugustSeptemberOctoberNovemberDecember\"\n")
		appendString(&buf, "    var monthsLongIndices = []int{0, 7, 15, 20, 25, 28, 32, 36, 42, 51, 58, 66, 74}\n")
	}
	if cg.isU {
		appendString(&buf, "    var uFromWeekday = []string{\"7\", \"1\", \"2\", \"3\", \"4\", \"5\", \"6\"}\n")
	}
	if cg.isW {
		appendString(&buf, "    var wFromWeekday = []string{\"0\", \"1\", \"2\", \"3\", \"4\", \"5\", \"6\"}\n")
	}

	if cg.useAppend {
		// TODO: Upgrade so that function appends to existing byte slice.
		appendString(&buf, `
    if len(buf) > 0 {
        buf = buf[:0]
    }

`)
	} else {
		appendString(&buf, `
    if len(buf) < %d {
        buf = make([]byte, %d)
    }

`, cg.maxLength, cg.maxLength)
	}

	// dynamically generated variable initializations
	completedInits := make(map[string]struct{})
	for _, symbol := range cg.orderedSymbols {
		init, ok := cg.initFromSymbol[symbol]
		if !ok {
			return fmt.Errorf("cannot find initialization for %q", symbol)
		}

		if _, ok := completedInits[init]; ok {
			continue // this init and its symbols have already been written
		}
		completedInits[init] = struct{}{}

		values, ok := cg.valuesFromInit[init]
		if !ok {
			return fmt.Errorf("cannot find values for %q, for %q", init, symbol)
		}
		appendString(&buf, "    %s := %s\n", strings.Join(values.values, ", "), init)
	}
	appendString(&buf, "\n")

	// Emit all of the operations in their required sequence.
	buf = append(buf, sbuf...) // bb.Bytes()...)

	if !cg.useAppend {
		if cg.offset >= 0 {
			// Scanner was able to track offset because everything was fixed
			// width output.
			appendString(&buf, "\n    buf = buf[:%d]\n", cg.offset)
		} else {
			// Because one or more formatting verbs were not fixed width output,
			// scanner was not able to track offset at runtime, and had to emit
			// code to track it at runtime.
			appendString(&buf, "\n    buf = buf[:offset]\n")
		}
	}
	appendString(&buf, "    return buf\n}\n")

	// Because gofmt removes comments, we need to run gofmt first, then append
	// the result to after the header.
	if cg.reformat {
		buf, err = gofmt(buf)
		if err != nil {
			bail(err)
		}
	}
	header := make([]byte, 0, 100+len(buf))
	appendString(&header, `// This file was auto generated using the following command:
//    %s

`, cg.cmd)
	cg.buf = append(header, buf...)

	return nil
}

func appendString(buf *[]byte, f string, a ...interface{}) {
	*buf = append(*buf, fmt.Sprintf(f, a...)...)
}

func gofmt(source []byte) ([]byte, error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", string(source), parser.AllErrors)
	if err != nil {
		return nil, err
	}
	bb := new(bytes.Buffer)
	if err := format.Node(bb, fs, f); err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

func (cg *CodeGenerator) Bytes() []byte {
	return cg.buf
}

func (cg *CodeGenerator) String() string {
	return string(cg.buf)
}

func (cg *CodeGenerator) WriteTo(iow io.Writer) (int64, error) {
	n, err := iow.Write(cg.buf)
	return int64(n), err
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

func (cg *CodeGenerator) gensym(x, y int, format string, a ...interface{}) string {
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

func (cg *CodeGenerator) symbol() string {
	// Meant to be called by gensym method, but could be called from
	// elsewhere.
	symbol := fmt.Sprintf("gs%d", cg.gensymCounter)
	cg.orderedSymbols = append(cg.orderedSymbols, symbol)
	cg.gensymCounter++
	return symbol
}

func (cg *CodeGenerator) write2DigitsMin(value string) string {
	cg.isDigit = true
	cg.maxLength += 2

	if cg.useAppend {
		return fmt.Sprintf(`    // write2DigitsMin append
    quotient = %s / 10
	remainder = %s %% 10
	if quotient > 0 {
		buf = append(buf, digits[quotient])
	}
	buf = append(buf, digits[remainder])
`, value, value)
	}

	var off string
	if cg.offset >= 0 {
		off = fmt.Sprintf("    offset := %d // following formatting verb has variable length\n", cg.offset)
		cg.offset = -1 // must use dynamic offsets
	}

	return off + fmt.Sprintf(`    // write2DigitsMin runtime offset
    quotient = %s / 10
	remainder = %s %% 10
	if quotient > 0 {
        buf[offset] = digits[quotient]
        offset++
	}
    buf[offset] = digits[remainder]
    offset++
`, value, value)
}

func (cg *CodeGenerator) write2DigitsSpace(value string) string {
	cg.isDigit = true
	cg.maxLength += 2

	if cg.useAppend {
		return fmt.Sprintf(`    // write2DigitsSpace append
    quotient = %s / 10
	remainder = %s %% 10
	buf = append(buf, digits[10+quotient])
	buf = append(buf, digits[remainder])
`, value, value)
	}

	if cg.offset >= 0 {
		cg.offset += 2
		return fmt.Sprintf(`    // write2DigitsSpace codegen offset
    quotient = %s / 10
	remainder = %s %% 10
	buf[%d] = digits[10+quotient]
	buf[%d] = digits[remainder]
`, value, value, cg.offset-2, cg.offset-1)
	}

	return fmt.Sprintf(`    // write2DigitsSpace runtime offset
    quotient = %s / 10
	remainder = %s %% 10
	buf[offset] = digits[10+quotient]
	buf[offset+1] = digits[remainder]
    offset += 2
`, value, value)
}

func (cg *CodeGenerator) write2DigitsZero(value string) string {
	cg.isDigit = true
	cg.maxLength += 2

	if cg.useAppend {
		return fmt.Sprintf(`    // write2DigitsZero append
    quotient = %s / 10
	remainder = %s %% 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
`, value, value)
	}

	if cg.offset >= 0 {
		cg.offset += 2
		return fmt.Sprintf(`    // write2DigitsZero codegen offset
    quotient = %s / 10
	remainder = %s %% 10
	buf[%d] = digits[quotient]
	buf[%d] = digits[remainder]
`, value, value, cg.offset-2, cg.offset-1)
	}

	return fmt.Sprintf(`    // write2DigitsZero runtime offset
    quotient = %s / 10
	remainder = %s %% 10
	buf[offset] = digits[quotient]
	buf[offset+1] = digits[remainder]
    offset += 2
`, value, value)
}

func (cg *CodeGenerator) write3DigitsZero(value string) string {
	cg.isDigit = true
	cg.maxLength += 3

	if cg.useAppend {
		return fmt.Sprintf(`    // write3DigitsZero append
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

	if cg.offset >= 0 {
		cg.offset += 3
		return fmt.Sprintf(`    // write3DigitsZero codegen offset
	// hundreds
	quotient = %s / 100
	remainder = %s %% 100
	buf[%d] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[%d] = digits[quotient]
	// ones
	buf[%d] = digits[remainder]
`, value, value, cg.offset-3, cg.offset-2, cg.offset-1)
	}

	return fmt.Sprintf(`    // write3DigitsZero runtime offset
	// hundreds
	quotient = %s / 100
	remainder = %s %% 100
	buf[offset] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[offset+1] = digits[quotient]
	// ones
	buf[offset+2] = digits[remainder]
    offset += 3
`, value, value)
}

func (cg *CodeGenerator) write4DigitsZero(value string) string {
	cg.isDigit = true
	cg.maxLength += 4

	if cg.useAppend {
		return fmt.Sprintf(`    // write4DigitsZero append
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

	if cg.offset >= 0 {
		cg.offset += 4
		return fmt.Sprintf(`    // write4DigitsZero codegen offset
    // thousands
	quotient = %s / 1000
	remainder = %s %% 1000
	buf[%d] = digits[quotient]
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf[%d] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[%d] = digits[quotient]
	// ones
	buf[%d] = digits[remainder]
`, value, value, cg.offset-4, cg.offset-3, cg.offset-2, cg.offset-1)
	}

	return fmt.Sprintf(`    // write4DigitsZero runtime offset
    // thousands
	quotient = %s / 1000
	remainder = %s %% 1000
	buf[offset] = digits[quotient]
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf[offset+1] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[offset+2] = digits[quotient]
	// ones
	buf[offset+3] = digits[remainder]
    offset += 4
`, value, value)
}

func (cg *CodeGenerator) write6DigitsZero(value string) string {
	cg.isDigit = true
	cg.maxLength += 6

	if cg.useAppend {
		return fmt.Sprintf(`    // write6DigitsZero append
	// hundred-thousands
	quotient = %s / 100000
	remainder = %s %% 100000
	buf = append(buf, digits[quotient])
	// ten-thousands
	quotient = remainder / 10000
	remainder %%= 10000
	buf = append(buf, digits[quotient])
	// thousands
	quotient = remainder / 1000
	remainder %%= 1000
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

	if cg.offset >= 0 {
		cg.offset += 6
		return fmt.Sprintf(`    // write6DigitsZero codegen offset
	// hundred-thousands
	quotient = %s / 100000
	remainder = %s %% 100000
	buf[%d] = digits[quotient]
	// ten-thousands
	quotient = remainder / 10000
	remainder %%= 10000
	buf[%d] = digits[quotient]
	// thousands
	quotient = remainder / 1000
	remainder %%= 1000
	buf[%d] = digits[quotient]
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf[%d] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[%d] = digits[quotient]
	// ones
	buf[%d] = digits[remainder]
`, value, value, cg.offset-6, cg.offset-5, cg.offset-3, cg.offset-2, cg.offset-1)
	}

	return fmt.Sprintf(`    // write6DigitsZero runtime offset
	// hundred-thousands
	quotient = %s / 100000
	remainder = %s %% 100000
	buf[offset] = digits[quotient]
	// ten-thousands
	quotient = remainder / 10000
	remainder %%= 10000
	buf[offset+1] = digits[quotient]
	// thousands
	quotient = remainder / 1000
	remainder %%= 1000
	buf[offset+2] = digits[quotient]
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf[offset+3] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[offset+4] = digits[quotient]
	// ones
	buf[offset+5] = digits[remainder]
    offset += 6
`, value, value)
}

func (cg *CodeGenerator) write9DigitsZero(value string) string {
	cg.isDigit = true
	cg.maxLength += 9

	if cg.useAppend {
		return fmt.Sprintf(`    // write9DigitsZero append
	// hundred-millions
	quotient = %s / 100000000
	remainder = %s %% 100000000
	buf = append(buf, digits[quotient])
	// ten-millions
	quotient = remainder / 10000000
	remainder %%= 10000000
	buf = append(buf, digits[quotient])
	// millions
	quotient = remainder / 1000000
	remainder %%= 1000000
	buf = append(buf, digits[quotient])
	// hundred-thousands
	quotient = remainder / 100000
	remainder %%= 100000
	buf = append(buf, digits[quotient])
	// ten-thousands
	quotient = remainder / 10000
	remainder %%= 10000
	buf = append(buf, digits[quotient])
	// thousands
	quotient = remainder / 1000
	remainder %%= 1000
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

	if cg.offset >= 0 {
		cg.offset += 9
		return fmt.Sprintf(`    // write9DigitsZero codegen offset
	// hundred-millions
	quotient = %s / 100000000
	remainder = %s %% 100000000
	buf[%d] = digits[quotient]
	// ten-millions
	quotient = remainder / 10000000
	remainder %%= 10000000
	buf[%d] = digits[quotient]
	// millions
	quotient = remainder / 1000000
	remainder %%= 1000000
	buf[%d] = digits[quotient]
	// hundred-thousands
	quotient = remainder / 100000
	remainder %%= 100000
	buf[%d] = digits[quotient]
	// ten-thousands
	quotient = remainder / 10000
	remainder %%= 10000
	buf[%d] = digits[quotient]
	// thousands
	quotient = remainder / 1000
	remainder %%= 1000
	buf[%d] = digits[quotient]
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf[%d] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[%d] = digits[quotient]
	// ones
	buf[%d] = digits[remainder]
`,
			value, value,
			cg.offset-9, cg.offset-8, cg.offset-7,
			cg.offset-6, cg.offset-5, cg.offset-4,
			cg.offset-3, cg.offset-2, cg.offset-1)
	}

	return fmt.Sprintf(`    // write9DigitsZero runtime offset
	// hundred-millions
	quotient = %s / 100000000
	remainder = %s %% 100000000
	buf[offset] = digits[quotient]
	// ten-millions
	quotient = remainder / 10000000
	remainder %%= 10000000
	buf[offset+1] = digits[quotient]
	// millions
	quotient = remainder / 1000000
	remainder %%= 1000000
	buf[offset+2] = digits[quotient]
	// hundred-thousands
	quotient = remainder / 100000
	remainder %%= 100000
	buf[offset+3] = digits[quotient]
	// ten-thousands
	quotient = remainder / 10000
	remainder %%= 10000
	buf[offset+4] = digits[quotient]
	// thousands
	quotient = remainder / 1000
	remainder %%= 1000
	buf[offset+5] = digits[quotient]
	// hundreds
	quotient = remainder / 100
	remainder %%= 100
	buf[offset+6] = digits[quotient]
	// tens
	quotient = remainder / 10
	remainder %%= 10
	buf[offset+7] = digits[quotient]
	// ones
	buf[offset+8] = digits[remainder]
    offset += 9
`, value, value)
}

func (cg *CodeGenerator) writeWeekdayShort() string {
	cg.isWeekdays = true
	cg.maxLength += 3

	wd := cg.gensym(1, 1, "t.Weekday()")
	indexL := cg.gensym(1, 1, "weekdaysLongIndices[%s]", wd)
	indexR := cg.gensym(1, 1, "%s + 3", indexL)

	if cg.useAppend {
		return fmt.Sprintf(`
    // Weekday Short append
	buf = append(buf, weekdaysLong[%s:%s]...)
`, indexL, indexR)
	}

	if cg.offset >= 0 {
		foo := fmt.Sprintf("    // Weekday short codegen offset\n    _ = %s\n", indexR)
		for i := 0; i < 3; i++ {
			foo += fmt.Sprintf("    buf[%d] = weekdaysLong[%s+%d]\n", cg.offset, indexL, i)
			cg.offset++
		}
		return foo
	}

	return fmt.Sprintf(`
    // Weekday Short runtime offset
    offset += copy(buf[offset:], weekdaysLong[%s:%s])
`, indexL, indexR)
}

func (cg *CodeGenerator) writeWeekdayLong() string {
	cg.isWeekdays = true
	cg.maxLength += 9 // Wednesday

	wd1 := cg.gensym(1, 1, "t.Weekday()")
	wd2 := cg.gensym(1, 1, "%s + 1", wd1)
	wdli1 := cg.gensym(1, 1, "weekdaysLongIndices[%s]", wd1)
	wdli2 := cg.gensym(1, 1, "weekdaysLongIndices[%s]", wd2)

	if cg.useAppend {
		return fmt.Sprintf(`
    // Weekday Long append
	buf = append(buf, weekdaysLong[%s:%s]...)
`, wdli1, wdli2)
	}

	var off string
	if cg.offset >= 0 {
		off = fmt.Sprintf("    offset := %d // following formatting verb has variable length\n", cg.offset)
		cg.offset = -1 // must use dynamic offsets
	}

	return off + fmt.Sprintf(`
    // Weekday Long runtime offset
	offset += copy(buf[offset:], weekdaysLong[%s:%s])
`, wdli1, wdli2)
}

func (cg *CodeGenerator) writeMonthShort() string {
	cg.isMonths = true
	cg.maxLength += 3

	month := cg.gensym(2, 3, "t.Date()")
	monthMinusOne := cg.gensym(1, 1, "%s - 1", month)
	indexL := cg.gensym(1, 1, "monthsLongIndices[%s]", monthMinusOne)
	indexR := cg.gensym(1, 1, "%s + 3", indexL)

	if cg.useAppend {
		return fmt.Sprintf(`
    // Month Short append
    buf = append(buf, monthsLong[%s:%s]...)
`, indexL, indexR)
	}

	if cg.offset >= 0 {
		foo := fmt.Sprintf("    // Month short codegen offset\n    _ = %s\n", indexR)
		for i := 0; i < 3; i++ {
			foo += fmt.Sprintf("    buf[%d] = monthsLong[%s+%d]\n", cg.offset, indexL, i)
			cg.offset++
		}
		return foo
	}

	return fmt.Sprintf(`
    // Month Short runtime offset
    offset += copy(buf[offset:], monthsLong[%s:%s])
`, indexL, indexR)

}

func (cg *CodeGenerator) writeMonthLong() string {
	cg.isMonths = true
	cg.maxLength += 9 // september

	month := cg.gensym(2, 3, "t.Date()")
	monthMinusOne := cg.gensym(1, 1, "%s - 1", month)
	indexL := cg.gensym(1, 1, "monthsLongIndices[%s]", monthMinusOne)
	indexR := cg.gensym(1, 1, "monthsLongIndices[%s]", month)

	if cg.useAppend {
		return fmt.Sprintf(`
    // Month Long append
	buf = append(buf, monthsLong[%s:%s]...)
`, indexL, indexR)
	}

	var off string
	if cg.offset >= 0 {
		off = fmt.Sprintf("    offset := %d // following formatting verb has variable length\n", cg.offset)
		cg.offset = -1 // must use dynamic offsets
	}

	return off + fmt.Sprintf(`
    // Month Long runtime offset
	offset += copy(buf[offset:], monthsLong[%s:%s])
`, indexL, indexR)
}

func (cg *CodeGenerator) writeStringConstant(someString string) string {
	ls := len(someString)
	if ls == 0 {
		return ""
	}
	cg.maxLength += ls

	if cg.useAppend {
		return fmt.Sprintf("    buf = append(buf, %q...) // writeStringConstant append\n", someString)
	}

	if cg.offset >= 0 {
		// According to Go standard library, runtime.memmove optimizes
		// transfers of byte slices less than 2K characters. This does
		// not check the size of the string constants, because it is
		// unrealistic that they are going to be longer than 2 KiB.
		foo := "    // writeStringConstant codegen offset\n"
		for i := 0; i < ls; i++ {
			foo += fmt.Sprintf("    buf[%d] = %q\n", cg.offset, someString[i])
			cg.offset++
		}
		return foo
	}

	return fmt.Sprintf("    offset += copy(buf[offset:], %q) // writeStringConstant runtime offset\n", someString)
}

func (cg *CodeGenerator) writeStringValue(someValue string) string {
	if cg.useAppend {
		return fmt.Sprintf("    buf = append(buf, %s...) // writeStringValue\n", someValue)
	}
	var off string
	if cg.offset >= 0 {
		off = fmt.Sprintf("    offset := %d // following formatting verb has variable length\n", cg.offset)
		cg.offset = -1 // must use dynamic offsets
	}
	return off + fmt.Sprintf("    offset += copy(buf[offset:], %s) // writeStringValue runtime offset\n", someValue)
}

func (cg *CodeGenerator) writeC() string {
	foo := "\n    // writeC\n"
	foo += cg.writeWeekdayShort()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeMonthShort()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeE()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeTC()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeYC()
	return foo
}

func (cg *CodeGenerator) writeCC() string {
	year := cg.gensym(1, 3, "t.Date()")
	century := cg.gensym(1, 1, "%s / 100", year)
	return "\n    // writeCC\n" + cg.write2DigitsZero(century)
}

func (cg *CodeGenerator) writeD() string {
	date := cg.gensym(3, 3, "t.Date()")
	return "\n    // writeD\n" + cg.write2DigitsZero(date)
}

func (cg *CodeGenerator) writeDC() string {
	foo := "\n    // writeDC\n"
	foo += cg.write2DigitsZero(cg.gensym(1, 1, "int(%s)", cg.gensym(2, 3, "t.Date()")))
	foo += cg.writeStringConstant("/")
	foo += cg.write2DigitsZero(cg.gensym(3, 3, "t.Date()"))
	foo += cg.writeStringConstant("/")
	foo += cg.write2DigitsZero(cg.gensym(1, 1, "%s %% 100", cg.gensym(1, 3, "t.Date()")))
	return foo
}

func (cg *CodeGenerator) writeE() string {
	date := cg.gensym(3, 3, "t.Date()")
	return "\n    // writeE\n" + cg.write2DigitsSpace(date)
}

func (cg *CodeGenerator) writeFC() string {
	year := cg.gensym(1, 3, "t.Date()")
	month := cg.gensym(2, 3, "t.Date()")
	date := cg.gensym(3, 3, "t.Date()")
	monthInt := cg.gensym(1, 1, "int(%s)", month)
	foo := "\n    // writeFC\n"
	foo += cg.write4DigitsZero(year)
	foo += cg.writeStringConstant("-")
	foo += cg.write2DigitsZero(monthInt)
	foo += cg.writeStringConstant("-")
	foo += cg.write2DigitsZero(date)
	return foo
}

func (cg *CodeGenerator) writeG() string {
	year := cg.gensym(1, 2, "t.ISOWeek()")
	year2 := cg.gensym(1, 1, "%s %% 100", year)
	return "\n    // writeG\n" + cg.write2DigitsZero(year2)
}

func (cg *CodeGenerator) writeGC() string {
	year := cg.gensym(1, 2, "t.ISOWeek()")
	return "\n    // writeGC\n" + cg.write4DigitsZero(year)
}

func (cg *CodeGenerator) writeHC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	return "\n    // writeHC\n" + cg.write2DigitsZero(hour)
}

func (cg *CodeGenerator) writeIC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)
	return "\n    // writeIC\n" + cg.write2DigitsZero(hour12)
}

func (cg *CodeGenerator) writeJ() string {
	yearday := cg.gensym(1, 1, "t.YearDay()")
	return "\n    // writeJ\n" + cg.write3DigitsZero(yearday)
}

func (cg *CodeGenerator) writeK() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	return "\n    // writeK\n" + cg.write2DigitsZero(hour)
}

func (cg *CodeGenerator) writeL() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)
	return "\n    // writeL\n" + cg.write2DigitsSpace(hour12)
}

func (cg *CodeGenerator) writeLMin() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)
	return "\n    // writeLMin\n" + cg.write2DigitsMin(hour12)
}

func (cg *CodeGenerator) writeM() string {
	month := cg.gensym(2, 3, "t.Date()")
	monthInt := cg.gensym(1, 1, "int(%s)", month)
	return "\n    // writeM\n" + cg.write2DigitsZero(monthInt)
}

func (cg *CodeGenerator) writeMC() string {
	minute := cg.gensym(2, 3, "t.Clock()")
	return "\n    // writeMC\n" + cg.write2DigitsZero(minute)
}

func (cg *CodeGenerator) writeN() string {
	return "\n    // writeN\n" + cg.writeStringConstant("\n")
}

func (cg *CodeGenerator) writeNC() string {
	nanos := cg.gensym(1, 1, "t.Nanosecond()")
	return "\n    // writeNC\n" + cg.write9DigitsZero(nanos)
}

func (cg *CodeGenerator) writeMicro() string {
	nanos := cg.gensym(1, 1, "t.Nanosecond()")
	micros := cg.gensym(1, 1, "%s / 1000", nanos)
	return "\n    // writeMicro\n" + cg.write6DigitsZero(micros)
}

func (cg *CodeGenerator) writeMilli() string {
	nanos := cg.gensym(1, 1, "t.Nanosecond()")
	millis := cg.gensym(1, 1, "%s / 1000000", nanos)
	return "\n    // writeMillis\n" + cg.write3DigitsZero(millis)
}

func (cg *CodeGenerator) writeP() string {
	cg.isMC = true
	cg.maxLength += 2

	hour := cg.gensym(1, 3, "t.Clock()")
	hourIndex := cg.gensym(1, 1, "ampmIndex[%s]", hour)
	hourIndex2 := cg.gensym(1, 1, "%s + 2", hourIndex)

	if cg.useAppend {
		return fmt.Sprintf(`
    // writeP append
    buf = append(buf, ampmc[%s:%s]...)
`, hourIndex, hourIndex2)
	}

	if cg.offset >= 0 {
		foo := fmt.Sprintf("    // writeP codegen offset\n    _ = %s\n", hourIndex2)
		for i := 0; i < 2; i++ {
			foo += fmt.Sprintf("    buf[%d] = ampmc[%s+%d]\n", cg.offset, hourIndex, i)
			cg.offset++
		}
		return foo
	}

	return fmt.Sprintf(`
    // writeP runtime offset
    offset += copy(buf[offset:], ampmc[%s:%s])
`, hourIndex, hourIndex2)
}

func (cg *CodeGenerator) writePC() string {
	cg.isM = true
	cg.maxLength += 2

	hour := cg.gensym(1, 3, "t.Clock()")
	hourIndex := cg.gensym(1, 1, "ampmIndex[%s]", hour)
	hourIndex2 := cg.gensym(1, 1, "%s + 2", hourIndex)

	if cg.useAppend {
		return fmt.Sprintf(`
    // writePC append
    buf = append(buf, ampm[%s:%s]...)
`, hourIndex, hourIndex2)
	}

	if cg.offset >= 0 {
		foo := fmt.Sprintf("    // writePC codegen offset\n    _ = %s\n", hourIndex2)
		for i := 0; i < 2; i++ {
			foo += fmt.Sprintf("    buf[%d] = ampm[%s+%d]\n", cg.offset, hourIndex, i)
			cg.offset++
		}
		return foo
	}

	return fmt.Sprintf(`
    // writePC runtime offset
    offset += copy(buf[offset:], ampm[%s:%s])
`, hourIndex, hourIndex2)
}

func (cg *CodeGenerator) writeR() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	minute := cg.gensym(2, 3, "t.Clock()")
	second := cg.gensym(3, 3, "t.Clock()")
	hour12 := cg.gensym(1, 1, "%s %% 12", hour)

	foo := "\n    // writeR\n"
	foo += cg.write2DigitsZero(hour12)
	foo += cg.writeStringConstant(":")
	foo += cg.write2DigitsZero(minute)
	foo += cg.writeStringConstant(":")
	foo += cg.write2DigitsZero(second)
	foo += cg.writeStringConstant(" ")
	foo += cg.writeP()
	return foo
}

func (cg *CodeGenerator) writeRC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	minute := cg.gensym(2, 3, "t.Clock()")
	foo := "\n    // writeRC\n"
	foo += cg.write2DigitsZero(hour)
	foo += cg.writeStringConstant(":")
	foo += cg.write2DigitsZero(minute)
	return foo
}

func (cg *CodeGenerator) writeS() string {
	cg.libraries["strconv"] = struct{}{}
	cg.maxLength += 10

	epoch := cg.gensym(1, 1, "t.Unix()")
	epochS := cg.gensym(1, 1, "strconv.FormatInt(%s, 10)", epoch)

	return "\n    // writeS\n" + cg.writeStringValue(epochS)
}

func (cg *CodeGenerator) writeSC() string {
	second := cg.gensym(3, 3, "t.Clock()")
	return "\n    // writeSC\n" + cg.write2DigitsZero(second)
}

func (cg *CodeGenerator) writeT() string {
	return "\n    // writeT\n" + cg.writeStringConstant("\t")
}

func (cg *CodeGenerator) writeTC() string {
	hour := cg.gensym(1, 3, "t.Clock()")
	minute := cg.gensym(2, 3, "t.Clock()")
	second := cg.gensym(3, 3, "t.Clock()")
	foo := "\n    // writeTC\n"
	foo += cg.write2DigitsZero(hour)
	foo += cg.writeStringConstant(":")
	foo += cg.write2DigitsZero(minute)
	foo += cg.writeStringConstant(":")
	foo += cg.write2DigitsZero(second)
	return foo
}

func (cg *CodeGenerator) writeU() string {
	cg.isU = true
	cg.maxLength++

	wd := cg.gensym(1, 1, "t.Weekday()")
	u := cg.gensym(1, 1, "uFromWeekday[%s]", wd)

	return cg.writeStringValue(u)
}

func (cg *CodeGenerator) writeW() string {
	cg.isW = true
	cg.maxLength++

	wd := cg.gensym(1, 1, "t.Weekday()")
	w := cg.gensym(1, 1, "wFromWeekday[%s]", wd)

	return cg.writeStringValue(w)
}

func (cg *CodeGenerator) writeY() string {
	year := cg.gensym(1, 3, "t.Date()")
	year2 := cg.gensym(1, 1, "%s %% 100", year)
	return "\n    // writeY\n" + cg.write2DigitsZero(year2)
}

func (cg *CodeGenerator) writeYC() string {
	year := cg.gensym(1, 3, "t.Date()")
	return "\n    // writeYC\n" + cg.write4DigitsZero(year)
}

func (cg *CodeGenerator) writeZ() string {
	cg.maxLength++    // account for the sign
	cg.maxLength -= 4 // we only write 4 digits, even though we write code to handle digits

	zoneSeconds := cg.gensym(2, 2, "t.Zone()")

	zoneHourPositive := cg.gensym(1, 1, "%s / 3600", zoneSeconds)
	zoneMinutePositive := cg.gensym(1, 1, "%s %% 3600 / 60", zoneSeconds)

	zoneNegative := cg.gensym(1, 1, "-"+zoneSeconds)
	zoneHourNegative := cg.gensym(1, 1, "%s / 3600", zoneNegative)
	zoneMinuteNegative := cg.gensym(1, 1, "%s %% 3600 / 60", zoneNegative)

	if cg.useAppend {
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

	var off string
	if cg.offset >= 0 {
		off = fmt.Sprintf("    offset := %d // following formatting verb has variable length\n", cg.offset)
		cg.offset = -1 // must use dynamic offsets
	}

	return off + fmt.Sprintf(`
    // writeZ
    if %s >= 0 {
        buf[offset] = '+'
        offset++
        %s
        %s
    } else {
        buf[offset] = '-'
        offset++
        %s
        %s
    }
`,
		zoneSeconds,
		cg.write2DigitsZero(zoneHourPositive), cg.write2DigitsZero(zoneMinutePositive),
		cg.write2DigitsZero(zoneHourNegative), cg.write2DigitsZero(zoneMinuteNegative))
}

func (cg *CodeGenerator) writeZC() string {
	cg.maxLength += 3 // ??? can zone name be longer than 3 bytes
	zoneName := cg.gensym(1, 2, "t.Zone()")
	return cg.writeStringValue(zoneName)
}

func (cg *CodeGenerator) writeTZ() string {
	cg.maxLength += 2 // account for sign and colon
	cg.maxLength -= 4 // we only write 4 digits, even though we write code to handle digits

	zoneSeconds := cg.gensym(2, 2, "t.Zone()")

	zoneHourPositive := cg.gensym(1, 1, "%s / 3600", zoneSeconds)
	zoneMinutePositive := cg.gensym(1, 1, "%s %% 3600 / 60", zoneSeconds)

	zoneNegative := cg.gensym(1, 1, "-"+zoneSeconds)
	zoneHourNegative := cg.gensym(1, 1, "%s / 3600", zoneNegative)
	zoneMinuteNegative := cg.gensym(1, 1, "%s %% 3600 / 60", zoneNegative)

	if cg.useAppend {
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

	var off string
	if cg.offset >= 0 {
		off = fmt.Sprintf("    offset := %d // following formatting verb has variable length\n", cg.offset)
		cg.offset = -1 // must use dynamic offsets
	}

	return off + fmt.Sprintf(`
    // writeTZ
    if %s == 0 {
        buf[offset] = 'Z'
        offset++
    } else if %s > 0 {
        buf[offset] = '+'
        offset++
        %s
        buf[offset] = ':'
        offset++
        %s
    } else {
        buf[offset] = '-'
        offset++
        %s
        buf[offset] = ':'
        offset++
        %s
    }
`,
		zoneSeconds,
		zoneSeconds,
		cg.write2DigitsZero(zoneHourPositive), cg.write2DigitsZero(zoneMinutePositive),
		cg.write2DigitsZero(zoneHourNegative), cg.write2DigitsZero(zoneMinuteNegative))
}

func (cg *CodeGenerator) writePercent() string {
	return cg.writeStringConstant("%")
}

func (cg *CodeGenerator) writePlus() string {
	foo := "\n    // writePlus\n"
	foo += cg.writeWeekdayShort()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeMonthShort()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeE()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeTC()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeP()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeZC()
	foo += cg.writeStringConstant(" ")
	foo += cg.writeYC()
	return foo
}
