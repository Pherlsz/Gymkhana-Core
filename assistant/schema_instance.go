package assistant

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// ValidatePortableJSONValue validates one arbitrary JSON value using the same
// duplicate-key, Unicode, size, depth, and node-count rules as portable objects.
// It intentionally preserves the lexical JSON boundary instead of relying on a
// language's default float/object coercion behavior.
func ValidatePortableJSONValue(raw json.RawMessage) error {
	if len(raw) == 0 || len(raw) > maxPortableJSONBytes || !utf8.Valid(raw) || !validJSONUnicodeEscapes(raw) {
		return validationError(CodeInvalidJSON, "json")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	nodes := 0
	if err := scanJSONValue(decoder, 0, &nodes, false); err != nil {
		return validationError(CodeInvalidJSON, "json")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return validationError(CodeInvalidJSON, "json")
	}
	return nil
}

// ValidatePortableJSONInstance validates an instance against
// portable_json_schema/v1. It is intentionally not a full JSON Schema engine.
func ValidatePortableJSONInstance(schemaRaw, instanceRaw json.RawMessage) error {
	if err := ValidatePortableJSONSchema(schemaRaw); err != nil {
		return err
	}
	if err := ValidatePortableJSONValue(instanceRaw); err != nil {
		return err
	}

	schemaDecoder := json.NewDecoder(bytes.NewReader(schemaRaw))
	schemaDecoder.UseNumber()
	var schema map[string]any
	if err := schemaDecoder.Decode(&schema); err != nil || schema == nil {
		return validationError(CodeInvalidSchema, "schema")
	}

	instanceDecoder := json.NewDecoder(bytes.NewReader(instanceRaw))
	instanceDecoder.UseNumber()
	var instance any
	if err := instanceDecoder.Decode(&instance); err != nil {
		return validationError(CodeInvalidJSON, "instance")
	}
	if err := validatePortableInstanceNode(schema, instance, 0); err != nil {
		return err
	}
	return nil
}

// ValidateToolArguments validates provider-generated tool arguments against the
// exact portable schema exposed in the request before a host considers executing
// the call. Name equality is required so a schema cannot be borrowed from a
// different allowed tool.
func ValidateToolArguments(definition ToolDefinition, call ToolCall) error {
	if err := ValidateToolDefinition(definition); err != nil {
		return err
	}
	if err := ValidateToolCall(call); err != nil {
		return err
	}
	if definition.Name != call.Name {
		return validationError(CodeInvalidToolCall, "tool_call.name")
	}
	if err := ValidatePortableJSONInstance(definition.InputSchema, call.Arguments); err != nil {
		return validationError(CodeInvalidToolCall, "tool_call.arguments")
	}
	return nil
}

func validatePortableInstanceNode(schema map[string]any, value any, depth int) error {
	if depth > maxSchemaDepth {
		return validationError(CodeInvalidSchema, "instance.depth")
	}
	if len(schema) == 0 {
		// `{}` is the unconstrained-object compatibility form.
		if _, ok := value.(map[string]any); !ok {
			return validationError(CodeInvalidSchema, "instance.type")
		}
		return validatePortableValueNumbers(value, depth)
	}
	types, err := portableSchemaTypes(schema["type"])
	if err != nil || !portableInstanceTypeMatches(value, types) {
		return validationError(CodeInvalidSchema, "instance.type")
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
			if err := validatePortableInstanceNode(childSchema, childValue, depth+1); err != nil {
				return err
			}
		}
		if additional, ok := schema["additionalProperties"].(bool); ok && !additional {
			for name := range object {
				if _, declared := properties[name]; !declared {
					return validationError(CodeInvalidSchema, "instance.additional_properties")
				}
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
		if minSet && int64(len(array)) < minItems || maxSet && int64(len(array)) > maxItems {
			return validationError(CodeInvalidSchema, "instance.items")
		}
		itemSchema, ok := schema["items"].(map[string]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.items")
		}
		for _, item := range array {
			if err := validatePortableInstanceNode(itemSchema, item, depth+1); err != nil {
				return err
			}
		}
	case "integer":
		if err := validatePortableInteger(value); err != nil {
			return err
		}
		if err := validateInstanceNumericBounds(schema, value); err != nil {
			return err
		}
	case "number":
		if err := validatePortableNumber(value); err != nil {
			return err
		}
		if err := validateInstanceNumericBounds(schema, value); err != nil {
			return err
		}
	case "string":
		text, ok := value.(string)
		if !ok || !utf8.ValidString(text) || utf8.RuneCountInString(text) > 1<<20 {
			return validationError(CodeInvalidSchema, "instance.string")
		}
	case "boolean", "null":
	}

	if enumRaw, ok := schema["enum"]; ok {
		values, ok := enumRaw.([]any)
		if !ok {
			return validationError(CodeInvalidSchema, "instance.enum")
		}
		matched := false
		for _, candidate := range values {
			if portableJSONScalarEqual(value, candidate) {
				matched = true
				break
			}
		}
		if !matched {
			return validationError(CodeInvalidSchema, "instance.enum")
		}
	}
	return nil
}

func portableInstanceTypeMatches(value any, types []string) bool {
	for _, schemaType := range types {
		switch schemaType {
		case "null":
			if value == nil {
				return true
			}
		case "object":
			_, ok := value.(map[string]any)
			if ok {
				return true
			}
		case "array":
			_, ok := value.([]any)
			if ok {
				return true
			}
		case "string":
			_, ok := value.(string)
			if ok {
				return true
			}
		case "boolean":
			_, ok := value.(bool)
			if ok {
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

func validatePortableValueNumbers(value any, depth int) error {
	if depth > maxJSONDepth {
		return validationError(CodeInvalidJSON, "json.depth")
	}
	switch typed := value.(type) {
	case json.Number:
		return validatePortableNumber(typed)
	case []any:
		for _, item := range typed {
			if err := validatePortableValueNumbers(item, depth+1); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, item := range typed {
			if err := validatePortableValueNumbers(item, depth+1); err != nil {
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
	if err != nil || integer < -maxPortableJSONInteger || integer > maxPortableJSONInteger {
		return validationError(CodeInvalidSchema, "instance.integer")
	}
	return nil
}

func validatePortableNumber(value any) error {
	number, ok := value.(json.Number)
	if !ok {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	parsed, err := strconv.ParseFloat(number.String(), 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || math.Abs(parsed) > float64(maxPortableJSONInteger) {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	return nil
}

func validateInstanceNumericBounds(schema map[string]any, value any) error {
	number, ok := value.(json.Number)
	if !ok {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	actual, err := strconv.ParseFloat(number.String(), 64)
	if err != nil || math.IsNaN(actual) || math.IsInf(actual, 0) {
		return validationError(CodeInvalidSchema, "instance.number")
	}
	if minimum, set, err := portableSchemaNumber(schema["minimum"]); err != nil {
		return validationError(CodeInvalidSchema, "instance.minimum")
	} else if set && actual < minimum {
		return validationError(CodeInvalidSchema, "instance.minimum")
	}
	if maximum, set, err := portableSchemaNumber(schema["maximum"]); err != nil {
		return validationError(CodeInvalidSchema, "instance.maximum")
	} else if set && actual > maximum {
		return validationError(CodeInvalidSchema, "instance.maximum")
	}
	return nil
}

func portableJSONScalarEqual(left, right any) bool {
	leftNumber, leftIsNumber := left.(json.Number)
	rightNumber, rightIsNumber := right.(json.Number)
	if leftIsNumber || rightIsNumber {
		if !leftIsNumber || !rightIsNumber {
			return false
		}
		leftFloat, leftErr := strconv.ParseFloat(leftNumber.String(), 64)
		rightFloat, rightErr := strconv.ParseFloat(rightNumber.String(), 64)
		return leftErr == nil && rightErr == nil && leftFloat == rightFloat
	}
	return reflect.DeepEqual(left, right)
}
