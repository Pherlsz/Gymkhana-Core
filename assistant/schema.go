package assistant

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"unicode/utf8"
)

const (
	PortableJSONSchemaV1 = "portable_json_schema/v1"
	maxPortableJSONBytes = 1 << 20
	maxJSONDepth         = 64
	maxJSONNodes         = 100000
	maxSchemaDepth       = 16
	maxSchemaProperties  = 256
	maxSchemaArrayItems  = 10000
)

// ValidatePortableJSONObject validates an untrusted JSON object without
// normalizing it. It rejects invalid UTF-8, duplicate object keys, ambiguous
// unpaired UTF-16 surrogate escapes, excessive depth/node counts, trailing
// values, and payloads above the portable bound.
func ValidatePortableJSONObject(raw json.RawMessage) error {
	if len(raw) == 0 || len(raw) > maxPortableJSONBytes || !utf8.Valid(raw) || !validJSONUnicodeEscapes(raw) {
		return validationError(CodeInvalidJSON, "json")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	nodes := 0
	if err := scanJSONValue(decoder, 0, &nodes, true); err != nil {
		return validationError(CodeInvalidJSON, "json")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return validationError(CodeInvalidJSON, "json")
	}
	return nil
}

// ValidatePortableJSONSchema validates the conservative cross-provider schema
// profile used by Core tools and structured output in Spec 0.3. It intentionally
// supports fewer keywords than full JSON Schema so adapters can translate the
// same semantics across providers that each implement different subsets.
func ValidatePortableJSONSchema(raw json.RawMessage) error {
	if err := ValidatePortableJSONObject(raw); err != nil {
		return validationError(CodeInvalidSchema, "schema")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil || root == nil {
		return validationError(CodeInvalidSchema, "schema")
	}
	if err := validateSchemaNode(root, 0); err != nil {
		return err
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder, depth int, nodes *int, requireObject bool) error {
	if depth > maxJSONDepth || *nodes >= maxJSONNodes {
		return errors.New("json bound exceeded")
	}
	*nodes = *nodes + 1
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, isDelim := token.(json.Delim)
	if requireObject && (!isDelim || delim != '{') {
		return errors.New("object required")
	}
	if !isDelim {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || !utf8.ValidString(key) {
				return errors.New("invalid object key")
			}
			if _, duplicate := seen[key]; duplicate {
				return errors.New("duplicate object key")
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder, depth+1, nodes, false); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("invalid object ending")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder, depth+1, nodes, false); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("invalid array ending")
		}
	default:
		return errors.New("unexpected delimiter")
	}
	return nil
}

func validateSchemaNode(node map[string]any, depth int) error {
	if depth > maxSchemaDepth {
		return validationError(CodeInvalidSchema, "schema.depth")
	}
	// `{}` is the explicit escape hatch for an unconstrained object. Any schema
	// that declares a type must otherwise follow the portable strict profile.
	if len(node) == 0 {
		return nil
	}
	allowed := map[string]struct{}{
		"type": {}, "properties": {}, "required": {}, "additionalProperties": {},
		"items": {}, "enum": {}, "description": {}, "title": {},
		"minimum": {}, "maximum": {}, "minItems": {}, "maxItems": {},
	}
	for key := range node {
		if _, ok := allowed[key]; !ok {
			return validationError(CodeInvalidSchema, "schema.keyword")
		}
	}

	types, err := portableSchemaTypes(node["type"])
	if err != nil || len(types) == 0 {
		return validationError(CodeInvalidSchema, "schema.type")
	}
	primary := types[0]

	if description, ok := node["description"]; ok {
		value, ok := description.(string)
		if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 4096 {
			return validationError(CodeInvalidSchema, "schema.description")
		}
	}
	if title, ok := node["title"]; ok {
		value, ok := title.(string)
		if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 256 {
			return validationError(CodeInvalidSchema, "schema.title")
		}
	}

	switch primary {
	case "object":
		if err := validateObjectSchema(node, depth); err != nil {
			return err
		}
	case "array":
		items, ok := node["items"].(map[string]any)
		if !ok {
			return validationError(CodeInvalidSchema, "schema.items")
		}
		if err := validateSchemaNode(items, depth+1); err != nil {
			return err
		}
		if err := validateSchemaIntegerPair(node, "minItems", "maxItems"); err != nil {
			return err
		}
	case "integer", "number":
		if err := validateSchemaNumberPair(node); err != nil {
			return err
		}
	case "string", "boolean", "null":
	default:
		return validationError(CodeInvalidSchema, "schema.type")
	}

	if enumValue, ok := node["enum"]; ok {
		values, ok := enumValue.([]any)
		if !ok || len(values) == 0 || len(values) > 256 {
			return validationError(CodeInvalidSchema, "schema.enum")
		}
		for _, value := range values {
			if !schemaEnumValueMatches(value, types) {
				return validationError(CodeInvalidSchema, "schema.enum")
			}
		}
	}
	return nil
}

func portableSchemaTypes(raw any) ([]string, error) {
	valid := func(value string) bool {
		switch value {
		case "object", "array", "string", "integer", "number", "boolean", "null":
			return true
		default:
			return false
		}
	}
	switch value := raw.(type) {
	case string:
		if !valid(value) {
			return nil, errors.New("invalid schema type")
		}
		return []string{value}, nil
	case []any:
		if len(value) != 2 {
			return nil, errors.New("only nullable union supported")
		}
		result := make([]string, 2)
		nullCount := 0
		for i, item := range value {
			text, ok := item.(string)
			if !ok || !valid(text) {
				return nil, errors.New("invalid schema type")
			}
			result[i] = text
			if text == "null" {
				nullCount++
			}
		}
		if nullCount != 1 || result[0] == result[1] {
			return nil, errors.New("invalid nullable union")
		}
		if result[0] == "null" {
			result[0], result[1] = result[1], result[0]
		}
		return result, nil
	default:
		return nil, errors.New("missing schema type")
	}
}

func validateObjectSchema(node map[string]any, depth int) error {
	propertiesRaw, hasProperties := node["properties"]
	if !hasProperties {
		propertiesRaw = map[string]any{}
	}
	properties, ok := propertiesRaw.(map[string]any)
	if !ok || len(properties) > maxSchemaProperties {
		return validationError(CodeInvalidSchema, "schema.properties")
	}
	for name, rawChild := range properties {
		if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 256 {
			return validationError(CodeInvalidSchema, "schema.properties")
		}
		child, ok := rawChild.(map[string]any)
		if !ok {
			return validationError(CodeInvalidSchema, "schema.properties")
		}
		if err := validateSchemaNode(child, depth+1); err != nil {
			return err
		}
	}
	additional, ok := node["additionalProperties"].(bool)
	if !ok || additional {
		return validationError(CodeInvalidSchema, "schema.additionalProperties")
	}
	requiredRaw, ok := node["required"].([]any)
	if !ok || len(requiredRaw) != len(properties) {
		return validationError(CodeInvalidSchema, "schema.required")
	}
	seen := make(map[string]struct{}, len(requiredRaw))
	for _, item := range requiredRaw {
		name, ok := item.(string)
		if !ok {
			return validationError(CodeInvalidSchema, "schema.required")
		}
		if _, exists := properties[name]; !exists {
			return validationError(CodeInvalidSchema, "schema.required")
		}
		if _, duplicate := seen[name]; duplicate {
			return validationError(CodeInvalidSchema, "schema.required")
		}
		seen[name] = struct{}{}
	}
	return nil
}

func validateSchemaIntegerPair(node map[string]any, minKey, maxKey string) error {
	minValue, minSet, err := schemaNonNegativeInteger(node[minKey])
	if err != nil {
		return validationError(CodeInvalidSchema, "schema."+minKey)
	}
	maxValue, maxSet, err := schemaNonNegativeInteger(node[maxKey])
	if err != nil {
		return validationError(CodeInvalidSchema, "schema."+maxKey)
	}
	if (minSet && minValue > maxSchemaArrayItems) || (maxSet && maxValue > maxSchemaArrayItems) || (minSet && maxSet && minValue > maxValue) {
		return validationError(CodeInvalidSchema, "schema."+maxKey)
	}
	return nil
}

func schemaNonNegativeInteger(raw any) (int64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	number, ok := raw.(json.Number)
	if !ok {
		return 0, false, errors.New("integer required")
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil || value < 0 || value > maxPortableJSONInteger {
		return 0, false, errors.New("invalid integer")
	}
	return value, true, nil
}

func validateSchemaNumberPair(node map[string]any) error {
	minimum, minSet, err := portableSchemaNumber(node["minimum"])
	if err != nil {
		return validationError(CodeInvalidSchema, "schema.minimum")
	}
	maximum, maxSet, err := portableSchemaNumber(node["maximum"])
	if err != nil {
		return validationError(CodeInvalidSchema, "schema.maximum")
	}
	if minSet && maxSet && minimum > maximum {
		return validationError(CodeInvalidSchema, "schema.maximum")
	}
	return nil
}

func portableSchemaNumber(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	number, ok := raw.(json.Number)
	if !ok {
		return 0, false, errors.New("number required")
	}
	value, err := strconv.ParseFloat(number.String(), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > float64(maxPortableJSONInteger) {
		return 0, false, errors.New("non-portable number")
	}
	return value, true, nil
}

func schemaEnumValueMatches(value any, types []string) bool {
	for _, schemaType := range types {
		switch schemaType {
		case "null":
			if value == nil {
				return true
			}
		case "string":
			if text, ok := value.(string); ok && utf8.ValidString(text) && utf8.RuneCountInString(text) <= 4096 {
				return true
			}
		case "boolean":
			if _, ok := value.(bool); ok {
				return true
			}
		case "integer":
			if number, ok := value.(json.Number); ok {
				parsed, err := strconv.ParseInt(number.String(), 10, 64)
				if err == nil && parsed >= -maxPortableJSONInteger && parsed <= maxPortableJSONInteger {
					return true
				}
			}
		case "number":
			if number, ok := value.(json.Number); ok {
				parsed, err := strconv.ParseFloat(number.String(), 64)
				if err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) && math.Abs(parsed) <= float64(maxPortableJSONInteger) {
					return true
				}
			}
		}
	}
	return false
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
			if first >= 0xD800 && first <= 0xDBFF {
				if i+10 >= len(raw) || raw[i+5] != '\\' || raw[i+6] != 'u' {
					return false
				}
				second, ok := parseHex4(raw[i+7 : i+11])
				if !ok || second < 0xDC00 || second > 0xDFFF {
					return false
				}
				i += 10
				continue
			}
			if first >= 0xDC00 && first <= 0xDFFF {
				return false
			}
			i += 4
		}
	}
	return true
}

func parseHex4(value []byte) (uint16, bool) {
	if len(value) != 4 {
		return 0, false
	}
	var result uint16
	for _, b := range value {
		result <<= 4
		switch {
		case b >= '0' && b <= '9':
			result |= uint16(b - '0')
		case b >= 'a' && b <= 'f':
			result |= uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			result |= uint16(b-'A') + 10
		default:
			return 0, false
		}
	}
	return result, true
}
