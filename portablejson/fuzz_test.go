package portablejson_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

func FuzzValue(f *testing.F) {
	for _, seed := range []string{
		`{}`,
		`{"a":1}`,
		`{"nested":{"x":[1,2,3]}}`,
		`{"a":1,"a":2}`,
		`[]`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 32768 {
			return
		}
		raw := json.RawMessage(value)
		if err := portablejson.ValidateValue(raw); err != nil {
			return
		}
		var decoded any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&decoded); err != nil {
			t.Fatalf("portable value failed standard JSON decode: %v", err)
		}
	})
}

func FuzzSchema(f *testing.F) {
	for _, seed := range []string{
		`{}`,
		`{"type":"string"}`,
		`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}`,
		`{"type":"string","pattern":".*"}`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 32768 {
			return
		}
		raw := json.RawMessage(value)
		if err := portablejson.ValidateSchema(raw); err != nil {
			return
		}
		if err := portablejson.ValidateObject(raw); err != nil {
			t.Fatalf("valid portable schema is not a JSON object: %v", err)
		}
	})
}
