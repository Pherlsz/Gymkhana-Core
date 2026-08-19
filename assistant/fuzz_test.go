package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func FuzzAssistantMessageJSON(f *testing.F) {
	for _, seed := range []string{
		"{\"role\":\"user\",\"content\":[{\"type\":\"text\",\"text\":\"hello\"}]}",
		"{\"role\":\"assistant\",\"content\":[{\"type\":\"tool_call\",\"tool_call\":{\"id\":\"call_01\",\"name\":\"lookup_weather\",\"arguments\":{\"location\":\"Tokyo\"}}}]}",
		"{\"role\":\"tool\",\"content\":[{\"type\":\"tool_result\",\"tool_result\":{\"call_id\":\"call_01\",\"content\":[{\"type\":\"text\",\"text\":\"21 C\"}]}}]}",
		"{\"role\":\"user\",\"content\":[{\"type\":\"audio\",\"media\":{\"uri\":\"https://example.test/audio.wav\"}}]}",
		"{}",
		"null",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 8192 {
			return
		}
		var message assistant.Message
		if err := json.Unmarshal([]byte(value), &message); err != nil {
			return
		}
		if err := assistant.ValidateMessage(message); err != nil {
			return
		}
		encoded, err := json.Marshal(message)
		if err != nil {
			t.Fatalf("valid message failed to marshal: %v", err)
		}
		var roundTrip assistant.Message
		if err := json.Unmarshal(encoded, &roundTrip); err != nil {
			t.Fatalf("marshaled valid message failed to decode: %v", err)
		}
		if err := assistant.ValidateMessage(roundTrip); err != nil {
			t.Fatalf("valid message became invalid after JSON round trip: %v", err)
		}
	})
}

func FuzzAssistantDefinitionJSON(f *testing.F) {
	for _, seed := range []string{
		"{\"id\":\"chat\",\"name\":\"Chat\",\"modules\":[\"text\"],\"model_policy\":{\"mode\":\"dynamic\"},\"credentials\":{\"allowed_modes\":[\"managed\"]}}",
		"{\"id\":\"research\",\"name\":\"Research\",\"modules\":[\"text\",\"retrieval\"],\"optional_modules\":[\"vision\"],\"model_policy\":{\"mode\":\"dynamic\",\"allowed_access\":[\"free\",\"paid\"]},\"credentials\":{\"allowed_modes\":[\"managed\",\"byok\"]},\"retrieval\":{\"mode\":\"on_demand\",\"search\":\"hybrid\",\"query_transform\":\"rewrite\",\"candidate_limit\":20,\"context_limit\":6,\"rerank\":true,\"grounding\":\"required\",\"require_citations\":true}}",
		"{}",
		"null",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 16384 {
			return
		}
		var definition assistant.AssistantDefinition
		if err := json.Unmarshal([]byte(value), &definition); err != nil {
			return
		}
		if err := assistant.ValidateAssistantDefinition(definition); err != nil {
			return
		}
		encoded, err := json.Marshal(definition)
		if err != nil {
			t.Fatalf("valid AssistantDefinition failed to marshal: %v", err)
		}
		var roundTrip assistant.AssistantDefinition
		if err := json.Unmarshal(encoded, &roundTrip); err != nil {
			t.Fatalf("marshaled AssistantDefinition failed to decode: %v", err)
		}
		if err := assistant.ValidateAssistantDefinition(roundTrip); err != nil {
			t.Fatalf("valid AssistantDefinition became invalid after JSON round trip: %v", err)
		}
	})
}
