# Assistant Adapter Conformance

## Purpose

`assistant/adaptertest` is the Go integration harness for concrete implementations of `assistant.ProviderAdapter`.

The harness exists to test the portable boundary against materially different providers, gateways, and local runtimes without adding provider SDKs, HTTP payloads, credentials, or volatile model tables to Gymkhana Core.

This is implementation/testing infrastructure. It does not change Core Spec `0.4` semantics.

## What the harness validates

For one explicitly selected generation model, the harness exercises the adapter through the public `AdapterRegistry` boundary and checks:

- provider descriptor validity and provider identity consistency;
- credential-reference compatibility through the existing registry validation path;
- live model catalog normalization and presence of the selected model;
- generation-role and text-capability metadata for the baseline smoke model;
- one low-cost normalized text generation request;
- additional adapter-owned probes for tools, structured output, and multimodal request shapes;
- optional normalized quota observations when the adapter implements `QuotaProvider`;
- representative vendor/transport error mappings when the adapter implements `FailureClassifier`.

Additional probes declare the portable capabilities they intend to exercise. The harness rejects a probe when the selected model does not advertise those capabilities instead of silently testing an unsupported request shape.

## What the harness does not own

The harness does not:

- embed OpenAI, Anthropic, Google, Ollama, gateway, or other SDKs;
- store or manufacture API keys;
- freeze provider model IDs, pricing, limits, or availability in Core;
- define provider-specific prompts, retry payloads, HTTP semantics, or deployment topology;
- turn live-provider behavior into new portable semantics automatically;
- require provider network calls from Core's own unit-test suite.

Concrete adapters may live in independent packages or repositories. Their integration tests import `assistant/adaptertest` and supply their adapter, selected model, optional opaque credential reference, and provider-specific probes.

## Recommended provider-family validation

The adapter boundary should be stress-tested against at least three materially different families before it is considered mature:

1. **OpenAI-compatible HTTP family / gateway** — validates common chat/tool/structured-output translation while ensuring compatibility layers do not leak into Core types.
2. **A non-OpenAI-native family such as Anthropic or Google** — stresses materially different role, tool, structured-output, usage, and error semantics.
3. **A local runtime such as Ollama** — stresses credential-free/local access, model discovery, local capability metadata, and absence of cloud quota assumptions.

A gateway such as OpenRouter or a second local/runtime family can be added when it exposes a materially different behavior rather than merely increasing provider count.

## Suggested external test shape

```go
func TestAdapterConformance(t *testing.T) {
    adaptertest.Run(t, adaptertest.Config{
        Adapter: adapter,
        Model: assistant.ModelRef{
            Provider: "provider-id",
            Model:    "selected-model",
        },
        Credential: credentialRef,
        Probes: []adaptertest.Probe{
            // Provider-specific integration probes using portable request types.
        },
    })
}
```

Live credentials and model selection remain test-environment configuration owned by the adapter project.

## Completion criteria for roadmap step 6

The adapter-validation roadmap step is complete when:

- the shared harness is stable;
- at least three materially different provider families pass it;
- tool and structured-output paths have been exercised by adapters that genuinely support them;
- at least one BYOK/cloud adapter and one credential-free local adapter have been exercised;
- normalized failure mappings have representative coverage for auth, rate-limit/quota, timeout/network/unavailable, safety/invalid-request, and unsupported-capability behavior where those distinctions are observable;
- any portability gaps discovered by real adapters are resolved in the smallest appropriate layer, with Core Spec changes made only when portable semantics truly need to change.

Until those conditions are met, `assistant/adaptertest` is infrastructure for roadmap step 6 rather than evidence that step 6 itself is complete.
