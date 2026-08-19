package assistant

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
)

// ValidatePortableJSONObject validates an untrusted JSON object without
// normalizing it. It rejects invalid UTF-8, duplicate object keys, excessive
// depth/node counts, trailing values, and payloads above the portable bound.
func ValidatePortableJSONObject(raw json.RawMessage) error {
	if len(raw) == 0 || len(raw) > maxPortableJSONBytes || !utf8.Valid(raw) {
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
	*nodes++
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
	if err != nil {
		return validationError(CodeInvalidSchema, "schema.type")
	}
	if len(types) == 0 {
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
		if err := validateSchemaNumber(node, "minimum"); err != nil {
			return err
		}
		if err := validateSchemaNumber(node, "maximum"); err != nil {
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
	if len(properties) == 0 {
		if required, exists := node["required"]; exists {
			values, ok := required.([]any)
			if !ok || len(values) != 0 {
				return validationError(CodeInvalidSchema, "schema.required")
			}
		}
		return nil
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
	if minSet && maxSet && minValue > maxValue {
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

func validateSchemaNumber(node map[string]any, key string) error {
	raw, exists := node[key]
	if !exists {
		return nil
	}
	if _, ok := raw.(json.Number); !ok {
		return validationError(CodeInvalidSchema, "schema."+key)
	}
	return nil
}

func schemaEnumValueMatches(value any, types []string) bool {
	for _, schemaType := range types {
		switch schemaType {
		case "null":
			if value == nil {
				return true
			}
		case "string":
			if _, ok := value.(string); ok {
				return true
			}
		case "boolean":
			if _, ok := value.(bool); ok {
				return true
			}
		case "integer":
			if number, ok := value.(json.Number); ok {
				_, err := strconv.ParseInt(number.String(), 10, 64)
				if err == nil {
					return true
				}
			}
		case "number":
			if _, ok := value.(json.Number); ok {
				return true
			}
		}
	}
	return false
}
