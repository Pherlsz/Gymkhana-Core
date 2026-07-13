package civiltime_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/civiltime"
)

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
