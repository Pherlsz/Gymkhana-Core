package portablejson

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	SchemaV1             = "portable_json_schema/v1"
	MaxSchemaDepth       = 16
	MaxSchemaProperties  = 256
	MaxSchemaArrayItems  = 10000
	MaxNumberLexemeBytes = 128
	MaxDecimalExponent   = 1024
	maxSchemaEnumValues  = 256
	maxSchemaDescription = 4096
	maxSchemaTitle       = 256
)

var maxSafeIntegerRat = new(big.Rat).SetInt64(MaxSafeInteger)

// CompiledSchema is an immutable, concurrency-safe pre-compiled representation
// of a SchemaV1 document.
type CompiledSchema struct {
	root *compiledNode
}

type compiledNode struct {
	unconstrainedObject bool
	types               []string
	primary             string
	properties          map[string]*compiledNode
	required            map[string]struct{}
	items               *compiledNode
	minItems            int64
	minItemsSet         bool
	maxItems            int64
	maxItemsSet         bool
	minimum             *big.Rat
	minimumSet          bool
	maximum             *big.Rat
	maximumSet          bool
	hasEnum             bool
	enumValues          []compiledEnumValue
}

type compiledEnumValue struct {
	isNull bool
	raw    any
	rat    *big.Rat
}

// CompileSchema validates and compiles raw into an immutable CompiledSchema.
func CompileSchema(raw json.RawMessage) (*CompiledSchema, error) {
	rootMap, err := decodeSchema(raw)
	if err != nil {
		return nil, err
	}
	compiled, err := compileSchemaNode(rootMap, 0)
	if err != nil {
		return nil, err
	}
	return &CompiledSchema{root: compiled}, nil
}

// CompileObjectSchema validates and compiles raw as an object-boundary schema.
// The `{}` compatibility form is accepted as an unconstrained object; otherwise
// the root must be non-nullable `object`.
func CompileObjectSchema(raw json.RawMessage) (*CompiledSchema, error) {
	compiled, err := CompileSchema(raw)
	if err != nil {
		return nil, err
	}
	if compiled.root.unconstrainedObject {
		return compiled, nil
	}
	if len(compiled.root.types) != 1 || compiled.root.types[0] != "object" {
		return nil, validationError(CodeInvalidSchema, "schema.type")
	}
	return compiled, nil
}

// ValidateSchema validates the conservative cross-runtime schema profile used
// by Core. It intentionally supports less than full JSON Schema so independent
// language implementations can preserve the same semantics.
func ValidateSchema(raw json.RawMessage) error {
	_, err := CompileSchema(raw)
	return err
}

// ValidateObjectSchema validates a schema intended for a JSON object boundary,
// such as tool-call arguments. The `{}` compatibility form is accepted as an
// unconstrained object; otherwise the root must be non-nullable `object`.
func ValidateObjectSchema(raw json.RawMessage) error {
	_, err := CompileObjectSchema(raw)
	return err
}

func decodeSchema(raw json.RawMessage) (map[string]any, error) {
	decoded, err := decodeStrictValue(raw, true)
	if err != nil {
		return nil, validationError(CodeInvalidSchema, "schema")
	}
	root, ok := decoded.(map[string]any)
	if !ok || root == nil {
		return nil, validationError(CodeInvalidSchema, "schema")
	}
	return root, nil
}

