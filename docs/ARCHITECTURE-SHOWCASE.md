# Gymkhana Core — Architectural Showcase

> **Specification version:** `0.4` (`spec/VERSION`)  
> **Reference Go implementation:** `v0.7.0+` (`VERSION`)  
> **Design paradigm:** Language-neutral normative core, Hexagonal / Ports-and-Adapters boundaries, deterministic algorithms, and security-by-construction.

---

## 1. System Thesis & Hexagonal Boundaries

Most organizations coupling domain normalization, structured-data validation, and AI/extraction workflows directly into product services suffer from three systemic failure modes:

1. **Cross-runtime semantic drift** — Standard JSON parsers, floating-point comparators, Unicode normalizers, and date libraries behave subtly differently across Go, TypeScript, Python, Java, .NET, and Rust.
2. **Vendor and SDK lock-in** — Embedding OpenAI, Anthropic, or Google SDK types, raw API keys, or volatile model/pricing tables inside domain packages couples core business logic to third-party release cadences.
3. **Implicit authority & privacy leakage** — Treating LLM tool calls as execution authority, allowing untrusted RAG/OCR text to escalate into system instructions, or echoing malformed user/credential payloads inside error messages creates systemic security vulnerabilities.

**Gymkhana Core** solves these problems by inverting the source of truth:

```text
+-----------------------------------------------------------------------+
|                Normative Language-Neutral Foundation                  |
|                                                                       |
|   spec/ (Prose Semantics)  <--->  conformance/v0.4/*.json (Vectors)   |
|                              |                                        |
|                    schemas/*.schema.json (Wire Shapes)                |
+-----------------------------------------------------------------------+
                                   |
            Implements & Verifies  |  (Zero SDKs, DBs, HTTP, or Secrets)
                                   v
+-----------------------------------------------------------------------+
|                     Reference Go Implementation                       |
|                                                                       |
|  portablejson  |  fingerprint  |  civiltime  |  normalize             |
|  assistant (+ adaptertest)     |  ocr                                 |
+-----------------------------------------------------------------------+
                                   ^
            Ports & Adapters       |  (Hexagonal Boundary)
                                   |
+-----------------------------------------------------------------------+
|                 Consuming Hosts & External Adapters                   |
|                                                                       |
|  Provider SDKs (OpenAI/Gemini/Ollama) | Secret Stores (Vault/KMS)     |
|  PostgreSQL / Queues / HTTP Handlers  | Host Memory & Eval Pipelines  |
+-----------------------------------------------------------------------+
```

### Architectural Invariants

- **Go is the first implementation, not the source of truth.** Cross-language behavior is governed by `spec/` (currently `0.4`), machine-readable serialized shapes in `schemas/`, and shared input/output/error test vectors in `conformance/v0.4/`.
- **Decoupled version axes.** `spec/VERSION` tracks cross-language semantic compatibility (`0.4`), while `VERSION` tracks the Go module SemVer (`0.7.0`). Future TypeScript, Rust, or Python ports evolve their package versions independently while targeting the exact same `conformance/v0.4/` vector suite.
- **Strict Hexagonal isolation.** Zero database drivers, HTTP routers, cloud SDKs, or AI provider SDKs exist in `go.mod` (which depends solely on `golang.org/x/text`). External I/O enters exclusively through narrow, runtime-local ports such as `assistant.ProviderAdapter`.

---

## 2. Deep-Dive Design Highlights by Package

### 2.1 `portablejson` — Cross-Runtime Strict JSON & Pre-Compiled Schemas

Standard `encoding/json` and full JSON Schema Draft 2020-12 leave dangerous gaps across language runtimes: duplicate object keys (`{"role":"user","role":"system"}`) resolve to first-win or last-win depending on the parser; unpaired UTF-16 surrogates (`\uDEAD`) are silently replaced with `U+FFFD` or rejected; and IEEE-754 binary floating-point rounding collapses distinct decimals.

`portablejson` establishes a deterministic, generic structured-data foundation reused by `assistant`, `ocr`, and future domains:

