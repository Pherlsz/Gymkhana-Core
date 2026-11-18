package portablejson_test

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
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
		"duplicate_key":  json.RawMessage("{\"a\":1,\"a\":2}"),
		"array_root":     json.RawMessage("[1,2,3]"),
		"trailing_value": json.RawMessage("{} {}"),
	} {
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

func TestCompiledSchemaConcurrencyAndNilSafety(t *testing.T) {
	t.Parallel()

	var nilErr *portablejson.ValidationError
	if got := nilErr.Error(); got != "portablejson: validation error" {
		t.Fatalf("nil ValidationError.Error() = %q, want %q", got, "portablejson: validation error")
	}
	if portablejson.IsCode(error(nilErr), "") {
		t.Fatal("IsCode matched typed-nil *ValidationError")
	}
	if got := (&portablejson.ValidationError{Code: portablejson.CodeInvalidJSON}).Error(); got != "invalid_json" {
		t.Fatalf("ValidationError without field = %q", got)
	}
	if got := (&portablejson.ValidationError{Code: portablejson.CodeInvalidSchema, Field: "schema.type"}).Error(); got != "schema.type: invalid_schema" {
		t.Fatalf("ValidationError with field = %q", got)
	}

	var nilSchema *portablejson.CompiledSchema
	if err := nilSchema.ValidateInstance(json.RawMessage("{}")); !portablejson.IsCode(err, portablejson.CodeInvalidSchema) {
		t.Fatalf("nil CompiledSchema.ValidateInstance = %v, want invalid_schema", err)
	}

	schemaRaw := json.RawMessage(`{
		"type": "object",
		"properties": {
			"id": {"type": "integer", "minimum": 1, "maximum": 1000},
			"status": {"type": ["string", "null"], "enum": ["pending", "done", null]},
			"scores": {"type": "array", "items": {"type": "number", "minimum": 0.5, "maximum": 99.5}, "minItems": 1, "maxItems": 4}
		},
		"required": ["id", "status", "scores"],
		"additionalProperties": false
	}`)

	compiled, err := portablejson.CompileObjectSchema(schemaRaw)
	if err != nil {
		t.Fatalf("CompileObjectSchema() error = %v", err)
	}

	validInstance := json.RawMessage(`{"id": 42, "status": "done", "scores": [1.5, 50.0, 99.5]}`)
	invalidInstance := json.RawMessage(`{"id": 0, "status": "done", "scores": [1.5]}`)

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				if err := compiled.ValidateInstance(validInstance); err != nil {
					t.Errorf("concurrent valid instance failed: %v", err)
					return
				}
				if err := compiled.ValidateInstance(invalidInstance); !portablejson.IsCode(err, portablejson.CodeInvalidSchema) {
					t.Errorf("concurrent invalid instance got %v, want invalid_schema", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestUnicodeSurrogateEdgeCases(t *testing.T) {
	t.Parallel()

	// Valid surrogate pair \uD83D\uDE00 (😀)
	validPair := json.RawMessage(`{"emoji":"\uD83D\uDE00"}`)
	if err := portablejson.ValidateObject(validPair); err != nil {
		t.Fatalf("ValidateObject(valid surrogate pair) = %v", err)
	}

	for name, raw := range map[string]json.RawMessage{
		"high_only":          json.RawMessage(`{"bad":"\uD800"}`),
		"low_only":           json.RawMessage(`{"bad":"\uDC00"}`),
		"high_followed_text": json.RawMessage(`{"bad":"\uD800abc"}`),
		"high_followed_high": json.RawMessage(`{"bad":"\uD800\uD800"}`),
		"truncated_escape":   json.RawMessage(`{"bad":"\u12"}`),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := portablejson.ValidateObject(raw); !portablejson.IsCode(err, portablejson.CodeInvalidJSON) {
				t.Fatalf("ValidateObject(%s) = %v, want invalid_json", name, err)
			}
		})
	}
}

func TestUnconstrainedObjectSchemaValidatesNestedPortableNumbers(t *testing.T) {
	t.Parallel()

	unconstrained := json.RawMessage(`{}`)
	if err := portablejson.ValidateObjectSchema(unconstrained); err != nil {
		t.Fatalf("ValidateObjectSchema({}) = %v", err)
	}
	validNested := json.RawMessage(`{"items":[1, 2.5, {"count": 42}], "ok": true}`)
	if err := portablejson.ValidateInstance(unconstrained, validNested); err != nil {
		t.Fatalf("ValidateInstance({}, validNested) = %v", err)
	}
	if err := portablejson.ValidateInstance(unconstrained, json.RawMessage(`[1, 2]`)); !portablejson.IsCode(err, portablejson.CodeInvalidSchema) {
		t.Fatalf("ValidateInstance({}, array) = %v, want invalid_schema", err)
	}
	if err := portablejson.ValidateInstance(unconstrained, json.RawMessage(`{"items":[9007199254740992]}`)); !portablejson.IsCode(err, portablejson.CodeInvalidSchema) {
		t.Fatalf("ValidateInstance({}, unsafe integer) = %v, want invalid_schema", err)
	}
}
