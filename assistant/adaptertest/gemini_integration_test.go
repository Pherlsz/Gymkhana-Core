package adaptertest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
	"github.com/Pherlsz/Gymkhana-Core/assistant/adaptertest"
)

type geminiAdapter struct {
	baseURL  string
	client   *http.Client
	resolve  credentialResolver
	profiles map[assistant.ModelID][]assistant.Capability
}

func (adapter *geminiAdapter) Descriptor() assistant.ProviderDescriptor {
	return assistant.ProviderDescriptor{ID: "google", DisplayName: "Google Gemini", CredentialModes: []assistant.CredentialMode{assistant.CredentialBYOK}}
}

func (adapter *geminiAdapter) ListModels(ctx context.Context, credential *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	secret, err := adapter.resolve(credential)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			InputTokenLimit            int64    `json:"inputTokenLimit"`
			OutputTokenLimit           int64    `json:"outputTokenLimit"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := doJSON(ctx, adapter.client, http.MethodGet, adapter.baseURL+"/v1beta/models", map[string]string{"x-goog-api-key": secret}, nil, &payload); err != nil {
		return nil, err
	}
	models := make([]assistant.ModelDescriptor, 0, len(payload.Models))
	for _, item := range payload.Models {
		if !containsString(item.SupportedGenerationMethods, "generateContent") {
			continue
		}
		id := assistant.ModelID(strings.TrimPrefix(item.Name, "models/"))
		capabilities, ok := adapter.profiles[id]
		if !ok {
			continue
		}
		models = append(models, assistant.ModelDescriptor{
			Ref:             assistant.ModelRef{Provider: "google", Model: id},
			DisplayName:     item.DisplayName,
			Access:          assistant.AccessUnknown,
			Roles:           []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities:    capabilities,
			ContextWindow:   item.InputTokenLimit,
			MaxOutputTokens: item.OutputTokenLimit,
		})
	}
	return models, nil
}

func (adapter *geminiAdapter) Generate(ctx context.Context, request assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	secret, err := adapter.resolve(request.Credential)
	if err != nil {
		return assistant.GenerationResponse{}, err
	}
	contents := make([]any, 0, len(request.Messages))
	var systemParts []any
	for _, message := range request.Messages {
		parts := make([]any, 0, len(message.Content))
		for _, part := range message.Content {
			if part.Type == assistant.PartText {
				parts = append(parts, map[string]any{"text": part.Text})
			}
		}
		if len(parts) == 0 {
			continue
		}
		switch message.Role {
		case assistant.RoleSystem, assistant.RoleDeveloper:
			systemParts = append(systemParts, parts...)
		case assistant.RoleAssistant:
			contents = append(contents, map[string]any{"role": "model", "parts": parts})
		default:
			contents = append(contents, map[string]any{"role": "user", "parts": parts})
		}
	}
	payload := map[string]any{"contents": contents}
	if len(systemParts) > 0 {
		payload["systemInstruction"] = map[string]any{"parts": systemParts}
	}
	if len(request.Tools) > 0 {
		declarations := make([]any, 0, len(request.Tools))
		for _, tool := range request.Tools {
			declarations = append(declarations, map[string]any{"name": tool.Name, "description": tool.Description, "parameters": json.RawMessage(tool.InputSchema)})
		}
		payload["tools"] = []any{map[string]any{"functionDeclarations": declarations}}
	}
	if len(request.ResponseSchema) > 0 {
		payload["generationConfig"] = map[string]any{"responseMimeType": "application/json", "responseJsonSchema": json.RawMessage(request.ResponseSchema)}
	}
	var response struct {
		Candidates []struct {
			FinishReason string `json:"finishReason"`
			Content      struct {
				Parts []struct {
					Text         string `json:"text"`
					FunctionCall *struct {
						ID   string         `json:"id"`
						Name string         `json:"name"`
						Args map[string]any `json:"args"`
					} `json:"functionCall"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Usage struct {
			PromptTokens    int64 `json:"promptTokenCount"`
			CandidateTokens int64 `json:"candidatesTokenCount"`
			CachedTokens    int64 `json:"cachedContentTokenCount"`
			ReasoningTokens int64 `json:"thoughtsTokenCount"`
		} `json:"usageMetadata"`
	}
	endpoint := adapter.baseURL + "/v1beta/models/" + string(request.Model.Model) + ":generateContent"
	if err := doJSON(ctx, adapter.client, http.MethodPost, endpoint, map[string]string{"x-goog-api-key": secret}, payload, &response); err != nil {
		return assistant.GenerationResponse{}, err
	}
	if len(response.Candidates) != 1 {
		return assistant.GenerationResponse{}, errors.New("expected exactly one candidate")
	}
	candidate := response.Candidates[0]
	parts := make([]assistant.ContentPart, 0, len(candidate.Content.Parts))
	toolCalls := 0
	for index, part := range candidate.Content.Parts {
		if part.FunctionCall != nil {
			arguments, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				return assistant.GenerationResponse{}, err
			}
			id := part.FunctionCall.ID
			if id == "" {
				id = "gemini_call_" + strconv.Itoa(index)
			}
			parts = append(parts, assistant.ContentPart{Type: assistant.PartToolCall, ToolCall: &assistant.ToolCall{ID: id, Name: part.FunctionCall.Name, Arguments: arguments}})
			toolCalls++
		} else if part.Text != "" {
			parts = append(parts, assistant.ContentPart{Type: assistant.PartText, Text: part.Text})
		}
	}
	finish := assistant.FinishOther
	if toolCalls > 0 {
		finish = assistant.FinishToolCalls
	} else {
		switch candidate.FinishReason {
		case "STOP":
			finish = assistant.FinishStop
		case "MAX_TOKENS":
			finish = assistant.FinishLength
		case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT":
			finish = assistant.FinishContentFilter
		}
	}
	return assistant.GenerationResponse{
		Model:        request.Model,
		Message:      assistant.Message{Role: assistant.RoleAssistant, Content: parts},
		FinishReason: finish,
		Usage: assistant.Usage{
			InputTokens:       response.Usage.PromptTokens,
			OutputTokens:      response.Usage.CandidateTokens,
			CachedInputTokens: response.Usage.CachedTokens,
			ReasoningTokens:   response.Usage.ReasoningTokens,
		},
	}, nil
}

