package fingerprint_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/fingerprint"
)

func FuzzFingerprintParse(f *testing.F) {
	for _, seed := range []string{
		"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD",
		"",
		"abc",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 256 {
			return
		}

		digest, err := fingerprint.Parse(value)
		if err != nil {
			return
		}
		again, err := fingerprint.Parse(digest.String())
		if err != nil {
			t.Fatalf("canonical digest rejected: %v", err)
		}
		if again != digest {
			t.Fatalf("digest round trip changed value: %s != %s", again, digest)
		}
	})
}
