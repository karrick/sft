package main

import (
    "fmt"
    "strconv"
    "time"
)

func main() {
    // fmt.Println(string(appendTime(nil, time.Now())))
	fmt.Println(string(appendTime(nil, time.Date(2021, time.August, 22, 0, 0, 0, 1, time.UTC))))
}

func appendTime(buf []byte, t time.Time) []byte {
    // situational constant initializations
    const ampm = "ampm"
    var ampmIndex = []int{0,0,0,0,0,0,0,0,0,0,0,0,2,2,2,2,2,2,2,2,2,2,2,2}
    const ampmc = "AMPM"
    const digits = "0123456789 123456789"
    var quotient, remainder int
    const weekdaysLong = "SundayMondayTuesdayWednesdayThursdayFridaySaturday"
    var weekdaysLongIndices = []int{0, 6, 12, 19, 28, 36, 42, 50}
    const monthsLong = "JanuaryFebruaryMarchAprilMayJuneJulyAugustSeptemberOctoberNovemberDecember"
    var monthsLongIndices = []int{0, 7, 15, 20, 25, 28, 32, 36, 42, 51, 58, 66, 74}
    var uFromWeekday = []byte{'7', '1', '2', '3', '4', '5', '6'}
    var wFromWeekday = []byte{'0', '1', '2', '3', '4', '5', '6'}
    // dynamically generated variable initializations
    gs0 := t.Weekday()
    gs1 := weekdaysLongIndices[gs0]
    gs2 := gs1 + 3
    gs3 := gs0 + 1
    gs4 := weekdaysLongIndices[gs3]
    gs14, gs5, gs10 := t.Date()
    gs6 := gs5 - 1
    gs7 := monthsLongIndices[gs6]
    gs8 := gs7 + 3
    gs9 := monthsLongIndices[gs5]
    gs11, gs12, gs13 := t.Clock()
    gs15 := gs14 / 100
    gs16 := int(gs5)
    gs17 := gs14 % 100
    gs18, _ := t.ISOWeek()
    gs19 := gs18 % 100
    gs20 := gs11 % 12
    gs21 := t.YearDay()
    gs22 := t.Nanosecond()
    gs23 := ampmIndex[gs11]
    gs24 := gs23 + 2
    gs25 := t.Unix()
    gs26 := uFromWeekday[gs0]
    gs27 := wFromWeekday[gs0]
    gs34, gs28 := t.Zone()
    gs29 := gs28 / 3600
    gs30 := gs28 % 3600 / 60
    gs31 := -gs28
    gs32 := gs31 / 3600
    gs33 := gs31 % 3600 / 60
    gs35 := gs22 / 1000000
    gs36 := gs22 / 1000

    // Append Weekday Short
	buf = append(buf, weekdaysLong[gs1:gs2]...)
    buf = append(buf, "<->"...)

    // Append Weekday Long
	buf = append(buf, weekdaysLong[gs1:gs4]...)
    buf = append(buf, ' ')

    // Append Month Short
    buf = append(buf, monthsLong[gs7:gs8]...)
    buf = append(buf, ' ')

    // Append Month Long
	buf = append(buf, monthsLong[gs7:gs9]...)
    buf = append(buf, ' ')

    // writeC

    // Append Weekday Short
	buf = append(buf, weekdaysLong[gs1:gs2]...)
    buf = append(buf, ' ')

    // Append Month Short
    buf = append(buf, monthsLong[gs7:gs8]...)
    buf = append(buf, ' ')

    // writeE
    // write2DigitsSpace
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[10+quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeTC
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeYC
    // write4DigitsZero
    // thousands
	quotient = gs14 / 1000
	remainder = gs14 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeCC
    // write2DigitsZero
    quotient = gs15 / 10
	remainder = gs15 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeD
    // write2DigitsZero
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeDC
    // write2DigitsZero
    quotient = gs16 / 10
	remainder = gs16 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, '/')
    // write2DigitsZero
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, '/')
    // write2DigitsZero
    quotient = gs17 / 10
	remainder = gs17 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeE
    // write2DigitsSpace
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[10+quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeFC
    // write4DigitsZero
    // thousands
	quotient = gs14 / 1000
	remainder = gs14 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, '-')
    // write2DigitsZero
    quotient = gs16 / 10
	remainder = gs16 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, '-')
    // write2DigitsZero
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeG
    // write2DigitsZero
    quotient = gs19 / 10
	remainder = gs19 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeGC
    // write4DigitsZero
    // thousands
	quotient = gs18 / 1000
	remainder = gs18 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // Append Month Short
    buf = append(buf, monthsLong[gs7:gs8]...)
    buf = append(buf, ' ')

    // writeHC
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeIC
    // write2DigitsZero
    quotient = gs20 / 10
	remainder = gs20 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeJ
    // write3DigitsZero
	// hundreds
	quotient = gs21 / 100
	remainder = gs21 % 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeK
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeL
    // write2DigitsSpace
    quotient = gs20 / 10
	remainder = gs20 % 10
	buf = append(buf, digits[10+quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeM
    // write2DigitsZero
    quotient = gs16 / 10
	remainder = gs16 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeMC
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeN
    buf = append(buf, '\n')
    buf = append(buf, ' ')

    // writeNC
    // write9DigitsZero
	// hundred-millions
	quotient = gs22 / 100000000
	remainder = gs22 % 100000000
	buf = append(buf, digits[quotient])
	// ten-millions
	quotient = remainder / 10000000
	remainder = gs22 % 10000000
	buf = append(buf, digits[quotient])
	// millions
	quotient = remainder / 1000000
	remainder = gs22 % 1000000
	buf = append(buf, digits[quotient])
	// hundred-thousands
	quotient = remainder / 100000
	remainder = gs22 % 100000
	buf = append(buf, digits[quotient])
	// ten-thousands
	quotient = remainder / 10000
	remainder = gs22 % 10000
	buf = append(buf, digits[quotient])
	// thousands
	quotient = remainder / 1000
	remainder = gs22 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeP
    buf = append(buf, ampmc[gs23:gs24]...)
    buf = append(buf, ' ')

    // writePC
    buf = append(buf, ampm[gs23:gs24]...)
    buf = append(buf, ' ')

    // writeR
    // write2DigitsZero
    quotient = gs20 / 10
	remainder = gs20 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeP
    buf = append(buf, ampmc[gs23:gs24]...)
    buf = append(buf, ' ')

    // writeRC
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')
    buf = strconv.AppendInt(buf, gs25, 10) // writeS
    buf = append(buf, ' ')

    // writeSC
    // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeT
    buf = append(buf, '\t')
    buf = append(buf, ' ')

    // writeTC
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')
    // writeU
    buf = append(buf, gs26)
    buf = append(buf, ' ')
    // writeW
    buf = append(buf, gs27)
    buf = append(buf, ' ')

    // writeDC
    // write2DigitsZero
    quotient = gs16 / 10
	remainder = gs16 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, '/')
    // write2DigitsZero
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, '/')
    // write2DigitsZero
    quotient = gs17 / 10
	remainder = gs17 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeTC
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeY
    // write2DigitsZero
    quotient = gs17 / 10
	remainder = gs17 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeYC
    // write4DigitsZero
    // thousands
	quotient = gs14 / 1000
	remainder = gs14 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeZ
    if gs28 >= 0 {
        buf = append(buf, '+')
            // write2DigitsZero
    quotient = gs29 / 10
	remainder = gs29 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

            // write2DigitsZero
    quotient = gs30 / 10
	remainder = gs30 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

    } else {
        buf = append(buf, '-')
            // write2DigitsZero
    quotient = gs32 / 10
	remainder = gs32 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

            // write2DigitsZero
    quotient = gs33 / 10
	remainder = gs33 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

    }
    buf = append(buf, ' ')
    buf = append(buf, gs34...)
    // writeZC
    buf = append(buf, ' ')
    buf = append(buf, '%')
    buf = append(buf, ' ')

    // writePlus

    // Append Weekday Short
	buf = append(buf, weekdaysLong[gs1:gs2]...)
    buf = append(buf, ' ')

    // Append Month Short
    buf = append(buf, monthsLong[gs7:gs8]...)
    buf = append(buf, ' ')

    // writeE
    // write2DigitsSpace
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[10+quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeTC
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs12 / 10
	remainder = gs12 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeP
    buf = append(buf, ampmc[gs23:gs24]...)
    buf = append(buf, ' ')
    buf = append(buf, gs34...)
    // writeZC
    buf = append(buf, ' ')

    // writeYC
    // write4DigitsZero
    // thousands
	quotient = gs14 / 1000
	remainder = gs14 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeTZ
    if gs28 == 0 {
        buf = append(buf, 'Z')
    } else if gs28 > 0 {
        buf = append(buf, '+')
            // write2DigitsZero
    quotient = gs29 / 10
	remainder = gs29 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

        buf = append(buf, ':')
            // write2DigitsZero
    quotient = gs30 / 10
	remainder = gs30 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

    } else {
        buf = append(buf, '-')
            // write2DigitsZero
    quotient = gs32 / 10
	remainder = gs32 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

        buf = append(buf, ':')
            // write2DigitsZero
    quotient = gs33 / 10
	remainder = gs33 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

    }
    buf = append(buf, ' ')

    // writeLMin
    // write2DigitsMin
    quotient = gs20 / 10
	remainder = gs20 % 10
	if quotient > 0 {
		buf = append(buf, digits[quotient])
	}
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeMillis
    // write3DigitsZero
	// hundreds
	quotient = gs35 / 100
	remainder = gs35 % 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeMicro
    // write6DigitsZero
	// hundred-thousands
	quotient = gs36 / 100000
	remainder = gs36 % 100000
	buf = append(buf, digits[quotient])
	// ten-thousands
	quotient = remainder / 10000
	remainder = gs36 % 10000
	buf = append(buf, digits[quotient])
	// thousands
	quotient = remainder / 1000
	remainder = gs36 % 1000
	buf = append(buf, digits[quotient])
	// hundreds
	quotient = remainder / 100
	remainder %= 100
	buf = append(buf, digits[quotient])
	// tens
	quotient = remainder / 10
	remainder %= 10
	buf = append(buf, digits[quotient])
	// ones
	buf = append(buf, digits[remainder])

    return buf
}