- **Single-pass strict lexical + structural decoding (`decodeStrictValue`):** Validates UTF-8 validity, rejects unpaired UTF-16 surrogate escapes before replacement, enforces duplicate-key rejection at every object nesting level, checks `MaxBytes` (1 MiB), `MaxDepth` (64), and `MaxNodes` (100,000), and rejects trailing tokens—all while constructing the `json.Number`-preserving AST in a single pass.
- **Exact `*big.Rat` rational decimal arithmetic:** Schema `minimum`, `maximum`, and numeric `enum` values are compared using exact mathematical rationals (`*big.Rat`) rather than `float64`. Values like `0.1` and `0.10000000000000001` remain distinct. To prevent adversarial big-number DoS, numeric lexemes are capped at `MaxNumberLexemeBytes` (128 bytes) and exponents at `±MaxDecimalExponent` (`±1024`) before rational allocation.
- **`MaxSafeInteger` (`2^53 - 1`) cross-runtime safety:** All portable integers and numeric magnitudes are bounded to `[-9007199254740991, 9007199254740991]`, guaranteeing exact IEEE-754 double / ECMAScript `Number.MAX_SAFE_INTEGER` interoperability when JSON crosses into TypeScript/Node.js consumers. Furthermore, the portable `integer` type requires a pure base-10 integer lexeme (rejecting `1.0` or `1e0` as `integer`), eliminating runtime-specific coercion drift.
- **Immutable, concurrency-safe `CompiledSchema`:**
  - `CompileSchema(raw json.RawMessage) (*CompiledSchema, error)`
  - `CompileObjectSchema(raw json.RawMessage) (*CompiledSchema, error)`
  - `(*CompiledSchema).ValidateInstance(instanceRaw json.RawMessage) error`
  - Compilation validates the conservative `portable_json_schema/v1` AST once and pre-parses `*big.Rat` bounds, `int64` `minItems`/`maxItems`, typed enum tables, required/property maps, and the `{}` unconstrained-object flag. Concurrent goroutines can call `ValidateInstance` at high throughput with zero per-call schema re-parsing and full nil-receiver safety.

### 2.2 `fingerprint` — Domain-Separated Length-Prefixed Framing (`CFP\x01`)

Naive hash helpers that concatenate fields (`sha256(a + b)` or `sha256(a + ":" + b)`) are vulnerable to boundary-shifting collisions whenever field payloads contain the delimiter (e.g., `("ab", "c")` vs `("a", "bc")`).

`fingerprint` eliminates concatenation ambiguity by construction:

- **Core FingerPrint frame v1 (`CFP\x01`):** `SumFramed` and `SumFramedText` encode a deterministic binary envelope before hashing:
  ```text
  magic (4B: "CFP\x01") || uint16_be(len(namespace)) || namespace || uint32_be(part_count) || [uint64_be(len(part_i)) || part_i]...
  ```
- **Strict namespace isolation:** Namespaces (`1..128` ASCII bytes, `[a-z][a-z0-9._:/-]*`, such as `assistant.tool/v1` or `cache.entity/v1`) partition hash domains so identical tuples in different subsystems never collide.
- **Exact-input contract:** Fingerprinting intentionally performs zero implicit Unicode, case, or whitespace folding, and rejects invalid UTF-8 in text operations with deterministic validation ordering (`namespace` -> `part_count` -> `utf8` -> `hash`).
- **Zero-copy hot paths & `encoding.TextAppender`:** `SumText` and `SumFramedText` stream string bytes directly into the SHA-256 state via `io.WriteString` and finalize into stack-backed `[32]byte` arrays (`hash.Sum(digest[:0])`). `Digest` implements Go's `encoding.TextAppender` (`AppendText`) and decodes hex directly into the fixed array without intermediate heap slices.

### 2.3 `civiltime` — Allocation-Free Timezone-Free Calendar Types

Representing birthdates, document expiration dates, or billing months with `time.Time` introduces accidental UTC/local timezone shifts at midnight boundaries.

