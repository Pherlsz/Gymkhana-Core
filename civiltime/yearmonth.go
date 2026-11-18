package civiltime

import (
	"bytes"
	"encoding/json"
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
	var buf [7]byte
	m.formatInto(&buf)
	return string(buf[:])
}

func (m YearMonth) formatInto(buf *[7]byte) {
	y := m.year
	buf[0] = byte('0' + y/1000)
	buf[1] = byte('0' + (y/100)%10)
	buf[2] = byte('0' + (y/10)%10)
	buf[3] = byte('0' + y%10)
	buf[4] = '-'
	mo := m.month
	buf[5] = byte('0' + mo/10)
	buf[6] = byte('0' + mo%10)
}

func (m YearMonth) ordinal() int {
	return int(m.year)*100 + int(m.month)
}

// AppendText implements encoding.TextAppender.
func (m YearMonth) AppendText(b []byte) ([]byte, error) {
	if m.IsZero() {
		return b, nil
	}
	var buf [7]byte
	m.formatInto(&buf)
	return append(b, buf[:]...), nil
}

// MarshalText implements encoding.TextMarshaler.
func (m YearMonth) MarshalText() ([]byte, error) {
	if m.IsZero() {
		return []byte{}, nil
	}
	var buf [7]byte
	m.formatInto(&buf)
	return buf[:], nil
}

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
	var buf [9]byte
	buf[0] = '"'
	m.formatInto((*[7]byte)(buf[1:8]))
	buf[8] = '"'
	return buf[:], nil
}

// UnmarshalJSON accepts JSON null, an empty string, or a strict ISO year-month.
func (m *YearMonth) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*m = YearMonth{}
		return nil
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
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
