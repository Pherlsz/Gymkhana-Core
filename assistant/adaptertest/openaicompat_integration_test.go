package adaptertest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
	"github.com/Pherlsz/Gymkhana-Core/assistant/adaptertest"
)

type openAICompatibleAdapter struct {
	baseURL  string
	client   *http.Client
	resolve  credentialResolver
	profiles map[assistant.ModelID][]assistant.Capability
}

func (adapter *openAICompatibleAdapter) Descriptor() assistant.ProviderDescriptor {
	return assistant.ProviderDescriptor{ID: "openai", DisplayName: "OpenAI-compatible", CredentialModes: []assistant.CredentialMode{assistant.CredentialBYOK}}
}

func (adapter *openAICompatibleAdapter) ListModels(ctx context.Context, credential *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	secret, err := adapter.resolve(credential)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := doJSON(ctx, adapter.client, http.MethodGet, adapter.baseURL+"/v1/models", map[string]string{"Authorization": "Bearer " + secret}, nil, &payload); err != nil {
		return nil, err
	}
	models := make([]assistant.ModelDescriptor, 0, len(payload.Data))
	for _, item := range payload.Data {
		capabilities, ok := adapter.profiles[assistant.ModelID(item.ID)]
		if !ok {
			continue
		}
		models = append(models, assistant.ModelDescriptor{
			Ref:          assistant.ModelRef{Provider: "openai", Model: assistant.ModelID(item.ID)},
			DisplayName:  item.ID,
			Access:       assistant.AccessUnknown,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: capabilities,
		})
	}
	return models, nil
}

func (adapter *openAICompatibleAdapter) Generate(ctx context.Context, request assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	secret, err := adapter.resolve(request.Credential)
	if err != nil {
		return assistant.GenerationResponse{}, err
	}
	type wireMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	wireMessages := make([]wireMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		text := ""
		for _, part := range message.Content {
			if part.Type == assistant.PartText {
				text += part.Text
			}
		}
		wireMessages = append(wireMessages, wireMessage{Role: string(message.Role), Content: text})
	}
	payload := map[string]any{"model": string(request.Model.Model), "messages": wireMessages, "stream": false}
	if len(request.Tools) > 0 {
		tools := make([]any, 0, len(request.Tools))
		for _, tool := range request.Tools {
			tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": tool.Name, "description": tool.Description, "parameters": json.RawMessage(tool.InputSchema)}})
		}
		payload["tools"] = tools
	}
	if len(request.ResponseSchema) > 0 {
		payload["response_format"] = map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "response", "strict": true, "schema": json.RawMessage(request.ResponseSchema)}}
	}
	var response struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := doJSON(ctx, adapter.client, http.MethodPost, adapter.baseURL+"/v1/chat/completions", map[string]string{"Authorization": "Bearer " + secret}, payload, &response); err != nil {
		return assistant.GenerationResponse{}, err
	}
	if len(response.Choices) != 1 {
		return assistant.GenerationResponse{}, errors.New("expected exactly one choice")
	}
	choice := response.Choices[0]
	parts := make([]assistant.ContentPart, 0, 1+len(choice.Message.ToolCalls))
	if choice.Message.Content != "" {
		parts = append(parts, assistant.ContentPart{Type: assistant.PartText, Text: choice.Message.Content})
	}
	for _, call := range choice.Message.ToolCalls {
		parts = append(parts, assistant.ContentPart{Type: assistant.PartToolCall, ToolCall: &assistant.ToolCall{ID: call.ID, Name: call.Function.Name, Arguments: call.Function.Arguments}})
	}
	finish := assistant.FinishOther
	switch choice.FinishReason {
	case "stop":
		finish = assistant.FinishStop
	case "length":
		finish = assistant.FinishLength
	case "content_filter":
		finish = assistant.FinishContentFilter
	case "tool_calls":
		finish = assistant.FinishToolCalls
	}
	return assistant.GenerationResponse{
		Model:        request.Model,
		Message:      assistant.Message{Role: assistant.RoleAssistant, Content: parts},
		FinishReason: finish,
		Usage:        assistant.Usage{InputTokens: response.Usage.PromptTokens, OutputTokens: response.Usage.CompletionTokens},
	}, nil
}

func (adapter *openAICompatibleAdapter) ClassifyFailure(err error) assistant.FailureClass {
	return genericFailureClass(err)
}

func TestOpenAICompatibleAdapterProtocol(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer test-secret" {
			http.Error(writer, `{"error":{"type":"authentication_error"}}`, http.StatusUnauthorized)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/v1/models":
			_, _ = writer.Write([]byte(`{"data":[{"id":"gpt-test"}]}`))
		case request.Method == http.MethodPost && request.URL.Path == "/v1/chat/completions":
			var payload map[string]json.RawMessage
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode chat request: %v", err)
			}
			if len(payload["tools"]) > 0 {
				_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"tool_calls","message":{"content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup_weather","arguments":"{\"city\":\"Porto Alegre\"}"}}]}}],"usage":{"prompt_tokens":8,"completion_tokens":3}}`))
				return
			}
			if len(payload["response_format"]) > 0 {
				_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"{\"ok\":true}"}}],"usage":{"prompt_tokens":6,"completion_tokens":4}}`))
				return
			}
			_, _ = writer.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"content":"OK"}}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	credential := &assistant.CredentialRef{ID: "test-key", Provider: "openai", Mode: assistant.CredentialBYOK, Reference: "test:key"}
	adaptertest.Run(t, adaptertest.Config{
		Adapter: &openAICompatibleAdapter{
			baseURL: server.URL,
			client:  server.Client(),
			resolve: staticCredentialResolver("test:key", "test-secret"),
			profiles: map[assistant.ModelID][]assistant.Capability{
				"gpt-test": {assistant.CapabilityText, assistant.CapabilityToolCalling, assistant.CapabilityStructuredOutput},
			},
		},
		Credential: credential,
		Model:      assistant.ModelRef{Provider: "openai", Model: "gpt-test"},
		Probes:     commonProbes(),
		FailureProbes: []adaptertest.FailureProbe{
			{Name: "auth", Err: providerHTTPError{status: http.StatusUnauthorized}, Want: assistant.FailureAuth},
			{Name: "rate-limit", Err: providerHTTPError{status: http.StatusTooManyRequests, code: "rate_limit_exceeded"}, Want: assistant.FailureRateLimit},
			{Name: "quota", Err: providerHTTPError{status: http.StatusTooManyRequests, code: "insufficient_quota"}, Want: assistant.FailureQuota},
			{Name: "unavailable", Err: providerHTTPError{status: http.StatusServiceUnavailable}, Want: assistant.FailureUnavailable},
		},
	})
}
