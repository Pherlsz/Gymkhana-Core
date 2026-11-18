package portablejson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

const (
	// MaxBytes is the maximum portable JSON payload accepted by the generic
	// validation layer. Higher-level domains may impose smaller limits.
	MaxBytes = 1 << 20
	// MaxDepth bounds recursive JSON containers.
	MaxDepth = 64
	// MaxNodes bounds work performed while validating one JSON payload.
	MaxNodes = 100000
	// MaxSafeInteger is the largest integer exactly representable by JSON-backed
	// JavaScript/TypeScript runtimes.
	MaxSafeInteger int64 = 1<<53 - 1
)

// ValidateObject validates one JSON object without normalizing it. It rejects
// invalid UTF-8, duplicate object keys at any depth, unpaired UTF-16 surrogate
// escapes, excessive size/depth/node counts, and trailing JSON values.
func ValidateObject(raw json.RawMessage) error {
	return validateJSON(raw, true)
}

// ValidateValue validates one arbitrary JSON value with the same lexical and
// resource bounds as ValidateObject.
func ValidateValue(raw json.RawMessage) error {
	return validateJSON(raw, false)
}

func validateJSON(raw json.RawMessage, requireObject bool) error {
	if len(raw) == 0 || len(raw) > MaxBytes || !utf8.Valid(raw) || !validJSONUnicodeEscapes(raw) {
		return validationError(CodeInvalidJSON, "json")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	nodes := 0
	if _, err := scanOrBuildJSONValue(decoder, 0, &nodes, requireObject, false); err != nil {
		return validationError(CodeInvalidJSON, "json")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return validationError(CodeInvalidJSON, "json")
	}
	return nil
}

func decodeStrictValue(raw json.RawMessage, requireObject bool) (any, error) {
	if len(raw) == 0 || len(raw) > MaxBytes || !utf8.Valid(raw) || !validJSONUnicodeEscapes(raw) {
		return nil, validationError(CodeInvalidJSON, "json")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	nodes := 0
	value, err := scanOrBuildJSONValue(decoder, 0, &nodes, requireObject, true)
	if err != nil {
		return nil, validationError(CodeInvalidJSON, "json")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, validationError(CodeInvalidJSON, "json")
	}
	return value, nil
}

func scanOrBuildJSONValue(decoder *json.Decoder, depth int, nodes *int, requireObject bool, build bool) (any, error) {
	if depth > MaxDepth || *nodes >= MaxNodes {
		return nil, errors.New("json bound exceeded")
	}
	*nodes = *nodes + 1
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, isDelim := token.(json.Delim)
	if requireObject && (!isDelim || delim != '{') {
		return nil, errors.New("object required")
	}
	if !isDelim {
		return token, nil
	}

	switch delim {
	case '{':
		var obj map[string]any
		var seen map[string]struct{}
		if build {
			obj = make(map[string]any)
		} else {
			seen = make(map[string]struct{})
		}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok || !utf8.ValidString(key) {
				return nil, errors.New("invalid object key")
			}
			if build {
				if _, duplicate := obj[key]; duplicate {
					return nil, errors.New("duplicate object key")
				}
			} else {
				if _, duplicate := seen[key]; duplicate {
					return nil, errors.New("duplicate object key")
				}
				seen[key] = struct{}{}
			}
			child, err := scanOrBuildJSONValue(decoder, depth+1, nodes, false, build)
			if err != nil {
				return nil, err
			}
			if build {
				obj[key] = child
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return nil, errors.New("invalid object ending")
		}
		return obj, nil
	case '[':
		var arr []any
		if build {
			arr = []any{}
		}
		for decoder.More() {
			child, err := scanOrBuildJSONValue(decoder, depth+1, nodes, false, build)
			if err != nil {
				return nil, err
			}
			if build {
				arr = append(arr, child)
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return nil, errors.New("invalid array ending")
		}
		return arr, nil
	default:
		return nil, errors.New("unexpected delimiter")
	}
}

func validJSONUnicodeEscapes(raw []byte) bool {
	inString := false
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '"':
			inString = !inString
		case '\\':
			if !inString {
				continue
			}
			i++
			if i >= len(raw) {
				return false
			}
			if raw[i] != 'u' {
				continue
			}
			if i+4 >= len(raw) {
				return false
			}
			first, ok := parseHex4(raw[i+1 : i+5])
			if !ok {
				return false
			}
			switch {
			case first >= 0xD800 && first <= 0xDBFF:
				if i+10 >= len(raw) || raw[i+5] != '\\' || raw[i+6] != 'u' {
					return false
				}
				second, ok := parseHex4(raw[i+7 : i+11])
				if !ok || second < 0xDC00 || second > 0xDFFF {
					return false
				}
				i += 10
			case first >= 0xDC00 && first <= 0xDFFF:
				return false
			default:
				i += 4
			}
		}
	}
	return !inString
}

func parseHex4(raw []byte) (uint16, bool) {
	if len(raw) != 4 {
		return 0, false
	}
	var value uint16
	for _, b := range raw {
		value <<= 4
		switch {
		case b >= '0' && b <= '9':
			value |= uint16(b - '0')
		case b >= 'a' && b <= 'f':
			value |= uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			value |= uint16(b-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
