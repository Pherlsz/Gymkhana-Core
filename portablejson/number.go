package portablejson

import "encoding/json"

// EqualNumbers reports exact mathematical equality for two portable JSON
// number lexemes. It returns false when either value falls outside the
// portable numeric bounds enforced by this package.
func EqualNumbers(left, right json.Number) bool {
	if !json.Valid([]byte(left.String())) || !json.Valid([]byte(right.String())) {
		return false
	}
	leftValue, leftErr := parsePortableNumber(left.String())
	rightValue, rightErr := parsePortableNumber(right.String())
	return leftErr == nil && rightErr == nil && leftValue.Cmp(rightValue) == 0
}
