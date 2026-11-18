package civiltime_test

import (
	"encoding"
	"encoding/json"
	"math"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/civiltime"
)

var (
	_ encoding.TextAppender  = civiltime.CivilDate{}
	_ encoding.TextMarshaler = civiltime.CivilDate{}
	_ encoding.TextAppender  = civiltime.YearMonth{}
	_ encoding.TextMarshaler = civiltime.YearMonth{}
)

func TestValidationErrorNilReceiver(t *testing.T) {
	t.Parallel()

	var err *civiltime.ValidationError
	if got := err.Error(); got != "civiltime: validation error" {
		t.Fatalf("nil ValidationError.Error() = %q, want %q", got, "civiltime: validation error")
	}
	if civiltime.IsCode(error(err), "") {
		t.Fatal("IsCode matched typed-nil *ValidationError")
	}

	nonNil := &civiltime.ValidationError{Kind: civiltime.KindCivilDate, Code: civiltime.CodeInvalidDay}
	if got := nonNil.Error(); got != "civil_date: invalid_day" {
		t.Fatalf("ValidationError.Error() = %q, want %q", got, "civil_date: invalid_day")
	}
}

func TestTextAppenderAndWhitespaceJSON(t *testing.T) {
	t.Parallel()

	date, err := civiltime.NewCivilDate(2026, 11, 18)
	if err != nil {
		t.Fatal(err)
	}
	appendedDate, err := date.AppendText([]byte("date="))
	if err != nil || string(appendedDate) != "date=2026-11-18" {
		t.Fatalf("CivilDate.AppendText = %q, %v", appendedDate, err)
	}
	zeroDateAppended, err := (civiltime.CivilDate{}).AppendText([]byte("prefix:"))
	if err != nil || string(zeroDateAppended) != "prefix:" {
		t.Fatalf("zero CivilDate.AppendText = %q, %v", zeroDateAppended, err)
	}

	ym, err := civiltime.NewYearMonth(2026, 11)
	if err != nil {
		t.Fatal(err)
	}
	appendedYM, err := ym.AppendText([]byte("ym="))
	if err != nil || string(appendedYM) != "ym=2026-11" {
		t.Fatalf("YearMonth.AppendText = %q, %v", appendedYM, err)
	}
	zeroYMAppended, err := (civiltime.YearMonth{}).AppendText([]byte("prefix:"))
	if err != nil || string(zeroYMAppended) != "prefix:" {
		t.Fatalf("zero YearMonth.AppendText = %q, %v", zeroYMAppended, err)
	}

	decodedDate := date
	if err := decodedDate.UnmarshalJSON([]byte(" \t\nnull\r\n ")); err != nil {
		t.Fatalf("CivilDate.UnmarshalJSON(padded null) error = %v", err)
	}
	if !decodedDate.IsZero() {
		t.Fatalf("expected zero CivilDate after padded null, got %v", decodedDate)
	}

	decodedYM := ym
	if err := decodedYM.UnmarshalJSON([]byte(" \t\nnull\r\n ")); err != nil {
		t.Fatalf("YearMonth.UnmarshalJSON(padded null) error = %v", err)
	}
	if !decodedYM.IsZero() {
		t.Fatalf("expected zero YearMonth after padded null, got %v", decodedYM)
	}
}

