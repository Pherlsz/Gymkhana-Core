package fingerprint_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/fingerprint"
)

func BenchmarkSumText(b *testing.B) {
	const input = "The quick brown fox jumps over the lazy dog"
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()
	for range b.N {
		if _, err := fingerprint.SumText(input); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSumFramedText(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := fingerprint.SumFramedText("query.plan/v1", "alpha", "beta", "gamma", "delta"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDigest(b *testing.B) {
	const hexDigest = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	b.ReportAllocs()
	for range b.N {
		if _, err := fingerprint.Parse(hexDigest); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDigestAppendText(b *testing.B) {
	digest, err := fingerprint.SumText("abc")
	if err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, 0, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		var err error
		buf, err = digest.AppendText(buf[:0])
		if err != nil {
			b.Fatal(err)
		}
	}
}
