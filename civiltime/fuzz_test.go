package civiltime_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/civiltime"
)

func FuzzParseCivilDate(f *testing.F) {
	for _, seed := range []string{"2024-02-29", "2023-02-29", "0000-01-01", ""} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 64 {
			return
		}

		date, err := civiltime.ParseCivilDate(value)
		if err == nil {
			reparsed, reparsedErr := civiltime.ParseCivilDate(date.String())
			if reparsedErr != nil || !reparsed.Equal(date) {
				t.Fatalf("date round trip failed: %v, %v", reparsed, reparsedErr)
			}
		}
	})
}

func FuzzParseYearMonth(f *testing.F) {
	for _, seed := range []string{"2024-12", "2024-13", "0000-01", ""} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 64 {
			return
		}

		month, err := civiltime.ParseYearMonth(value)
		if err == nil {
			reparsed, reparsedErr := civiltime.ParseYearMonth(month.String())
			if reparsedErr != nil || !reparsed.Equal(month) {
				t.Fatalf("year-month round trip failed: %v, %v", reparsed, reparsedErr)
			}
		}
	})
}
