package assistant_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func BenchmarkResolveModelCandidates10000(b *testing.B) {
	catalog := make([]assistant.ModelDescriptor, 10000)
	for i := range catalog {
		access := assistant.AccessPaid
		if i%3 == 0 {
			access = assistant.AccessFree
		}
		catalog[i] = assistant.ModelDescriptor{
			Ref:          assistant.ModelRef{Provider: "gateway", Model: assistant.ModelID(fmt.Sprintf("vendor/model-%05d", i))},
			Access:       access,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityToolCalling},
		}
	}
	policy := assistant.ModelPolicy{
		Mode:                 assistant.ModelDynamic,
		AllowedAccess:        []assistant.AccessTier{assistant.AccessFree, assistant.AccessPaid},
		PreferredAccess:      []assistant.AccessTier{assistant.AccessFree, assistant.AccessPaid},
		RequiredCapabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityToolCalling},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		resolved, err := assistant.ResolveModelCandidates(policy, catalog)
		if err != nil {
			b.Fatal(err)
		}
		if len(resolved) != len(catalog) {
			b.Fatalf("resolved %d models, want %d", len(resolved), len(catalog))
		}
	}
}

func BenchmarkValidatePortableJSONObject64K(b *testing.B) {
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
	if len(raw) > 1<<20 {
		b.Fatalf("benchmark fixture exceeds portable bound: %d", len(raw))
	}

	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := assistant.ValidatePortableJSONObject(raw); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidatePortableJSONSchema(b *testing.B) {
	raw := json.RawMessage(`{
		"type":"object",
		"properties":{
			"name":{"type":"string"},
			"age":{"type":["integer","null"],"minimum":0,"maximum":200},
			"tags":{"type":"array","items":{"type":"string"},"minItems":0,"maxItems":64}
		},
		"required":["name","age","tags"],
		"additionalProperties":false
	}`)

	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := assistant.ValidatePortableJSONSchema(raw); err != nil {
			b.Fatal(err)
		}
	}
}
