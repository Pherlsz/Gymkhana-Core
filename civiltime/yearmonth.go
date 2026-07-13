package civiltime

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// YearMonth is a timezone-free calendar month. Its zero value represents an
// absent value and serializes as empty text or JSON null.
type YearMonth struct {
	year  uint16
	month uint8
}

// NewYearMonth constructs a validated year-month in the year range 1..9999.
func NewYearMonth(year, month int) (YearMonth, error) {
	if year < 1 || year > 9999 {
		return YearMonth{}, validationError(KindYearMonth, CodeInvalidYear)
	}
	if month < 1 || month > 12 {
		return YearMonth{}, validationError(KindYearMonth, CodeInvalidMonth)
	}
	return YearMonth{year: uint16(year), month: uint8(month)}, nil
}

// ParseYearMonth parses exactly YYYY-MM.
func ParseYearMonth(value string) (YearMonth, error) {
	if value == "" {
		return YearMonth{}, validationError(KindYearMonth, CodeEmpty)
	}
	if len(value) != 7 || value[4] != '-' {
		return YearMonth{}, validationError(KindYearMonth, CodeInvalidFormat)
	}
	year, okYear := parseDigits(value[0:4])
	month, okMonth := parseDigits(value[5:7])
	if !okYear || !okMonth {
		return YearMonth{}, validationError(KindYearMonth, CodeInvalidFormat)
	}
	return NewYearMonth(year, month)
}

// Year returns the year, or zero for the zero value.
func (m YearMonth) Year() int { return int(m.year) }

// Month returns the month number 1..12, or zero for the zero value.
func (m YearMonth) Month() int { return int(m.month) }

// IsZero reports whether the value represents an absent year-month.
func (m YearMonth) IsZero() bool { return m == YearMonth{} }

// Compare returns -1, 0, or 1 according to chronological order. Zero sorts before non-zero values.
func (m YearMonth) Compare(other YearMonth) int {
	left := m.ordinal()
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

// Before reports whether m is before other.
func (m YearMonth) Before(other YearMonth) bool { return m.Compare(other) < 0 }

// After reports whether m is after other.
func (m YearMonth) After(other YearMonth) bool { return m.Compare(other) > 0 }

// Equal reports whether m and other represent the same year-month.
func (m YearMonth) Equal(other YearMonth) bool { return m == other }

// AddMonths returns a new value shifted by delta months without timezone logic.
func (m YearMonth) AddMonths(delta int) (YearMonth, error) {
	if m.IsZero() {
		return YearMonth{}, validationError(KindYearMonth, CodeZeroValue)
	}

	base := int64(m.year-1)*12 + int64(m.month-1)
	maximum := int64(9999*12 - 1)
	shift := int64(delta)
	if shift < -base || shift > maximum-base {
		return YearMonth{}, validationError(KindYearMonth, CodeOutOfRange)
	}
	index := base + shift

	year := int(index/12) + 1
	month := int(index%12) + 1
	return NewYearMonth(year, month)
}

// FirstDay returns the first civil date of the month.
func (m YearMonth) FirstDay() (CivilDate, error) {
	if m.IsZero() {
		return CivilDate{}, validationError(KindYearMonth, CodeZeroValue)
	}
	return NewCivilDate(int(m.year), int(m.month), 1)
}

// LastDay returns the last civil date of the month.
func (m YearMonth) LastDay() (CivilDate, error) {
	if m.IsZero() {
		return CivilDate{}, validationError(KindYearMonth, CodeZeroValue)
	}
	return NewCivilDate(int(m.year), int(m.month), daysInMonthUnchecked(int(m.year), int(m.month)))
}

func (m YearMonth) String() string {
	if m.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d", m.year, m.month)
}

func (m YearMonth) ordinal() int {
	return int(m.year)*100 + int(m.month)
}

// MarshalText implements encoding.TextMarshaler.
func (m YearMonth) MarshalText() ([]byte, error) { return []byte(m.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler. Empty input resets to zero.
func (m *YearMonth) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*m = YearMonth{}
		return nil
	}
	parsed, err := ParseYearMonth(string(text))
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

// MarshalJSON emits JSON null for zero or an ISO year-month string otherwise.
func (m YearMonth) MarshalJSON() ([]byte, error) {
	if m.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(m.String())
}

// UnmarshalJSON accepts JSON null, an empty string, or a strict ISO year-month.
func (m *YearMonth) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*m = YearMonth{}
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return validationError(KindYearMonth, CodeInvalidFormat)
	}
	return m.UnmarshalText([]byte(value))
}

func parseDigits(value string) (int, bool) {
	result := 0
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return 0, false
		}
		result = result*10 + int(value[i]-'0')
	}
	return result, true
}
