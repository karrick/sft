package main

import (
    "fmt"
    "time"
)

func main() {
    // fmt.Println(string(appendTime(nil, time.Now())))
	fmt.Println(string(appendTime(nil, time.Date(2021, time.August, 22, 0, 0, 0, 1, time.UTC))))
}

func appendTime(buf []byte, t time.Time) []byte {
    // situational constant initializations
    const digits = "0123456789 123456789"
    var quotient, remainder int
    const weekdaysLong = "SundayMondayTuesdayWednesdayThursdayFridaySaturday"
    var weekdaysLongIndices = []int{0, 6, 12, 19, 28, 36, 42, 50}
    const monthsLong = "JanuaryFebruaryMarchAprilMayJuneJulyAugustSeptemberOctoberNovemberDecember"
    var monthsLongIndices = []int{0, 7, 15, 20, 25, 28, 32, 36, 42, 51, 58, 66, 74}
    // dynamically generated variable initializations
    gs0 := t.Weekday()
    gs1 := weekdaysLongIndices[gs0]
    gs2 := gs1 + 3
    gs8, gs4, gs3 := t.Date()
    gs5 := gs4 - 1
    gs6 := monthsLongIndices[gs5]
    gs7 := gs6 + 3
    gs9, gs10, gs11 := t.Clock()
    _, gs12 := t.Zone()
    gs13 := gs12 / 3600
    gs14 := gs12 % 3600 / 60
    gs15 := -gs12
    gs16 := gs15 / 3600
    gs17 := gs15 % 3600 / 60

    // Append Weekday Short
	buf = append(buf, weekdaysLong[gs1:gs2]...)
    buf = append(buf, ", "...)

    // writeD
    // write2DigitsZero
    quotient = gs3 / 10
	remainder = gs3 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // Append Month Short
    buf = append(buf, monthsLong[gs6:gs7]...)
    buf = append(buf, ' ')

    // writeYC
    // write4DigitsZero
    // thousands
	quotient = gs8 / 1000
	remainder = gs8 % 1000
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

    // writeTC
    // write2DigitsZero
    quotient = gs9 / 10
	remainder = gs9 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs10 / 10
	remainder = gs10 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ':')
    // write2DigitsZero
    quotient = gs11 / 10
	remainder = gs11 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])
    buf = append(buf, ' ')

    // writeZ
    if gs12 >= 0 {
        buf = append(buf, '+')
            // write2DigitsZero
    quotient = gs13 / 10
	remainder = gs13 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

            // write2DigitsZero
    quotient = gs14 / 10
	remainder = gs14 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

    } else {
        buf = append(buf, '-')
            // write2DigitsZero
    quotient = gs16 / 10
	remainder = gs16 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

            // write2DigitsZero
    quotient = gs17 / 10
	remainder = gs17 % 10
	buf = append(buf, digits[quotient])
	buf = append(buf, digits[remainder])

    }

    return buf
}