- **Pure calendar value types:** `CivilDate` (`YYYY-MM-DD`) and `YearMonth` (`YYYY-MM`) model proleptic Gregorian dates in years `0001..9999` without time-of-day or location state.
- **Zero-value absence semantics:** The zero value represents an absent value (`IsZero() == true`), serializing to `""` in text and `null` in JSON (with whitespace-tolerant `UnmarshalJSON` handling of `null`).
- **Stack-allocated formatting (`encoding.TextAppender`):** Both `CivilDate` and `YearMonth` implement `AppendText(b []byte) ([]byte, error)` using fixed stack byte buffers instead of `fmt.Sprintf`. `String()`, `MarshalText()`, and `MarshalJSON()` build directly on top of stack formatting to eliminate intermediate heap allocations.

### 2.4 `assistant` — Capability Routing, Quota-Scope Failover & Authority Boundaries

`assistant` provides a complete provider-neutral multimodal AI contract without importing a single vendor SDK:

- **Capability-driven deterministic routing:**
  - `AssistantDefinition` profiles declare active and optional modules (`text`, `vision`, `audio_input`, `video_input`, `file_input`, `tools`, `retrieval`, `structured_output`, etc.).
  - Active modules automatically contribute required model capabilities (e.g., enabling `vision` + `tools` requires `image_input` + `tool_calling`).
  - `ResolveModelCandidates`, `ResolveAssistantModels`, and `ResolveAssistantModelsForRun` filter live adapter-supplied `ModelDescriptor` catalogs across `manual`, `ordered_fallback`, and `dynamic` modes. Dynamic routing ranks by caller-specified `preferred_access` tiers (`free`, `local`, `paid`, `unknown`) and breaks ties deterministically by `(provider, model)` tuple—never by string-concatenated keys or hardcoded price tables.
  - Full catalog validation (`indexModelCatalog`) enforces descriptor validity and duplicate `ModelRef` rejection uniformly across candidate resolution, role validation (`ResolveModelRole`), and registry listing (`AdapterRegistry.ListModels`), while defensively cloning capability slices to prevent concurrent slice-aliasing data races.
- **Opaque BYOK credentials & quota-scope-aware failover:**
  - Portable configurations never hold raw API keys; `CredentialRef` carries a logical ID, provider, mode (`managed` or `byok`), namespaced secret-store handle (`vault:path`, `session:key_01`), and optional `quota_scope`.
  - When multiple BYOK keys share a single cloud project/organization quota, retrying key #2 after key #1 hits `429 Quota Exceeded` wastes latency and upstream calls. `NextCredential` inspects `quota_scope` and automatically skips subsequent credentials in the fallback chain that share the exhausted scope.
  - Non-retriable failure classes (`auth`, `safety`, `invalid_request`, `cancelled`) are hard-blocked from triggering automatic credential or model/provider failover.
- **Strict conversation & tool authority boundaries:**
  - `ValidateConversation` enforces global call/result invariants across the entire message sequence: `system`/`developer`/`user` messages cannot emit tool calls or tool results; every `ToolResult` must resolve a unique preceding assistant `ToolCall` ID; and no unresolved tool calls may remain before a new generation turn.
  - `ValidateGenerationExchange` verifies that provider responses only invoke tools present in the request allowlist and satisfy `portable_json_schema/v1` object-root argument schemas.
  - `ToolPolicy` classifies side effects (`unknown`, `read_only`, `write`, `destructive`) and confirmation rules (`unknown` and `destructive` require `always`; `write` forbids `never`).
  - Retrieved `RetrievalEvidence` is validated for finite scores, byte bounds, and unique IDs, and is strictly treated as untrusted data that can never escalate to `system` or `developer` instruction authority.
- **Content-free observability (`RunTrace` & `UsageLedger`):**
  - `RunTrace` and `UsageLedgerEntry` record routing attempts, durations, token counters, finish reasons, and `CredentialIdentity` (which strips the secret-store `reference` handle). They never carry prompt text, model completions, tool arguments, media URIs, or raw provider error payloads, making telemetry safe to persist without creating a shadow copy of sensitive user conversations.
