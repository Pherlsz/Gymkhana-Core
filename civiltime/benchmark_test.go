package civiltime_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/civiltime"
)

func BenchmarkParseCivilDate(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := civiltime.ParseCivilDate("2026-11-18"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCivilDateAppendText(b *testing.B) {
	date, err := civiltime.NewCivilDate(2026, 11, 18)
	if err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 0, 16)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		var err error
		buf, err = date.AppendText(buf[:0])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCivilDateMarshalJSON(b *testing.B) {
	date, err := civiltime.NewCivilDate(2026, 11, 18)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := date.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYearMonthAppendText(b *testing.B) {
	ym, err := civiltime.NewYearMonth(2026, 11)
	if err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 0, 16)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		var err error
		buf, err = ym.AppendText(buf[:0])
		if err != nil {
			b.Fatal(err)
		}
	}
}
