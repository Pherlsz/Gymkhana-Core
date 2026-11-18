package portablejson

import (
	"encoding/json"
	"math/big"
	"strconv"
	"unicode/utf8"
)

// ValidateInstance validates an arbitrary JSON value against SchemaV1. It is a
// deliberately small portable validator, not a general-purpose JSON Schema engine.
func ValidateInstance(schemaRaw, instanceRaw json.RawMessage) error {
	compiled, err := CompileSchema(schemaRaw)
	if err != nil {
		return err
	}
	return compiled.ValidateInstance(instanceRaw)
}

// ValidateInstance validates instanceRaw against the pre-compiled schema.
// It is safe for concurrent use by multiple goroutines.
func (s *CompiledSchema) ValidateInstance(instanceRaw json.RawMessage) error {
	if s == nil || s.root == nil {
		return validationError(CodeInvalidSchema, "schema")
	}
	instance, err := decodeStrictValue(instanceRaw, false)
	if err != nil {
		return err
	}
	return validateCompiledNode(s.root, instance, 0)
}

func validateCompiledNode(schema *compiledNode, value any, depth int) error {
	if depth > MaxSchemaDepth {
		return validationError(CodeInvalidSchema, "instance.depth")
	}
	if schema.unconstrainedObject {
		if _, ok := value.(map[string]any); !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		return validatePortableNumbers(value, depth)
	}

	if !instanceTypeMatches(value, schema.types) {
		return validationError(CodeInvalidSchema, "instance.type")
	}
	if schema.hasEnum {
		if !matchesCompiledEnum(value, schema.enumValues) {
			return validationError(CodeInvalidSchema, "instance.enum")
		}
	}
	if value == nil {
		return nil
	}

	switch schema.primary {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		for name, childSchema := range schema.properties {
			childValue, exists := object[name]
			if !exists {
				return validationError(CodeInvalidSchema, "instance.required")
			}
			if err := validateCompiledNode(childSchema, childValue, depth+1); err != nil {
				return err
			}
		}
		for name := range object {
			if _, declared := schema.properties[name]; !declared {
				return validationError(CodeInvalidSchema, "instance.additional_properties")
			}
		}
	case "array":
		array, ok := value.([]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		if (schema.minItemsSet && int64(len(array)) < schema.minItems) || (schema.maxItemsSet && int64(len(array)) > schema.maxItems) {
			return validationError(CodeInvalidSchema, "instance.items")
		}
		for _, item := range array {
			if err := validateCompiledNode(schema.items, item, depth+1); err != nil {
				return err
			}
		}
	case "integer":
		if err := validatePortableInteger(value); err != nil {
			return err
		}
		if err := validateCompiledNumericBounds(schema, value); err != nil {
			return err
		}
	case "number":
		if err := validatePortableNumber(value); err != nil {
			return err
		}
		if err := validateCompiledNumericBounds(schema, value); err != nil {
			return err
		}
	case "string":
		text, ok := value.(string)
		if !ok || !utf8.ValidString(text) || len(text) > MaxBytes {
			return validationError(CodeInvalidSchema, "instance.string")
		}
	case "boolean", "null":
	}
	return nil
}

func instanceTypeMatches(value any, types []string) bool {
	for _, schemaType := range types {
		switch schemaType {
		case "null":
			if value == nil {
				return true
			}
		case "object":
			if _, ok := value.(map[string]any); ok {
				return true
			}
		case "array":
			if _, ok := value.([]any); ok {
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
			if validatePortableInteger(value) == nil {
				return true
			}
		case "number":
			if validatePortableNumber(value) == nil {
				return true
			}
		}
	}
	return false
}

func validatePortableNumbers(value any, depth int) error {
	if depth > MaxDepth {
		return validationError(CodeInvalidJSON, "json.depth")
	}
	switch typed := value.(type) {
	case json.Number:
		return validatePortableNumber(typed)
	case []any:
		for _, item := range typed {
			if err := validatePortableNumbers(item, depth+1); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, item := range typed {
			if err := validatePortableNumbers(item, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func validatePortableInteger(value any) error {
	number, ok := value.(json.Number)
	if !ok {
		return validationError(CodeInvalidSchema, "instance.integer")
	}
	integer, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil || integer < -MaxSafeInteger || integer > MaxSafeInteger {
		return validationError(CodeInvalidSchema, "instance.integer")
	}
	return nil
}

func validatePortableNumber(value any) error {
	number, ok := value.(json.Number)
	if !ok {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	if _, err := parsePortableNumber(number.String()); err != nil {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	return nil
}

func validateCompiledNumericBounds(schema *compiledNode, value any) error {
	if !schema.minimumSet && !schema.maximumSet {
		return nil
	}
	number, ok := value.(json.Number)
	if !ok {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	actual, err := parsePortableNumber(number.String())
	if err != nil {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	if schema.minimumSet && actual.Cmp(schema.minimum) < 0 {
		return validationError(CodeInvalidSchema, "instance.minimum")
	}
	if schema.maximumSet && actual.Cmp(schema.maximum) > 0 {
		return validationError(CodeInvalidSchema, "instance.maximum")
	}
	return nil
}

func matchesCompiledEnum(value any, candidates []compiledEnumValue) bool {
	if value == nil {
		for _, candidate := range candidates {
			if candidate.isNull {
				return true
			}
		}
		return false
	}
	if number, isNumber := value.(json.Number); isNumber {
		var actualRat *big.Rat
		for _, candidate := range candidates {
			if candidate.rat == nil {
				continue
			}
			if actualRat == nil {
				parsed, err := parsePortableNumber(number.String())
				if err != nil {
					return false
				}
				actualRat = parsed
			}
			if actualRat.Cmp(candidate.rat) == 0 {
				return true
			}
		}
		return false
	}
	for _, candidate := range candidates {
		if !candidate.isNull && candidate.rat == nil && value == candidate.raw {
			return true
		}
	}
	return false
}