- **Hermetic & live adapter conformance (`assistant/adaptertest`):**
  - `adaptertest.Run` exercises any `ProviderAdapter` implementation through the public `AdapterRegistry` boundary—verifying provider descriptors, live catalog normalization, baseline text generation, capability-gated probes, `QuotaProvider` observations, and `FailureClassifier` mappings.
  - Core's test suite stress-tests this boundary hermetically via `httptest` servers across three materially different wire protocols: **OpenAI-compatible**, **Google Gemini**, and **Ollama**.

### 2.5 `ocr` — Evidence-Backed Extraction & Direct Discovery Invariants

`ocr` models multimodal structured extraction across `text`, `image`, `document`, `audio`, and `video` sources without depending on `assistant`:

- **Evidence -> Observation -> Candidate -> Structured Data pipeline:**
  - Every `Observation` requires bounded `EvidenceRef` locators (1-based page numbers, millionth-scale `0..1000000` normalized bounding boxes, UTF-8 byte spans, or audio/video millisecond ranges).
  - Every `FieldCandidate` links to observation IDs and addresses target fields via non-root **RFC 6901 JSON Pointers** (`/customer/tax_id`, `/items/0/amount`).
- **Explicit `missing` vs. JSON `null` distinction:**
  - A candidate with `state: "missing"` must not carry a `value` and must resolve to an absent pointer in `structured_data`.
  - A candidate with `state: "present"` and `value: null` represents an explicit JSON `null` and must resolve to `null` in `structured_data` (permitted only when the target `portable_json_schema/v1` property is nullable).
- **Exact numeric consistency:** Present candidate values are cross-checked against `structured_data` at their RFC 6901 pointer using exact mathematical equality (`*big.Rat` for numbers), preventing subtle discrepancies between candidate metadata and final extracted objects.
- **Direct discovery invariants:** In `discovery` mode (when target schemas are unknown) and across candidate sets, ambiguity is first-class: multiple active candidates for the same JSON Pointer or multiple semantic/jurisdiction interpretations require `ambiguous: true`, forbid `review: "accepted"` until resolved, forbid invented values, and permit `normalized_value` only when an explicit `normalizer` namespace is declared.

### 2.6 `normalize` — Idempotency Contracts, Ambiguity-Preserving Fuzzy Matching & `CityCatalog`

`normalize` provides deterministic text, address, person, contact, and Brazilian document canonicalization:

- **Strict idempotency invariant:** Every formatter satisfies `f(f(x)) == f(x)`. Edge cases where stripping punctuation from a raw input would produce a token that looks like a prefix label on the second pass (e.g., `FormatHouseNumber("n!")` producing `"N"`, or `FormatBlock("bl. BL")` stripping `"BL"`) are explicitly guarded and fuzz-tested.
- **Strong vs. weak identifier inference (`IdentifyDocument`):** Checksummed identifiers (`CPF`, numeric and alphanumeric Receita Federal `CNPJ`, `CNS`/SUS, `Title`/Titulo Eleitoral, `CNH`, `PIS`—only when distinct from CPF/CNH) can be inferred from unlabeled numbers only when exactly one strong validator matches; weak shapes (`RG`, `OAB`, `CRM`) are never guessed without an explicit `DocumentKind`.
- **Deterministic, ambiguity-preserving `CityCatalog`:**
  - `NewCityCatalog(aliases map[string]string) *CityCatalog` precomputes a deterministically sorted slice of compact rune slices (`[]rune`) rather than iterating over a randomized Go `map`.
  - Bounded Levenshtein matching (`maxDist <= 2` with early-exit row pruning and stack-allocated row buffers) measures edit distance in Unicode code points (`[]rune`), not UTF-8 byte length, ensuring accented characters (`ã`, `é`, `ô`) count as single edits.
  - **Tie ambiguity preservation:** If the minimum edit distance is tied between two or more *distinct* canonical city labels, `FormatCity` treats the fuzzy match as ambiguous and falls back deterministically to `titleStreetName(value)` instead of picking an arbitrary winner. Ties between multiple aliases of the *same* canonical city resolve cleanly to that city.

