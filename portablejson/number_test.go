package portablejson

import (
	"encoding/json"
	"testing"
)

func TestEqualNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		left, right json.Number
		want        bool
	}{
		{name: "integer and decimal lexical forms", left: "1", right: "1.0", want: true},
		{name: "distinct exact decimals", left: "0.1", right: "0.10000000000000001", want: false},
		{name: "excessive exponent", left: "1e1000000", right: "1", want: false},
		{name: "outside safe magnitude", left: "9007199254740992", right: "9007199254740992", want: false},
		{name: "invalid JSON numeric lexeme", left: "01", right: "1", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := EqualNumbers(test.left, test.right); got != test.want {
				t.Fatalf("EqualNumbers(%q, %q) = %v, want %v", test.left, test.right, got, test.want)
			}
		})
	}
}
