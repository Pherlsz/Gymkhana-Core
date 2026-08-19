package portablejson_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

func BenchmarkValidateObject64K(b *testing.B) {
	payload := make(map[string]any, 512)
	for i := range 512 {
		payload[fmt.Sprintf("field_%03d", i)] = map[string]any{
			"id":      i,
			"enabled": i%2 == 0,
			"label":   fmt.Sprintf("value-%03d", i),
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		b.Fatal(err)
	}
	if len(raw) > portablejson.MaxBytes {
		b.Fatalf("benchmark fixture exceeds portable bound: %d", len(raw))
	}

	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := portablejson.ValidateObject(raw); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateSchema(b *testing.B) {
	raw := json.RawMessage("{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"age\":{\"type\":[\"integer\",\"null\"],\"minimum\":0,\"maximum\":200},\"tags\":{\"type\":\"array\",\"items\":{\"type\":\"string\"},\"minItems\":0,\"maxItems\":64}},\"required\":[\"name\",\"age\",\"tags\"],\"additionalProperties\":false}")

	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := portablejson.ValidateSchema(raw); err != nil {
			b.Fatal(err)
		}
	}
}
