package conformance

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

func TestPortableJSONConformance(t *testing.T) {
	runSuite(t, "portablejson.json", executePortableJSON)
}

func executePortableJSON(testCase vector) (any, string) {
	switch testCase.Operation {
	case "portablejson.object.validate":
		return portableJSONValidationResult(portablejson.ValidateObject(testCase.Input))
	case "portablejson.value.validate":
		return portableJSONValidationResult(portablejson.ValidateValue(testCase.Input))
	case "portablejson.schema.validate":
		return portableJSONValidationResult(portablejson.ValidateSchema(testCase.Input))
	case "portablejson.object_schema.validate":
		return portableJSONValidationResult(portablejson.ValidateObjectSchema(testCase.Input))
	case "portablejson.instance.validate":
		var value portableJSONInstanceInput
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return portableJSONValidationResult(portablejson.ValidateInstance(value.Schema, value.Instance))
	default:
		return nil, "unsupported_operation"
	}
}

type portableJSONInstanceInput struct {
	Schema   json.RawMessage `json:"schema"`
	Instance json.RawMessage `json:"instance"`
}

func portableJSONValidationResult(err error) (any, string) {
	if err == nil {
		return true, ""
	}
	var validation *portablejson.ValidationError
	if errors.As(err, &validation) {
		return nil, string(validation.Code)
	}
	return nil, "unknown_error"
}