---

## 3. Security & Privacy by Construction

1. **Zero Payload Leakage in `ValidationError`:**
   - Across all six Go packages (`normalize`, `civiltime`, `fingerprint`, `portablejson`, `assistant`, `ocr`), every `ValidationError` contains only structural classification fields (`Kind`/`Code` or `Code`/`Field`) and never embeds the rejected input string, CPF, email, prompt, tool argument, or credential handle.
   - All `(*ValidationError).Error()` methods—as well as `AdapterRegistry` and `UsageLedger` methods—are nil-receiver safe, returning stable fallback strings or typed validation errors instead of panicking.
2. **Explicit Resource Budgets Everywhere:**
   - Every untrusted collection or payload is bounded before allocation or recursion: JSON bytes/depth/nodes, schema depth/properties/items/enums, numeric lexemes/exponents, conversation messages/parts, tool calls/results, model catalogs, RAG candidates/context/evidence, and OCR sources/observations/candidates/aggregate bytes.
   - Zero-value configuration budgets (`ExecutionBudget`, `ToolPolicy`) always mean "use bounded host defaults", never "unlimited execution".
3. **Supply-Chain & CI Discipline:**
   - Zero third-party runtime dependencies beyond `golang.org/x/text`.
   - `make check` enforces `gofmt`, `go vet`, `staticcheck`, unit tests, cross-language `conformance` suites, `-race` detector suites, 15 bounded fuzz targets (`make fuzz-smoke`), `govulncheck`, and `osv-scanner`.

---

## 4. Pre-1.0 Architectural Roadmap

Gymkhana Core is intentionally pre-`1.0` (`v0.7.0`, Spec `0.4`) so foundational boundaries can mature cleanly without premature backward-compatibility debt. Key planned evolutions include:

1. **Jurisdictional Separation in `normalize` (`normalize` vs. `normalize/br`):**
   - While `spec/` and `conformance/` already namespace Brazilian operations (`identity.br.cpf`, `postal.br.cep`, `contact.phone.br`), several Go helpers in `normalize/` still embed Portuguese/Brazilian heuristics at the package root—specifically Portuguese absence markers (`absent.go`: `"nao"`, `"nenhum"`, `"sem"`), civil status labels (`status.go`: `"solteiro"`, `"casado"`, `"Outro"`), Brazilian corporate suffixes/keywords (`person.go`: `companyWords` like `"ltda"`, `"eireli"`, `"mei"`), and Brazilian street prefixes/states (`address.go`: `"Rua"`, `"Avenida"`, `FormatUF`, `FormatCEP`, `defaultCityCatalog`).
   - Before `1.0`, universal Unicode/text/email/catalog primitives (`DisplayText`, `SearchText`, `Digits`, `Alphanumeric`, `CanonicalEmail`, `NewCityCatalog`) will be cleanly separated from explicit `normalize/br` jurisdictional formatters and locale vocabularies.
2. **Multi-Provider Live Adapter Validation & Streaming Wire Contract:**
   - Complete live integration validation across cloud BYOK providers and local runtimes using `assistant/adaptertest`.
   - Once multiple concrete adapters have exercised streaming text deltas, partial tool-call argument assembly, pause/resume turns, refusal frames, and usage timing, promote `streaming` from a capability flag to a normative event/delta specification.
3. **Deterministic `matching` and `solver` Packages:**
   - Introduce infrastructure-independent entity resolution (`matching`: weighted evidence comparison, support/contradiction scoring) and bounded constraint satisfaction (`solver`: candidate-to-slot assignment with deterministic pruning and execution budgets).
4. **First Non-Go Implementation (TypeScript Conformance Port):**
   - Implement the first non-Go port in TypeScript/Node.js consuming the exact `conformance/v0.4/*.json` vector suites and `schemas/*.schema.json` definitions to prove cross-language portability end-to-end.