func (adapter *geminiAdapter) ClassifyFailure(err error) assistant.FailureClass {
	return genericFailureClass(err)
}

func TestGeminiAdapterProtocol(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("x-goog-api-key") != "test-secret" {
			http.Error(writer, `{"error":{"code":"UNAUTHENTICATED"}}`, http.StatusUnauthorized)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/v1beta/models":
			_, _ = writer.Write([]byte(`{"models":[{"name":"models/gemini-test","displayName":"Gemini Test","inputTokenLimit":32768,"outputTokenLimit":8192,"supportedGenerationMethods":["generateContent"]}]}`))
		case request.Method == http.MethodPost && request.URL.Path == "/v1beta/models/gemini-test:generateContent":
			var payload map[string]json.RawMessage
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode generateContent request: %v", err)
			}
			if len(payload["tools"]) > 0 {
				_, _ = writer.Write([]byte(`{"candidates":[{"finishReason":"STOP","content":{"role":"model","parts":[{"functionCall":{"id":"call_1","name":"lookup_weather","args":{"city":"Porto Alegre"}}}]}}],"usageMetadata":{"promptTokenCount":8,"candidatesTokenCount":3}}`))
				return
			}
			if len(payload["generationConfig"]) > 0 {
				_, _ = writer.Write([]byte(`{"candidates":[{"finishReason":"STOP","content":{"role":"model","parts":[{"text":"{\"ok\":true}"}]}}],"usageMetadata":{"promptTokenCount":6,"candidatesTokenCount":4}}`))
				return
			}
			_, _ = writer.Write([]byte(`{"candidates":[{"finishReason":"STOP","content":{"role":"model","parts":[{"text":"OK"}]}}],"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":1}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	credential := &assistant.CredentialRef{ID: "test-key", Provider: "google", Mode: assistant.CredentialBYOK, Reference: "test:key"}
	adaptertest.Run(t, adaptertest.Config{
		Adapter: &geminiAdapter{
			baseURL: server.URL,
			client:  server.Client(),
			resolve: staticCredentialResolver("test:key", "test-secret"),
			profiles: map[assistant.ModelID][]assistant.Capability{
				"gemini-test": {assistant.CapabilityText, assistant.CapabilityToolCalling, assistant.CapabilityStructuredOutput},
			},
		},
		Credential: credential,
		Model:      assistant.ModelRef{Provider: "google", Model: "gemini-test"},
		Probes:     commonProbes(),
		FailureProbes: []adaptertest.FailureProbe{
			{Name: "auth", Err: providerHTTPError{status: http.StatusForbidden}, Want: assistant.FailureAuth},
			{Name: "rate-limit", Err: providerHTTPError{status: http.StatusTooManyRequests, code: "RESOURCE_EXHAUSTED"}, Want: assistant.FailureRateLimit},
			{Name: "timeout", Err: providerHTTPError{status: http.StatusGatewayTimeout}, Want: assistant.FailureTimeout},
			{Name: "invalid", Err: providerHTTPError{status: http.StatusBadRequest, code: "INVALID_ARGUMENT"}, Want: assistant.FailureInvalid},
		},
	})
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
