# Identity and identifier specification

Core treats identifiers as typed semantic values rather than assuming every identifier is a generic document number.

Conceptual categories may include:

- personal identifiers;
- tax identifiers;
- organization identifiers;
- government documents;
- professional registrations;
- financial identifiers;
- travel documents.

These categories are semantic metadata, not an object-oriented inheritance requirement.

## Jurisdiction

Jurisdiction-sensitive identifiers must include the jurisdiction in their specification key, for example:

- `identity.br.cpf`
- `identity.br.cnpj`
- future `identity.us.*`, `identity.pt.*`, and other jurisdictional identifiers.

Global identifiers that are defined by an international standard rather than one jurisdiction must use a global/standard namespace instead of pretending to belong to a country.

## Explicit context

Core must not silently infer a jurisdiction when multiple interpretations are possible. Generic detection APIs, when introduced, must receive an allowlist/context and return candidates rather than inventing authority.

## Canonicalization and display

Canonical values are stable machine representations. Display formatting and masking are separate operations. Consumers must not infer storage semantics from a presentation formatter.
