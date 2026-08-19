package fingerprint_test

import (
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/fingerprint"
)

func TestSumTextSHA256(t *testing.T) {
	t.Parallel()

	digest, err := fingerprint.SumText("abc")
	if err != nil {
		t.Fatalf("SumText() error = %v", err)
	}
	got := digest.String()
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("SumText() = %q, want %q", got, want)
	}
}

func TestSumFramedTextPortableVector(t *testing.T) {
	t.Parallel()

	got, err := fingerprint.SumFramedText("query.plan/v1", "alpha", "beta")
	if err != nil {
		t.Fatalf("SumFramedText() error = %v", err)
	}
	const want = "8f92cbf0d8c72b614311eadbf43c8842fc53e77f5a3943448db7a1f12ea470fd"
	if got.String() != want {
		t.Fatalf("SumFramedText() = %q, want %q", got, want)
	}
}

func TestFramedPreservesBoundariesAndNamespace(t *testing.T) {
	t.Parallel()

	left, err := fingerprint.SumFramedText("query.plan/v1", "ab", "c")
	if err != nil {
		t.Fatal(err)
	}
	right, err := fingerprint.SumFramedText("query.plan/v1", "a", "bc")
	if err != nil {
		t.Fatal(err)
	}
	otherNamespace, err := fingerprint.SumFramedText("task.spec/v1", "ab", "c")
	if err != nil {
		t.Fatal(err)
	}

	if left == right {
		t.Fatal("framed fingerprints ignore part boundaries")
	}
	if left == otherNamespace {
		t.Fatal("framed fingerprints ignore namespace")
	}
}

func TestFramedDistinguishesZeroPartsFromEmptyPart(t *testing.T) {
	t.Parallel()

	zero, err := fingerprint.SumFramedText("empty/v1")
	if err != nil {
		t.Fatal(err)
	}
	empty, err := fingerprint.SumFramedText("empty/v1", "")
	if err != nil {
		t.Fatal(err)
	}
	if zero.String() != "5a57cd4806fb6e845a350d798fe9e1346bfe2f22c0cea0e540c7bd1e0c7b0e08" {
		t.Fatalf("zero-part frame = %s", zero)
	}
	if empty.String() != "5d3c8cdd3b8bb65bb7b0192533816ff98df5ef007f3d9d385320ec12582d4e1f" {
		t.Fatalf("empty-part frame = %s", empty)
	}
	if zero == empty {
		t.Fatal("zero-part frame equals one-empty-part frame")
	}
}

func TestFramedBytesMatchesText(t *testing.T) {
	t.Parallel()

	textDigest, err := fingerprint.SumFramedText("example/v1", "João", "東京")
	if err != nil {
		t.Fatal(err)
	}
	bytesDigest, err := fingerprint.SumFramed("example/v1", []byte("João"), []byte("東京"))
	if err != nil {
		t.Fatal(err)
	}
	if textDigest != bytesDigest {
		t.Fatalf("text and byte framing differ: %s != %s", textDigest, bytesDigest)
	}
}

func TestTextAPIsRejectInvalidUTF8(t *testing.T) {
	t.Parallel()

	invalid := string([]byte{0xff, 'a'})
	if _, err := fingerprint.SumText(invalid); !fingerprint.IsCode(err, fingerprint.CodeInvalidUTF8) {
		t.Fatalf("SumText invalid UTF-8 error = %v", err)
	}
	if _, err := fingerprint.SumFramedText("example/v1", invalid); !fingerprint.IsCode(err, fingerprint.CodeInvalidUTF8) {
		t.Fatalf("SumFramedText invalid UTF-8 error = %v", err)
	}

	// Arbitrary binary data remains valid through the byte API.
	if _, err := fingerprint.SumFramed("example/v1", []byte{0xff, 'a'}); err != nil {
		t.Fatalf("SumFramed binary input error = %v", err)
	}
}

func TestFramedRejectsInvalidNamespace(t *testing.T) {
	t.Parallel()

	for _, namespace := range []string{"", "Upper/v1", "with space", "ümlaut/v1", strings.Repeat("a", fingerprint.MaxNamespaceBytes+1)} {
		namespace := namespace
		t.Run(namespace, func(t *testing.T) {
			t.Parallel()
			_, err := fingerprint.SumFramedText(namespace, "value")
			if !fingerprint.IsCode(err, fingerprint.CodeInvalidNamespace) {
				t.Fatalf("error = %v, want invalid_namespace", err)
			}
		})
	}
}

func TestDigestParseCanonicalizesHexCase(t *testing.T) {
	t.Parallel()

	const lower = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	parsed, err := fingerprint.Parse(strings.ToUpper(lower))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.String() != lower {
		t.Fatalf("Parse().String() = %q, want %q", parsed, lower)
	}

	encoded, err := parsed.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var decoded fingerprint.Digest
	if err := decoded.UnmarshalText(encoded); err != nil {
		t.Fatal(err)
	}
	if decoded != parsed {
		t.Fatalf("text round trip changed digest: %s != %s", decoded, parsed)
	}
}

func TestDigestParseRejectsMalformedValues(t *testing.T) {
	t.Parallel()

	if _, err := fingerprint.Parse("abc"); !fingerprint.IsCode(err, fingerprint.CodeInvalidLength) {
		t.Fatalf("short digest error = %v", err)
	}
	if _, err := fingerprint.Parse(strings.Repeat("z", 64)); !fingerprint.IsCode(err, fingerprint.CodeInvalidFormat) {
		t.Fatalf("invalid hex error = %v", err)
	}
}

func TestDigestBytesReturnsCopy(t *testing.T) {
	t.Parallel()

	digest, err := fingerprint.SumText("abc")
	if err != nil {
		t.Fatal(err)
	}
	value := digest.Bytes()
	value[0] ^= 0xff
	if digest.String() != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("Digest.Bytes exposed mutable digest storage")
	}
}
