package civiltime

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CivilDate is a timezone-free calendar date. Its zero value represents an
// absent date and serializes as empty text or JSON null.
type CivilDate struct {
	year  uint16
	month uint8
	day   uint8
}

// NewCivilDate constructs a validated date in the inclusive year range 1..9999.
func NewCivilDate(year, month, day int) (CivilDate, error) {
	if year < 1 || year > 9999 {
		return CivilDate{}, validationError(KindCivilDate, CodeInvalidYear)
	}
	if month < 1 || month > 12 {
		return CivilDate{}, validationError(KindCivilDate, CodeInvalidMonth)
	}
	maximumDay := daysInMonthUnchecked(year, month)
	if day < 1 || day > maximumDay {
		return CivilDate{}, validationError(KindCivilDate, CodeInvalidDay)
	}

	return CivilDate{year: uint16(year), month: uint8(month), day: uint8(day)}, nil
}

// ParseCivilDate parses exactly YYYY-MM-DD.
func ParseCivilDate(value string) (CivilDate, error) {
	if value == "" {
		return CivilDate{}, validationError(KindCivilDate, CodeEmpty)
	}
	if len(value) != 10 || value[4] != '-' || value[7] != '-' {
		return CivilDate{}, validationError(KindCivilDate, CodeInvalidFormat)
	}

	year, okYear := parseDigits(value[0:4])
	month, okMonth := parseDigits(value[5:7])
	day, okDay := parseDigits(value[8:10])
	if !okYear || !okMonth || !okDay {
		return CivilDate{}, validationError(KindCivilDate, CodeInvalidFormat)
	}

	return NewCivilDate(year, month, day)
}

// Year returns the four-digit year, or zero for the zero value.
func (d CivilDate) Year() int { return int(d.year) }

// Month returns the month number 1..12, or zero for the zero value.
func (d CivilDate) Month() int { return int(d.month) }

// Day returns the day number, or zero for the zero value.
func (d CivilDate) Day() int { return int(d.day) }

// IsZero reports whether the value represents an absent date.
func (d CivilDate) IsZero() bool { return d == CivilDate{} }

// Compare returns -1, 0, or 1 according to chronological order. Zero sorts before non-zero values.
func (d CivilDate) Compare(other CivilDate) int {
	left := d.ordinal()
	right := other.ordinal()
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

// Before reports whether d is chronologically before other.
func (d CivilDate) Before(other CivilDate) bool { return d.Compare(other) < 0 }

// After reports whether d is chronologically after other.
func (d CivilDate) After(other CivilDate) bool { return d.Compare(other) > 0 }

// Equal reports whether d and other represent the same civil date.
func (d CivilDate) Equal(other CivilDate) bool { return d == other }

// YearMonth returns the date's containing year-month, or zero for a zero date.
func (d CivilDate) YearMonth() YearMonth {
	if d.IsZero() {
		return YearMonth{}
	}
	return YearMonth{year: d.year, month: d.month}
}

func (d CivilDate) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (d CivilDate) ordinal() int {
	return int(d.year)*10000 + int(d.month)*100 + int(d.day)
}

// MarshalText implements encoding.TextMarshaler.
func (d CivilDate) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Empty input resets to zero.
func (d *CivilDate) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*d = CivilDate{}
		return nil
	}
	parsed, err := ParseCivilDate(string(text))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalJSON emits JSON null for zero or an ISO date string otherwise.
func (d CivilDate) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON accepts JSON null, an empty string, or a strict ISO date string.
func (d *CivilDate) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*d = CivilDate{}
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return validationError(KindCivilDate, CodeInvalidFormat)
	}
	return d.UnmarshalText([]byte(value))
}

// IsLeapYear reports whether year follows Gregorian leap-year rules.
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in a validated year and month.
func DaysInMonth(year, month int) (int, error) {
	if year < 1 || year > 9999 {
		return 0, validationError(KindCivilDate, CodeInvalidYear)
	}
	if month < 1 || month > 12 {
		return 0, validationError(KindCivilDate, CodeInvalidMonth)
	}
	return daysInMonthUnchecked(year, month), nil
}

func daysInMonthUnchecked(year, month int) int {
	days := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if month == 2 && IsLeapYear(year) {
		return 29
	}
	return days[month-1]
}
