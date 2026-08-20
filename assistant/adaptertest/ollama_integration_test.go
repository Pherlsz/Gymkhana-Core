package adaptertest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
	"github.com/Pherlsz/Gymkhana-Core/assistant/adaptertest"
)

type ollamaAdapter struct {
	baseURL          string
	client           *http.Client
	structuredModels map[assistant.ModelID]bool
}

func (adapter *ollamaAdapter) Descriptor() assistant.ProviderDescriptor {
	return assistant.ProviderDescriptor{ID: "ollama", DisplayName: "Ollama", CredentialModes: []assistant.CredentialMode{assistant.CredentialNone}}
}

func (adapter *ollamaAdapter) ListModels(ctx context.Context, credential *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	if credential != nil {
		return nil, errors.New("ollama does not accept credentials")
	}
	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := doJSON(ctx, adapter.client, http.MethodGet, adapter.baseURL+"/api/tags", nil, nil, &tags); err != nil {
		return nil, err
	}
	models := make([]assistant.ModelDescriptor, 0, len(tags.Models))
	for _, tagged := range tags.Models {
		var show struct {
			Capabilities []string       `json:"capabilities"`
			ModelInfo    map[string]any `json:"model_info"`
		}
		if err := doJSON(ctx, adapter.client, http.MethodPost, adapter.baseURL+"/api/show", nil, map[string]any{"model": tagged.Name}, &show); err != nil {
			return nil, err
		}
		capabilities := make([]assistant.Capability, 0, 4)
		roles := make([]assistant.ModelRole, 0, 1)
		for _, capability := range show.Capabilities {
			switch capability {
			case "completion":
				roles = append(roles, assistant.ModelRoleGeneration)
				capabilities = append(capabilities, assistant.CapabilityText)
			case "vision":
				capabilities = append(capabilities, assistant.CapabilityImageInput)
			case "tools":
				capabilities = append(capabilities, assistant.CapabilityToolCalling)
			}
		}
		id := assistant.ModelID(tagged.Name)
		if adapter.structuredModels[id] {
			capabilities = append(capabilities, assistant.CapabilityStructuredOutput)
		}
		if len(roles) == 0 {
			continue
		}
		models = append(models, assistant.ModelDescriptor{
			Ref:          assistant.ModelRef{Provider: "ollama", Model: id},
			DisplayName:  tagged.Name,
			Access:       assistant.AccessLocal,
			Roles:        roles,
			Capabilities: capabilities,
		})
	}
	return models, nil
}

func (adapter *ollamaAdapter) Generate(ctx context.Context, request assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	if request.Credential != nil {
		return assistant.GenerationResponse{}, errors.New("ollama does not accept credentials")
	}
	messages := make([]any, 0, len(request.Messages))
	for _, message := range request.Messages {
		text := ""
		for _, part := range message.Content {
			if part.Type == assistant.PartText {
				text += part.Text
			}
		}
		messages = append(messages, map[string]any{"role": string(message.Role), "content": text})
	}
	payload := map[string]any{"model": string(request.Model.Model), "messages": messages, "stream": false}
	if len(request.Tools) > 0 {
		tools := make([]any, 0, len(request.Tools))
		for _, tool := range request.Tools {
			tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": tool.Name, "description": tool.Description, "parameters": json.RawMessage(tool.InputSchema)}})
		}
		payload["tools"] = tools
	}
	if len(request.ResponseSchema) > 0 {
		payload["format"] = json.RawMessage(request.ResponseSchema)
	}
	var response struct {
		DoneReason string `json:"done_reason"`
		Message    struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string         `json:"name"`
					Arguments map[string]any `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
		PromptEvalCount int64 `json:"prompt_eval_count"`
		EvalCount       int64 `json:"eval_count"`
	}
	if err := doJSON(ctx, adapter.client, http.MethodPost, adapter.baseURL+"/api/chat", nil, payload, &response); err != nil {
		return assistant.GenerationResponse{}, err
	}
	parts := make([]assistant.ContentPart, 0, 1+len(response.Message.ToolCalls))
	if response.Message.Content != "" {
		parts = append(parts, assistant.ContentPart{Type: assistant.PartText, Text: response.Message.Content})
	}
	for index, call := range response.Message.ToolCalls {
		arguments, err := json.Marshal(call.Function.Arguments)
		if err != nil {
			return assistant.GenerationResponse{}, err
		}
		parts = append(parts, assistant.ContentPart{Type: assistant.PartToolCall, ToolCall: &assistant.ToolCall{ID: "ollama_call_" + strconv.Itoa(index), Name: call.Function.Name, Arguments: arguments}})
	}
	finish := assistant.FinishStop
	if len(response.Message.ToolCalls) > 0 {
		finish = assistant.FinishToolCalls
	} else if response.DoneReason == "length" {
		finish = assistant.FinishLength
	}
	return assistant.GenerationResponse{
		Model:        request.Model,
		Message:      assistant.Message{Role: assistant.RoleAssistant, Content: parts},
		FinishReason: finish,
		Usage:        assistant.Usage{InputTokens: response.PromptEvalCount, OutputTokens: response.EvalCount},
	}, nil
}

func (adapter *ollamaAdapter) ClassifyFailure(err error) assistant.FailureClass {
	return genericFailureClass(err)
}

func TestOllamaAdapterProtocol(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/api/tags":
			_, _ = writer.Write([]byte(`{"models":[{"name":"llama-test"}]}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/show":
			_, _ = writer.Write([]byte(`{"capabilities":["completion","tools"],"model_info":{"general.architecture":"llama"}}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/chat":
			var payload map[string]json.RawMessage
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Errorf("decode chat request: %v", err)
				http.Error(writer, "invalid test request", http.StatusBadRequest)
				return
			}
			if len(payload["tools"]) > 0 {
				_, _ = writer.Write([]byte(`{"model":"llama-test","message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"lookup_weather","arguments":{"city":"Porto Alegre"}}}]},"done":true,"done_reason":"stop","prompt_eval_count":8,"eval_count":3}`))
				return
			}
			if len(payload["format"]) > 0 {
				_, _ = writer.Write([]byte(`{"model":"llama-test","message":{"role":"assistant","content":"{\"ok\":true}"},"done":true,"done_reason":"stop","prompt_eval_count":6,"eval_count":4}`))
				return
			}
			_, _ = writer.Write([]byte(`{"model":"llama-test","message":{"role":"assistant","content":"OK"},"done":true,"done_reason":"stop","prompt_eval_count":3,"eval_count":1}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	adaptertest.Run(t, adaptertest.Config{
		Adapter: &ollamaAdapter{
			baseURL:          server.URL,
			client:           server.Client(),
			structuredModels: map[assistant.ModelID]bool{"llama-test": true},
		},
		Model:  assistant.ModelRef{Provider: "ollama", Model: "llama-test"},
		Probes: commonProbes(),
		FailureProbes: []adaptertest.FailureProbe{
			{Name: "model-missing", Err: providerHTTPError{status: http.StatusNotFound}, Want: assistant.FailureInvalid},
			{Name: "unavailable", Err: providerHTTPError{status: http.StatusServiceUnavailable}, Want: assistant.FailureUnavailable},
		},
	})
}
