package fingerprint_test

import (
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/fingerprint"
)

func TestStringSHA256(t *testing.T) {
	t.Parallel()

	got := fingerprint.String("abc").String()
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestFramedStringsPortableVector(t *testing.T) {
	t.Parallel()

	got, err := fingerprint.FramedStrings("query.plan/v1", "alpha", "beta")
	if err != nil {
		t.Fatalf("FramedStrings() error = %v", err)
	}
	const want = "8f92cbf0d8c72b614311eadbf43c8842fc53e77f5a3943448db7a1f12ea470fd"
	if got.String() != want {
		t.Fatalf("FramedStrings() = %q, want %q", got, want)
	}
}

func TestFramedPreservesBoundariesAndNamespace(t *testing.T) {
	t.Parallel()

	left, err := fingerprint.FramedStrings("query.plan/v1", "ab", "c")
	if err != nil {
		t.Fatal(err)
	}
	right, err := fingerprint.FramedStrings("query.plan/v1", "a", "bc")
	if err != nil {
		t.Fatal(err)
	}
	otherNamespace, err := fingerprint.FramedStrings("task.spec/v1", "ab", "c")
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

func TestFramedBytesMatchesStrings(t *testing.T) {
	t.Parallel()

	stringsDigest, err := fingerprint.FramedStrings("example/v1", "João", "東京")
	if err != nil {
		t.Fatal(err)
	}
	bytesDigest, err := fingerprint.Framed("example/v1", []byte("João"), []byte("東京"))
	if err != nil {
		t.Fatal(err)
	}
	if stringsDigest != bytesDigest {
		t.Fatalf("string and byte framing differ: %s != %s", stringsDigest, bytesDigest)
	}
}

func TestFramedRejectsInvalidNamespace(t *testing.T) {
	t.Parallel()

	for _, namespace := range []string{"", "Upper/v1", "with space", "ümlaut/v1", strings.Repeat("a", fingerprint.MaxNamespaceBytes+1)} {
		namespace := namespace
		t.Run(namespace, func(t *testing.T) {
			t.Parallel()
			_, err := fingerprint.FramedStrings(namespace, "value")
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

	digest := fingerprint.String("abc")
	value := digest.Bytes()
	value[0] ^= 0xff
	if digest.String() != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("Digest.Bytes exposed mutable digest storage")
	}
}