func compileSchemaNode(node map[string]any, depth int) (*compiledNode, error) {
	if depth > MaxSchemaDepth {
		return nil, validationError(CodeInvalidSchema, "schema.depth")
	}
	// `{}` is the explicit unconstrained-object compatibility form.
	if len(node) == 0 {
		return &compiledNode{unconstrainedObject: true}, nil
	}

	for key := range node {
		if !isAllowedSchemaKeyword(key) {
			return nil, validationError(CodeInvalidSchema, "schema.keyword")
		}
	}

	types, err := schemaTypes(node["type"])
	if err != nil || len(types) == 0 {
		return nil, validationError(CodeInvalidSchema, "schema.type")
	}
	primary := types[0]

	if description, ok := node["description"]; ok {
		value, ok := description.(string)
		if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxSchemaDescription {
			return nil, validationError(CodeInvalidSchema, "schema.description")
		}
	}
	if title, ok := node["title"]; ok {
		value, ok := title.(string)
		if !ok || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxSchemaTitle {
			return nil, validationError(CodeInvalidSchema, "schema.title")
		}
	}

	if err := validateKeywordApplicability(node, primary); err != nil {
		return nil, err
	}

	compiled := &compiledNode{
		types:   types,
		primary: primary,
	}

	switch primary {
	case "object":
		props, req, err := compileObjectSchemaNode(node, depth)
		if err != nil {
			return nil, err
		}
		compiled.properties = props
		compiled.required = req
	case "array":
		items, ok := node["items"].(map[string]any)
		if !ok {
			return nil, validationError(CodeInvalidSchema, "schema.items")
		}
		compiledItems, err := compileSchemaNode(items, depth+1)
		if err != nil {
			return nil, err
		}
		minVal, minSet, maxVal, maxSet, err := compileSchemaIntegerPair(node, "minItems", "maxItems")
		if err != nil {
			return nil, err
		}
		compiled.items = compiledItems
		compiled.minItems = minVal
		compiled.minItemsSet = minSet
		compiled.maxItems = maxVal
		compiled.maxItemsSet = maxSet
	case "integer", "number":
		minRat, minSet, maxRat, maxSet, err := compileSchemaNumberPair(node)
		if err != nil {
			return nil, err
		}
		compiled.minimum = minRat
		compiled.minimumSet = minSet
		compiled.maximum = maxRat
		compiled.maximumSet = maxSet
	case "string", "boolean", "null":
	default:
		return nil, validationError(CodeInvalidSchema, "schema.type")
	}

	if enumRaw, ok := node["enum"]; ok {
		values, ok := enumRaw.([]any)
		if !ok || len(values) == 0 || len(values) > maxSchemaEnumValues || primary == "object" || primary == "array" {
			return nil, validationError(CodeInvalidSchema, "schema.enum")
		}
		compiledEnums := make([]compiledEnumValue, len(values))
		for i, value := range values {
			ev, ok := compileEnumValue(value, types)
			if !ok {
				return nil, validationError(CodeInvalidSchema, "schema.enum")
			}
			compiledEnums[i] = ev
		}
		compiled.hasEnum = true
		compiled.enumValues = compiledEnums
	}
	return compiled, nil
}

