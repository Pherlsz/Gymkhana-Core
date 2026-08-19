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
		var raw json.RawMessage
		if err := json.Unmarshal(testCase.Input, &raw); err != nil {
			return nil, "invalid_conformance_input"
		}
		return portableJSONValidationResult(portablejson.ValidateObject(raw))
	case "portablejson.value.validate":
		var raw json.RawMessage
		if err := json.Unmarshal(testCase.Input, &raw); err != nil {
			return nil, "invalid_conformance_input"
		}
		return portableJSONValidationResult(portablejson.ValidateValue(raw))
	case "portablejson.schema.validate":
		var raw json.RawMessage
		if err := json.Unmarshal(testCase.Input, &raw); err != nil {
			return nil, "invalid_conformance_input"
		}
		return portableJSONValidationResult(portablejson.ValidateSchema(raw))
	case "portablejson.object_schema.validate":
		var raw json.RawMessage
		if err := json.Unmarshal(testCase.Input, &raw); err != nil {
			return nil, "invalid_conformance_input"
		}
		return portableJSONValidationResult(portablejson.ValidateObjectSchema(raw))
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
