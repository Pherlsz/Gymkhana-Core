package adaptertest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
	"github.com/Pherlsz/Gymkhana-Core/assistant/adaptertest"
)

var (
	toolSchema   = json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"],"additionalProperties":false}`)
	resultSchema = json.RawMessage(`{"type":"object","properties":{"ok":{"type":"boolean"}},"required":["ok"],"additionalProperties":false}`)
)

type providerHTTPError struct {
	status int
	code   string
}

func (e providerHTTPError) Error() string {
	if e.code == "" {
		return fmt.Sprintf("provider request failed with status %d", e.status)
	}
	return fmt.Sprintf("provider request failed with status %d (%s)", e.status, e.code)
}

type credentialResolver func(*assistant.CredentialRef) (string, error)

func staticCredentialResolver(wantRef, secret string) credentialResolver {
	return func(ref *assistant.CredentialRef) (string, error) {
		if ref == nil || ref.Reference != wantRef {
			return "", errors.New("unknown credential reference")
		}
		return secret, nil
	}
}

func doJSON(ctx context.Context, client *http.Client, method, endpoint string, headers map[string]string, requestBody any, responseBody any) error {
	var body io.Reader
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var payload struct {
			Error struct {
				Code string `json:"code"`
				Type string `json:"type"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload)
		code := payload.Error.Code
		if code == "" {
			code = payload.Error.Type
		}
		return providerHTTPError{status: response.StatusCode, code: code}
	}
	if responseBody == nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(responseBody); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func genericFailureClass(err error) assistant.FailureClass {
	if err == nil {
		return assistant.FailureUnknown
	}
	var httpErr providerHTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.status {
		case http.StatusUnauthorized, http.StatusForbidden:
			return assistant.FailureAuth
		case http.StatusRequestTimeout, http.StatusGatewayTimeout:
			return assistant.FailureTimeout
		case http.StatusTooManyRequests:
			if strings.Contains(strings.ToLower(httpErr.code), "quota") {
				return assistant.FailureQuota
			}
			return assistant.FailureRateLimit
		case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity:
			return assistant.FailureInvalid
		default:
			if httpErr.status >= 500 {
				return assistant.FailureUnavailable
			}
		}
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return assistant.FailureNetwork
	}
	return assistant.FailureUnknown
}

func commonProbes() []adaptertest.Probe {
	return []adaptertest.Probe{
		{
			Name:                 "tool-calling",
			RequiredCapabilities: []assistant.Capability{assistant.CapabilityToolCalling},
			Messages: []assistant.Message{{
				Role:    assistant.RoleUser,
				Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "What is the weather?"}},
			}},
			Tools: []assistant.ToolDefinition{{Name: "lookup_weather", Description: "Look up weather", InputSchema: toolSchema}},
			Check: func(response assistant.GenerationResponse) error {
				if response.FinishReason != assistant.FinishToolCalls || len(response.Message.Content) != 1 || response.Message.Content[0].ToolCall == nil {
					return errors.New("expected one normalized tool call")
				}
				if response.Message.Content[0].ToolCall.Name != "lookup_weather" {
					return fmt.Errorf("tool name = %q", response.Message.Content[0].ToolCall.Name)
				}
				return nil
			},
		},
		{
			Name:                 "structured-output",
			RequiredCapabilities: []assistant.Capability{assistant.CapabilityStructuredOutput},
			Messages: []assistant.Message{{
				Role:    assistant.RoleUser,
				Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "Return whether the check passed."}},
			}},
			ResponseSchema: resultSchema,
			Check: func(response assistant.GenerationResponse) error {
				if response.FinishReason != assistant.FinishStop || len(response.Message.Content) != 1 || response.Message.Content[0].Text != `{"ok":true}` {
					return errors.New("expected normalized structured JSON text")
				}
				return nil
			},
		},
	}
}
