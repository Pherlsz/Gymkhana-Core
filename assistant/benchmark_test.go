package assistant_test

import (
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
