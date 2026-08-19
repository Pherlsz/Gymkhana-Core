package portablejson_test

import (
	"encoding/json"
	"errors"
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
		"duplicate_key":      json.RawMessage("{\"a\":1,\"a\":2}"),
		"unpaired_surrogate": json.RawMessage([]byte{123, 34, 118, 97, 108, 117, 101, 34, 58, 34, 92, 117, 100, 56, 48, 48, 34, 125}),
		"array_root":         json.RawMessage("[1,2,3]"),
		"trailing_value":     json.RawMessage("{} {}"),
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
