package portablejson

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"unicode/utf8"
)

const (
	SchemaV1             = "portable_json_schema/v1"
	MaxSchemaDepth       = 16
	MaxSchemaProperties  = 256
	MaxSchemaArrayItems  = 10000
	maxSchemaEnumValues  = 256
	maxSchemaDescription = 4096
	maxSchemaTitle       = 256
)

// ValidateSchema validates the conservative cross-runtime schema profile used
// by Core. It intentionally supports less than full JSON Schema so independent
// language implementations can preserve the same semantics.
func ValidateSchema(raw json.RawMessage) error {
	root, err := decodeSchema(raw)
	if err != nil {
		return err
	}
	return validateSchemaNode(root, 0)
}

// ValidateObjectSchema validates a schema intended for a JSON object boundary,
// such as tool-call arguments. The `{}` compatibility form is accepted as an
// unconstrained object; otherwise the root must be non-nullable `object`.
func ValidateObjectSchema(raw json.RawMessage) error {
	root, err := decodeSchema(raw)
	if err != nil {
		return err
	}
	if err := validateSchemaNode(root, 0); err != nil {
		return err
	}
	if len(root) == 0 {
		return nil
	}
	types, err := schemaTypes(root["type"])
	if err != nil || len(types) != 1 || types[0] != "object" {
		return validationError(CodeInvalidSchema, "schema.type")
	}
	return nil
}

func decodeSchema(raw json.RawMessage) (map[string]any, error) {
	if err := ValidateObject(raw); err != nil {
		return nil, validationError(CodeInvalidSchema, "schema")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil || root == nil {
		return nil, validationError(CodeInvalidSchema, "schema")
	}
	return root, nil
}

func validateSchemaNode(node map[string]any, depth int) error {
	if depth > MaxSchemaDepth {
		return validationError(CodeInvalidSchema, "schema.depth")
	}
	// `{}` is the explicit unconstrained-object compatibility form.
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

	types, err := schemaTypes(node["type"])
	if err != nil || len(types) == 0 {
		return validationError(CodeInvalidSchema, "schema.type")
	}
	primary := types[0]

	if description, ok := node["description"]; ok {
		value, ok := description.(string)
		if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxSchemaDescription {
			return validationError(CodeInvalidSchema, "schema.description")
		}
	}
	if title, ok := node["title"]; ok {
		value, ok := title.(string)
		if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxSchemaTitle {
			return validationError(CodeInvalidSchema, "schema.title")
		}
	}

	if err := validateKeywordApplicability(node, primary); err != nil {
		return err
	}

	switch primary {
	case "object":
		if err := validateObjectSchemaNode(node, depth); err != nil {
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

	if enumRaw, ok := node["enum"]; ok {
		values, ok := enumRaw.([]any)
		if !ok || len(values) == 0 || len(values) > maxSchemaEnumValues || primary == "object" || primary == "array" {
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

func validateKeywordApplicability(node map[string]any, primary string) error {
	for key := range node {
		if key == "type" || key == "description" || key == "title" || key == "enum" {
			continue
		}
		valid := false
		switch primary {
		case "object":
			valid = key == "properties" || key == "required" || key == "additionalProperties"
		case "array":
			valid = key == "items" || key == "minItems" || key == "maxItems"
		case "integer", "number":
			valid = key == "minimum" || key == "maximum"
		case "string", "boolean", "null":
			valid = false
		}
		if !valid {
			return validationError(CodeInvalidSchema, "schema.keyword")
		}
	}
	return nil
}

func schemaTypes(raw any) ([]string, error) {
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

func validateObjectSchemaNode(node map[string]any, depth int) error {
	propertiesRaw, hasProperties := node["properties"]
	if !hasProperties {
		propertiesRaw = map[string]any{}
	}
	properties, ok := propertiesRaw.(map[string]any)
	if !ok || len(properties) > MaxSchemaProperties {
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
	if (minSet && minValue > MaxSchemaArrayItems) || (maxSet && maxValue > MaxSchemaArrayItems) || (minSet && maxSet && minValue > maxValue) {
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
	if err != nil || value < 0 || value > MaxSafeInteger {
		return 0, false, errors.New("invalid integer")
	}
	return value, true, nil
}

func validateSchemaNumberPair(node map[string]any) error {
	minimum, minSet, err := schemaNumber(node["minimum"])
	if err != nil {
		return validationError(CodeInvalidSchema, "schema.minimum")
	}
	maximum, maxSet, err := schemaNumber(node["maximum"])
	if err != nil {
		return validationError(CodeInvalidSchema, "schema.maximum")
	}
	if minSet && maxSet && minimum.Cmp(maximum) > 0 {
		return validationError(CodeInvalidSchema, "schema.maximum")
	}
	return nil
}

func schemaNumber(raw any) (*big.Rat, bool, error) {
	if raw == nil {
		return nil, false, nil
	}
	number, ok := raw.(json.Number)
	if !ok {
		return nil, false, errors.New("number required")
	}
	value, err := parsePortableNumber(number.String())
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func parsePortableNumber(text string) (*big.Rat, error) {
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		return nil, errors.New("invalid JSON number")
	}
	limit := new(big.Rat).SetInt64(MaxSafeInteger)
	if new(big.Rat).Abs(value).Cmp(limit) > 0 {
		return nil, errors.New("non-portable number")
	}
	return value, nil
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
				if err == nil && parsed >= -MaxSafeInteger && parsed <= MaxSafeInteger {
					return true
				}
			}
		case "number":
			if number, ok := value.(json.Number); ok {
				if _, err := parsePortableNumber(number.String()); err == nil {
					return true
				}
			}
		}
	}
	return false
}