func isAllowedSchemaKeyword(key string) bool {
	switch key {
	case "type", "properties", "required", "additionalProperties",
		"items", "enum", "description", "title",
		"minimum", "maximum", "minItems", "maxItems":
		return true
	default:
		return false
	}
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

func compileObjectSchemaNode(node map[string]any, depth int) (map[string]*compiledNode, map[string]struct{}, error) {
	propertiesRaw, hasProperties := node["properties"]
	if !hasProperties {
		propertiesRaw = map[string]any{}
	}
	properties, ok := propertiesRaw.(map[string]any)
	if !ok || len(properties) > MaxSchemaProperties {
		return nil, nil, validationError(CodeInvalidSchema, "schema.properties")
	}
	compiledProps := make(map[string]*compiledNode, len(properties))
	for name, rawChild := range properties {
		if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 256 {
			return nil, nil, validationError(CodeInvalidSchema, "schema.properties")
		}
		child, ok := rawChild.(map[string]any)
		if !ok {
			return nil, nil, validationError(CodeInvalidSchema, "schema.properties")
		}
		compiledChild, err := compileSchemaNode(child, depth+1)
		if err != nil {
			return nil, nil, err
		}
		compiledProps[name] = compiledChild
	}

	additional, ok := node["additionalProperties"].(bool)
	if !ok || additional {
		return nil, nil, validationError(CodeInvalidSchema, "schema.additionalProperties")
	}
	requiredRaw, ok := node["required"].([]any)
	if !ok || len(requiredRaw) != len(properties) {
		return nil, nil, validationError(CodeInvalidSchema, "schema.required")
	}
	seen := make(map[string]struct{}, len(requiredRaw))
	for _, item := range requiredRaw {
		name, ok := item.(string)
		if !ok {
			return nil, nil, validationError(CodeInvalidSchema, "schema.required")
		}
		if _, exists := properties[name]; !exists {
			return nil, nil, validationError(CodeInvalidSchema, "schema.required")
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, nil, validationError(CodeInvalidSchema, "schema.required")
		}
		seen[name] = struct{}{}
	}
	return compiledProps, seen, nil
}

func compileSchemaIntegerPair(node map[string]any, minKey, maxKey string) (int64, bool, int64, bool, error) {
	minValue, minSet, err := schemaNonNegativeInteger(node[minKey])
	if err != nil {
		return 0, false, 0, false, validationError(CodeInvalidSchema, "schema."+minKey)
	}
	maxValue, maxSet, err := schemaNonNegativeInteger(node[maxKey])
	if err != nil {
		return 0, false, 0, false, validationError(CodeInvalidSchema, "schema."+maxKey)
	}
	if (minSet && minValue > MaxSchemaArrayItems) || (maxSet && maxValue > MaxSchemaArrayItems) || (minSet && maxSet && minValue > maxValue) {
		return 0, false, 0, false, validationError(CodeInvalidSchema, "schema."+maxKey)
	}
	return minValue, minSet, maxValue, maxSet, nil
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

func compileSchemaNumberPair(node map[string]any) (*big.Rat, bool, *big.Rat, bool, error) {
	minimum, minSet, err := schemaNumber(node["minimum"])
	if err != nil {
		return nil, false, nil, false, validationError(CodeInvalidSchema, "schema.minimum")
	}
	maximum, maxSet, err := schemaNumber(node["maximum"])
	if err != nil {
		return nil, false, nil, false, validationError(CodeInvalidSchema, "schema.maximum")
	}
	if minSet && maxSet && minimum.Cmp(maximum) > 0 {
		return nil, false, nil, false, validationError(CodeInvalidSchema, "schema.maximum")
	}
	return minimum, minSet, maximum, maxSet, nil
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
	if len(text) == 0 || len(text) > MaxNumberLexemeBytes {
		return nil, errors.New("numeric lexeme too large")
	}
	if exponentAt := strings.IndexAny(text, "eE"); exponentAt >= 0 {
		exponent, err := strconv.ParseInt(text[exponentAt+1:], 10, 32)
		if err != nil || exponent < -MaxDecimalExponent || exponent > MaxDecimalExponent {
			return nil, errors.New("numeric exponent out of range")
		}
	}
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		return nil, errors.New("invalid JSON number")
	}
	if new(big.Rat).Abs(value).Cmp(maxSafeIntegerRat) > 0 {
		return nil, errors.New("non-portable number")
	}
	return value, nil
}

func compileEnumValue(value any, types []string) (compiledEnumValue, bool) {
	for _, schemaType := range types {
		switch schemaType {
		case "null":
			if value == nil {
				return compiledEnumValue{isNull: true}, true
			}
		case "string":
			if text, ok := value.(string); ok && utf8.ValidString(text) && utf8.RuneCountInString(text) <= 4096 {
				return compiledEnumValue{raw: text}, true
			}
		case "boolean":
			if b, ok := value.(bool); ok {
				return compiledEnumValue{raw: b}, true
			}
		case "integer":
			if number, ok := value.(json.Number); ok {
				parsed, err := strconv.ParseInt(number.String(), 10, 64)
				if err == nil && parsed >= -MaxSafeInteger && parsed <= MaxSafeInteger {
					rat := new(big.Rat).SetInt64(parsed)
					return compiledEnumValue{raw: number, rat: rat}, true
				}
			}
		case "number":
			if number, ok := value.(json.Number); ok {
				if rat, err := parsePortableNumber(number.String()); err == nil {
					return compiledEnumValue{raw: number, rat: rat}, true
				}
			}
		}
	}
	return compiledEnumValue{}, false
}
