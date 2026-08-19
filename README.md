# Gymkhana Core

Private, language-neutral foundation for reusable contracts, deterministic algorithms, validation, normalization, AI primitives, and other infrastructure-independent capabilities shared across applications and services.

Go is currently the first implementation. It is not the architectural source of truth.

## Foundation principles

Gymkhana Core is intentionally:

- **language-neutral** — semantics are designed to be implementable in Go, TypeScript/Node.js, Java, .NET/C#, Python, Rust, PHP, and future languages;
- **product-neutral** — Gymkhana Database and other applications are consumers, not owners of Core semantics or roadmap;
- **provider-neutral** — model/provider SDKs, cloud vendors, credentials, pricing, transport, and deployment remain outside Core;
- **jurisdiction-neutral** — Brazil-specific behavior is one jurisdictional module among international standards and future country modules;
- **infrastructure-independent** — persistence, HTTP, queues, UI, workers, and deployment are consumer concerns;
- **deterministic where possible** — equivalent semantic inputs produce equivalent canonical outputs across conforming implementations.

The language-neutral specification lives in [`spec/`](spec/). Shared behavior vectors live in [`conformance/`](conformance/), and portable serialized contracts live in [`schemas/`](schemas/).

## Repository model

```text
Gymkhana-Core/
├── spec/          # normative language-neutral semantics
├── conformance/   # shared machine-readable behavior vectors
├── schemas/       # portable serialized contracts
├── normalize/     # current Go implementation
├── civiltime/     # current Go implementation
├── fingerprint/   # current Go implementation
└── ...            # future language implementations/packages as justified
```

A language implementation may expose idiomatic APIs, but it must preserve the observable behavior of the specification version it claims to support.

## Current Go packages

### `normalize`

Deterministic normalization for:

- display and search text;
- Unicode letters and digits;
- person names;
- street, house number, block (`bl. A`), apartment (`apt. 202`), CEP, and UF;
- city, neighborhood, country, and nationality;
- gender, marital status, blood type, and club membership type;
- team, sector, football club, health plan, card, collection, animal, and vehicle;
- CPF validation and canonicalization;
- CNPJ validation and canonicalization, including the alphanumeric Receita Federal form;
- catalog document identifiers, formatting, validation, and strong kind inference;
- email canonicalization;
- Brazilian phone validation and E.164 canonicalization.

Some current Go formatters contain consumer-proven cleanup behavior. New architecture must distinguish universal primitives, jurisdictional rules, and application-specific data/aliases instead of treating all existing normalization as a global standard.

Normalization never logs or embeds the original value in validation errors.

### `civiltime`

Timezone-free value types for:

- strict `CivilDate` parsing and serialization;
- strict `YearMonth` parsing and serialization;
- comparison, leap-year rules, month arithmetic, and month boundaries.

Zero values represent absence and serialize as empty text or JSON `null`.

### `fingerprint`

Portable deterministic fingerprint primitives for:

- exact SHA-256 fingerprints of bytes and UTF-8 strings;
- canonical 64-character lowercase hexadecimal digest representation;
- strict digest parsing with stable errors;
- versioned, domain-separated framing of ordered byte/string parts;
- explicit part boundaries so concatenation ambiguity cannot change semantic identity.

`fingerprint` does not normalize caller data and deliberately does not define canonical JSON yet. Consumers canonicalize semantic values before hashing when representation equivalence matters.

Fingerprints are not encryption, anonymization, authentication, or password hashing. Hashing low-entropy sensitive identifiers does not make them private.

## Specification and conformance

`spec/VERSION` identifies the language-neutral specification version independently from the Go module release.

Core Spec `0.2` covers:

- global text semantics;
- civil temporal semantics;
- a generic identifier model;
- explicit jurisdiction modules, beginning with Brazil because those capabilities already exist in Go;
- deterministic fingerprint digest/framing semantics;
- an international-standards extension point;
- stable non-localized error semantics;
- machine-readable conformance suites executed by the current Go implementation.

Run the current Go conformance suites with:

```sh
make conformance
```

Future Java, .NET, TypeScript, Python, Rust, PHP, or other implementations must consume the same vectors rather than redefining expected behavior in language-local tests.

## Architectural boundaries

Core may contain coherent reusable packages such as:

- text/normalization primitives;
- civil temporal values;
- typed identifiers and jurisdiction modules;
- international standards;
- canonical serialization and fingerprints;
- provider-neutral Assistant/tool contracts;
- OCR suggestion/evidence contracts;
- deterministic matching primitives;
- generic constraint/composition solving algorithms.

Core must not contain:

- PostgreSQL, migrations, pgx, sqlc, or product persistence;
- HTTP handlers/status codes, sessions, permissions, or resource ownership;
- React or product UI components;
- River/R2/cloud deployment concerns;
- OpenAI, Google, Anthropic, or other provider SDK types;
- credentials, prompts tied to one application, pricing, or model selection;
- product-specific DTOs or complete Profile/document/bill models;
- arbitrary SQL/code execution or autonomous agent loops.

