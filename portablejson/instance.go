package portablejson

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// ValidateInstance validates an arbitrary JSON value against SchemaV1. It is a
// deliberately small portable validator, not a general-purpose JSON Schema engine.
func ValidateInstance(schemaRaw, instanceRaw json.RawMessage) error {
	root, err := decodeSchema(schemaRaw)
	if err != nil {
		return err
	}
	if err := validateSchemaNode(root, 0); err != nil {
		return err
	}
	if err := ValidateValue(instanceRaw); err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(instanceRaw))
	decoder.UseNumber()
	var instance any
	if err := decoder.Decode(&instance); err != nil {
		return validationError(CodeInvalidJSON, "instance")
	}
	return validateInstanceNode(root, instance, 0)
}

func validateInstanceNode(schema map[string]any, value any, depth int) error {
	if depth > MaxSchemaDepth {
		return validationError(CodeInvalidSchema, "instance.depth")
	}
	if len(schema) == 0 {
		if _, ok := value.(map[string]any); !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		return validatePortableNumbers(value, depth)
	}

	types, err := schemaTypes(schema["type"])
	if err != nil || !instanceTypeMatches(value, types) {
		return validationError(CodeInvalidSchema, "instance.type")
	}
	if enumRaw, ok := schema["enum"]; ok {
		values, ok := enumRaw.([]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.enum")
		}
		matched := false
		for _, candidate := range values {
			if scalarEqual(value, candidate) {
				matched = true
				break
			}
		}
		if !matched {
			return validationError(CodeInvalidSchema, "instance.enum")
		}
	}
	if value == nil {
		return nil
	}

	primary := types[0]
	switch primary {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		properties, _ := schema["properties"].(map[string]any)
		if properties == nil {
			properties = map[string]any{}
		}
		for name, childSchemaRaw := range properties {
			childValue, exists := object[name]
			if !exists {
				return validationError(CodeInvalidSchema, "instance.required")
			}
			childSchema, ok := childSchemaRaw.(map[string]any)
			if !ok {
				return validationError(CodeInvalidSchema, "instance.schema")
			}
			if err := validateInstanceNode(childSchema, childValue, depth+1); err != nil {
				return err
			}
		}
		for name := range object {
			if _, declared := properties[name]; !declared {
				return validationError(CodeInvalidSchema, "instance.additional_properties")
			}
		}
	case "array":
		array, ok := value.([]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		minItems, minSet, err := schemaNonNegativeInteger(schema["minItems"])
		if err != nil {
			return validationError(CodeInvalidSchema, "instance.minItems")
		}
		maxItems, maxSet, err := schemaNonNegativeInteger(schema["maxItems"])
		if err != nil {
			return validationError(CodeInvalidSchema, "instance.maxItems")
		}
		if (minSet && int64(len(array)) < minItems) || (maxSet && int64(len(array)) > maxItems) {
			return validationError(CodeInvalidSchema, "instance.items")
		}
		itemSchema, ok := schema["items"].(map[string]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.items")
		}
		for _, item := range array {
			if err := validateInstanceNode(itemSchema, item, depth+1); err != nil {
				return err
			}
		}
	case "integer":
		if err := validatePortableInteger(value); err != nil {
			return err
		}
		if err := validateNumericBounds(schema, value); err != nil {
			return err
		}
	case "number":
		if err := validatePortableNumber(value); err != nil {
			return err
		}
		if err := validateNumericBounds(schema, value); err != nil {
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

func validateNumericBounds(schema map[string]any, value any) error {
	number, ok := value.(json.Number)
	if !ok {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	actual, err := parsePortableNumber(number.String())
	if err != nil {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	if minimum, set, err := schemaNumber(schema["minimum"]); err != nil {
		return validationError(CodeInvalidSchema, "instance.minimum")
	} else if set && actual.Cmp(minimum) < 0 {
		return validationError(CodeInvalidSchema, "instance.minimum")
	}
	if maximum, set, err := schemaNumber(schema["maximum"]); err != nil {
		return validationError(CodeInvalidSchema, "instance.maximum")
	} else if set && actual.Cmp(maximum) > 0 {
		return validationError(CodeInvalidSchema, "instance.maximum")
	}
	return nil
}

func scalarEqual(left, right any) bool {
	leftNumber, leftIsNumber := left.(json.Number)
	rightNumber, rightIsNumber := right.(json.Number)
	if leftIsNumber || rightIsNumber {
		if !leftIsNumber || !rightIsNumber {
			return false
		}
		leftRat, leftErr := parsePortableNumber(leftNumber.String())
		rightRat, rightErr := parsePortableNumber(rightNumber.String())
		return leftErr == nil && rightErr == nil && leftRat.Cmp(rightRat) == 0
	}
	return reflect.DeepEqual(left, right)
}
