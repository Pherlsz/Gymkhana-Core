package portablejson_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

func TestValidateObject(t *testing.T) {
	t.Parallel()
	valid := json.RawMessage("{\"city\":\"Porto Alegre\",\"nested\":{\"value\":1}}")
	if err := portablejson.ValidateObject(valid); err != nil {
		t.Fatalf("ValidateObject(valid) = %v", err)
	}
	for name, raw := range map[string]json.RawMessage{
		"duplicate_key": json.RawMessage("{\"a\":1,\"a\":2}"),
		"array_root":    json.RawMessage("[1,2,3]"),
		"trailing_value": json.RawMessage("{} {}"),
	} {
		name, raw := name, raw
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := portablejson.ValidateObject(raw); err == nil {
				t.Fatalf("ValidateObject(%s) unexpectedly succeeded", raw)
			}
		})
	}
}

func TestValidateSchemaAndInstance(t *testing.T) {
	t.Parallel()
	schema := json.RawMessage("{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"age\":{\"type\":[\"integer\",\"null\"],\"minimum\":0,\"maximum\":200}},\"required\":[\"name\",\"age\"],\"additionalProperties\":false}")
	if err := portablejson.ValidateSchema(schema); err != nil {
		t.Fatalf("ValidateSchema(valid) = %v", err)
	}
	if err := portablejson.ValidateObjectSchema(schema); err != nil {
		t.Fatalf("ValidateObjectSchema(valid) = %v", err)
	}
	valid := json.RawMessage("{\"name\":\"Ana\",\"age\":32}")
	if err := portablejson.ValidateInstance(schema, valid); err != nil {
		t.Fatalf("ValidateInstance(valid) = %v", err)
	}
	for _, raw := range []json.RawMessage{
		json.RawMessage("{\"name\":\"Ana\"}"),
		json.RawMessage("{\"name\":\"Ana\",\"age\":201}"),
		json.RawMessage("{\"name\":\"Ana\",\"age\":32,\"extra\":true}"),
		json.RawMessage("{\"name\":\"Ana\",\"age\":9007199254740992}"),
	} {
		if err := portablejson.ValidateInstance(schema, raw); err == nil {
			t.Fatalf("ValidateInstance(%s) unexpectedly succeeded", raw)
		}
	}
}

func TestExactDecimalSemantics(t *testing.T) {
	t.Parallel()
	schema := json.RawMessage("{\"type\":\"number\",\"enum\":[0.1]}")
	if err := portablejson.ValidateInstance(schema, json.RawMessage("0.1")); err != nil {
		t.Fatalf("exact enum member = %v", err)
	}
	if err := portablejson.ValidateInstance(schema, json.RawMessage("0.10000000000000001")); err == nil {
		t.Fatal("distinct decimal unexpectedly matched enum")
	}
}

func TestExactDecimalParsingIsBounded(t *testing.T) {
	t.Parallel()
	schema := json.RawMessage("{\"type\":\"number\"}")
	for _, value := range []string{
		"1e-1025",
		"0." + strings.Repeat("0", portablejson.MaxNumberLexemeBytes) + "1",
	} {
		if err := portablejson.ValidateInstance(schema, json.RawMessage(value)); err == nil {
			t.Fatalf("unbounded numeric value %q unexpectedly accepted", value)
		}
	}
}

func TestNullableEnumSemantics(t *testing.T) {
	t.Parallel()
	rejectNull := json.RawMessage("{\"type\":[\"string\",\"null\"],\"enum\":[\"active\"]}")
	if err := portablejson.ValidateInstance(rejectNull, json.RawMessage("null")); err == nil {
		t.Fatal("null unexpectedly matched enum")
	}
	acceptNull := json.RawMessage("{\"type\":[\"string\",\"null\"],\"enum\":[\"active\",null]}")
	if err := portablejson.ValidateInstance(acceptNull, json.RawMessage("null")); err != nil {
		t.Fatalf("declared null enum member = %v", err)
	}
}

func TestValidateSchemaRejectsNonPortableSemantics(t *testing.T) {
	t.Parallel()
	cases := []json.RawMessage{
		json.RawMessage("{\"type\":\"string\",\"pattern\":\".*\"}"),
		json.RawMessage("{\"type\":\"string\",\"minimum\":0}"),
		json.RawMessage("{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"}},\"required\":[],\"additionalProperties\":false}"),
	}
	for _, raw := range cases {
		err := portablejson.ValidateSchema(raw)
		if err == nil {
			t.Fatalf("ValidateSchema(%s) unexpectedly succeeded", raw)
		}
		var validation *portablejson.ValidationError
		if !errors.As(err, &validation) || validation.Code != portablejson.CodeInvalidSchema {
			t.Fatalf("ValidateSchema(%s) = %v", raw, err)
		}
	}
}

func TestObjectSchemaRequiresObjectRoot(t *testing.T) {
	t.Parallel()
	if err := portablejson.ValidateSchema(json.RawMessage("{\"type\":\"string\"}")); err != nil {
		t.Fatalf("generic string schema = %v", err)
	}
	if err := portablejson.ValidateObjectSchema(json.RawMessage("{\"type\":\"string\"}")); err == nil {
		t.Fatal("object schema accepted string root")
	}
}