Do not create broad catch-all packages named `utils`, `helpers`, `common`, `shared`, or `core`.

## Jurisdiction and locale rules

Country, jurisdiction, language, locale, currency, and timezone are distinct concepts.

Core must not silently infer one from another when more than one interpretation is possible. Jurisdiction-sensitive operations use explicit keys/context such as `identity.br.cpf` rather than assuming Brazil globally.

Application-specific aliases and dirty-data heuristics should remain configurable or consumer-owned unless they represent a stable reusable contract or dataset.

## Requirements for the current Go implementation

- Go 1.26.5
- GNU Make or a compatible implementation
- GitHub authentication for private module consumption

## Setup

```sh
make setup
make check
```

Useful targeted commands:

```sh
make format
make format-check
make vet
make lint
make test
make conformance
make test-race
make fuzz-smoke
make security
```

Pinned tool binaries are installed into `./bin`, which is ignored by Git.

## Go consumption

Go consumers configure:

```sh
export GOPRIVATE=github.com/Pherlsz/Gymkhana-Core
```

and pin an exact release:

```sh
go get github.com/Pherlsz/Gymkhana-Core@v0.4.0
```

Permanent `replace` directives, copied source, submodules, and branch dependencies are not supported as production dependency mechanisms.

Future language implementations should be published through their normal ecosystem package managers once they exist; they do not need to share the same implementation SemVer as the Core specification.

## Development governance

Read [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/ORCHESTRATION.md](docs/ORCHESTRATION.md) before changing public contracts.

Core owns its own architecture and roadmap. Work may originate from any consumer repository or from an independently approved Core capability. When a consumer drives a change, link that consumer's issue/requirement, but do not make the consumer repository the source of truth for Core semantics.

A new capability belongs in Core when it forms a coherent reusable contract or algorithm with clear semantics, tests/conformance, and infrastructure independence. Proven reuse is strong evidence, but an intentionally designed foundation capability does not need to be duplicated in two applications before it may exist.

## CI and GitHub Actions budget policy

The personal GitHub Pro allowance is shared across private repositories owned by the account. Hosted Actions are final integration and publication gates, not a remote development environment.

Workflow behavior:

- draft PRs do not start runner jobs;
- non-draft synchronization runs formatting, vet, unit tests, conformance, and semantic-version validation through repository checks;
- Staticcheck, race tests, and fuzz smoke run when a PR first becomes reviewable, when a ready PR opens or reopens, or through manual execution;
- meaningful corrections return the PR to draft and trigger one new full run only after local validation;
- Security runs for relevant reviewable PRs, manually, and on the 1st and 15th of each month;
- successful PR quality checks are not repeated after merge;
- Release runs only when `VERSION` changes on `main` or when manually requested;
- Release creates the missing tag and GitHub Release without rerunning the complete quality suite;
- obsolete runs on the same PR are canceled;
- setting repository variable `CI_RUNNER` to a trusted Linux self-hosted runner label moves jobs away from GitHub-hosted minutes.

Mandatory usage rules:

1. Keep one implementation PR per active Core workstream. Helper, diagnostic, formatting, export, validation, and squash PRs are prohibited.
2. Open implementation PRs as drafts and run relevant commands locally before the first ready-for-review transition.
3. Run `make check` locally before a public-contract release or milestone acceptance.
4. Do not use repeated pushes, dummy commits, PR close/reopen cycles, or temporary workflows as a remote test loop.
5. Re-run only the failed job and only when the failure was caused by transient infrastructure.
6. Review account-wide Actions usage before starting large multi-language work. At 70% usage, move heavy gates to local or self-hosted execution; at 85%, reserve hosted runners for final merge gates and releases; at 95%, hosted execution requires owner approval.
7. Any workflow change that increases frequency, job count, timeout, matrix size, artifact retention, or runner cost must explain its monthly impact.
8. Dependabot changes require functional review and compatibility validation before merge.

## Protection for `main`

Use a branch ruleset that:

- blocks branch deletion and force pushes;
- requires changes through pull requests;
- requires review conversations to be resolved;
- requires linear history;
- does not require an independent approval while there is only one maintainer;
- does not require branches to be up to date before merge unless a real integration dependency requires it;
- grants no routine bypass.

Allow squash merge only and automatically delete merged branches.

## Releases

The functional pull request carries its own Go module semantic version and changelog section. Before the final commit, run:

```sh
python scripts/release-version.py write \
  --base-version <version-on-main> \
  --title "<final-pr-title>"
```

CI recalculates the expected version from the final pull request title:

- `fix` increments the patch version;
- `feat` increments the minor version;
- a `!` or `BREAKING CHANGE` increments the minor version before `1.0.0` and the major version afterwards;
- `chore`, `docs`, `test`, and other non-releasable changes keep the current version.

The Core specification version and language implementation versions are separate concepts. While Go is the only implementation, the repository may release them together operationally, but future ports may advance at different implementation versions while declaring the same supported specification version.

After a protected PR merge changes `VERSION`, the Release workflow creates the missing `vX.Y.Z` tag and GitHub Release directly.