func TestZeroValueAndBoundarySemantics(t *testing.T) {
	t.Parallel()

	var zeroDate civiltime.CivilDate
	if !zeroDate.IsZero() || zeroDate.Year() != 0 || zeroDate.Month() != 0 || zeroDate.Day() != 0 || !zeroDate.YearMonth().IsZero() {
		t.Fatalf("unexpected zero CivilDate state: %#v", zeroDate)
	}

	var zeroYM civiltime.YearMonth
	if !zeroYM.IsZero() || zeroYM.Year() != 0 || zeroYM.Month() != 0 {
		t.Fatalf("unexpected zero YearMonth state: %#v", zeroYM)
	}
	if _, err := zeroYM.AddMonths(1); !civiltime.IsCode(err, civiltime.CodeZeroValue) {
		t.Fatalf("zeroYM.AddMonths error = %v", err)
	}
	if _, err := zeroYM.FirstDay(); !civiltime.IsCode(err, civiltime.CodeZeroValue) {
		t.Fatalf("zeroYM.FirstDay error = %v", err)
	}
	if _, err := zeroYM.LastDay(); !civiltime.IsCode(err, civiltime.CodeZeroValue) {
		t.Fatalf("zeroYM.LastDay error = %v", err)
	}

	minYM, err := civiltime.NewYearMonth(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := minYM.AddMonths(-1); !civiltime.IsCode(err, civiltime.CodeOutOfRange) {
		t.Fatalf("0001-01 AddMonths(-1) = %v, want out_of_range", err)
	}
	maxYM, err := minYM.AddMonths(9999*12 - 1)
	if err != nil || maxYM.String() != "9999-12" {
		t.Fatalf("0001-01 AddMonths(max) = %v, %v", maxYM, err)
	}
	if _, err := maxYM.AddMonths(1); !civiltime.IsCode(err, civiltime.CodeOutOfRange) {
		t.Fatalf("9999-12 AddMonths(1) = %v, want out_of_range", err)
	}
}

func TestCivilDateLeapYearAndRoundTrip(t *testing.T) {
	t.Parallel()

	date, err := civiltime.ParseCivilDate("2024-02-29")
	if err != nil {
		t.Fatalf("ParseCivilDate() error = %v", err)
	}
	if date.String() != "2024-02-29" || !civiltime.IsLeapYear(date.Year()) {
		t.Fatalf("unexpected date: %s", date)
	}

	encoded, err := json.Marshal(date)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded civiltime.CivilDate
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !decoded.Equal(date) {
		t.Fatalf("round trip = %v, want %v", decoded, date)
	}
}

func TestCivilDateRejectsInvalidDates(t *testing.T) {
	t.Parallel()

	if _, err := civiltime.ParseCivilDate("2023-02-29"); !civiltime.IsCode(err, civiltime.CodeInvalidDay) {
		t.Fatalf("expected invalid day, got %v", err)
	}
	if _, err := civiltime.ParseCivilDate("2024-2-9"); !civiltime.IsCode(err, civiltime.CodeInvalidFormat) {
		t.Fatalf("expected invalid format, got %v", err)
	}
}

func TestYearMonthArithmetic(t *testing.T) {
	t.Parallel()

	month, err := civiltime.ParseYearMonth("2024-12")
	if err != nil {
		t.Fatalf("ParseYearMonth() error = %v", err)
	}
	next, err := month.AddMonths(2)
	if err != nil {
		t.Fatalf("AddMonths() error = %v", err)
	}
	if next.String() != "2025-02" {
		t.Fatalf("AddMonths() = %s", next)
	}

	lastDay, err := next.LastDay()
	if err != nil {
		t.Fatalf("LastDay() error = %v", err)
	}
	if lastDay.String() != "2025-02-28" {
		t.Fatalf("LastDay() = %s", lastDay)
	}

	if _, err := month.AddMonths(math.MaxInt); !civiltime.IsCode(err, civiltime.CodeOutOfRange) {
		t.Fatalf("expected upper bound error, got %v", err)
	}
	if _, err := month.AddMonths(math.MinInt); !civiltime.IsCode(err, civiltime.CodeOutOfRange) {
		t.Fatalf("expected lower bound error, got %v", err)
	}
}

func TestZeroValuesSerializeAsNull(t *testing.T) {
	t.Parallel()

	dateJSON, err := json.Marshal(civiltime.CivilDate{})
	if err != nil || string(dateJSON) != "null" {
		t.Fatalf("CivilDate zero JSON = %s, %v", dateJSON, err)
	}
	monthJSON, err := json.Marshal(civiltime.YearMonth{})
	if err != nil || string(monthJSON) != "null" {
		t.Fatalf("YearMonth zero JSON = %s, %v", monthJSON, err)
	}
}

func TestComparisonsDaysInMonthAndTextMarshaling(t *testing.T) {
	t.Parallel()

	d1, _ := civiltime.NewCivilDate(2026, 1, 15)
	d2, _ := civiltime.NewCivilDate(2026, 2, 1)
	if !d1.Before(d2) || !d2.After(d1) || d1.Compare(d1) != 0 || !(civiltime.CivilDate{}).Before(d1) {
		t.Fatal("CivilDate ordering mismatch")
	}
	if d1.YearMonth().String() != "2026-01" {
		t.Fatalf("YearMonth() = %s, want 2026-01", d1.YearMonth())
	}

	textDate, err := d1.MarshalText()
	if err != nil || string(textDate) != "2026-01-15" {
		t.Fatalf("CivilDate.MarshalText = %q, %v", textDate, err)
	}
	zeroDateText, err := (civiltime.CivilDate{}).MarshalText()
	if err != nil || len(zeroDateText) != 0 {
		t.Fatalf("zero CivilDate.MarshalText = %q, %v", zeroDateText, err)
	}
	if err := d1.UnmarshalText(nil); err != nil || !d1.IsZero() {
		t.Fatalf("CivilDate.UnmarshalText(empty) = %v, zero=%v", err, d1.IsZero())
	}

	ym1, _ := civiltime.NewYearMonth(2026, 1)
	ym2, _ := civiltime.NewYearMonth(2026, 3)
	if !ym1.Before(ym2) || !ym2.After(ym1) || ym1.Compare(ym1) != 0 || !ym1.Equal(ym1) || !(civiltime.YearMonth{}).Before(ym1) {
		t.Fatal("YearMonth ordering mismatch")
	}
	firstDay, err := ym1.FirstDay()
	if err != nil || firstDay.String() != "2026-01-01" {
		t.Fatalf("FirstDay = %s, %v", firstDay, err)
	}
	ymJSON, err := json.Marshal(ym1)
	if err != nil || string(ymJSON) != `"2026-01"` {
		t.Fatalf("YearMonth JSON = %s, %v", ymJSON, err)
	}
	var decodedYM civiltime.YearMonth
	if err := json.Unmarshal(ymJSON, &decodedYM); err != nil || !decodedYM.Equal(ym1) {
		t.Fatalf("YearMonth JSON round trip = %v, %v", decodedYM, err)
	}
	ymText, err := ym1.MarshalText()
	if err != nil || string(ymText) != "2026-01" {
		t.Fatalf("YearMonth.MarshalText = %q, %v", ymText, err)
	}
	zeroYMText, err := (civiltime.YearMonth{}).MarshalText()
	if err != nil || len(zeroYMText) != 0 {
		t.Fatalf("zero YearMonth.MarshalText = %q, %v", zeroYMText, err)
	}
	if err := decodedYM.UnmarshalText(nil); err != nil || !decodedYM.IsZero() {
		t.Fatalf("YearMonth.UnmarshalText(empty) = %v, zero=%v", err, decodedYM.IsZero())
	}

	if days, err := civiltime.DaysInMonth(2024, 2); err != nil || days != 29 {
		t.Fatalf("DaysInMonth(2024, 2) = %d, %v", days, err)
	}
	if _, err := civiltime.DaysInMonth(0, 2); !civiltime.IsCode(err, civiltime.CodeInvalidYear) {
		t.Fatalf("DaysInMonth invalid year = %v", err)
	}
	if _, err := civiltime.DaysInMonth(2024, 13); !civiltime.IsCode(err, civiltime.CodeInvalidMonth) {
		t.Fatalf("DaysInMonth invalid month = %v", err)
	}
}
