# Fingerprint semantics

Core fingerprints provide stable, non-secret identifiers for exact bytes and explicitly framed semantic parts.

The initial contract uses SHA-256 because it is widely available across supported implementation languages and requires no provider or infrastructure dependency.

## Digest representation

A fingerprint digest is 32 bytes produced by SHA-256.

Its canonical textual form is exactly 64 lowercase hexadecimal ASCII characters. Parsers may accept uppercase hexadecimal input but must emit lowercase canonical text.

Operation keys:

- `fingerprint.sha256.text` — SHA-256 of the exact UTF-8 bytes of a valid Unicode string;
- `fingerprint.sha256.framed_text` — SHA-256 of the Core FingerPrint frame for a namespace plus ordered valid UTF-8 string parts;
- `fingerprint.digest.parse` — validate a hexadecimal digest and return canonical lowercase text.

## Exact-input rule

Fingerprinting does not perform Unicode normalization, whitespace normalization, case folding, JSON normalization, locale conversion, or application-specific canonicalization.

String operations require input that can be represented as valid UTF-8 without replacement or repair. A runtime that can hold malformed string/code-unit sequences must reject them with `invalid_utf8`. Arbitrary binary values use the byte-oriented fingerprint operation/API instead.

If two representations should be semantically equivalent, the consumer must first canonicalize them using an appropriate Core or application contract and then fingerprint the canonical representation.

For example, `"João"` and a canonically decomposed Unicode spelling are different byte sequences unless they are normalized before fingerprinting.

## Core FingerPrint frame v1

Framed fingerprints provide domain separation and unambiguous part boundaries.

The byte sequence hashed by `fingerprint.sha256.framed_text` is:

```text
magic             4 bytes   43 46 50 01  (ASCII "CFP" + frame version 1)
namespace_length  2 bytes   unsigned big-endian
namespace         N bytes   ASCII namespace
part_count        4 bytes   unsigned big-endian
for each part, in order:
  part_length     8 bytes   unsigned big-endian
  part            N bytes   exact payload bytes
```

The namespace must:

- contain 1 through 128 ASCII bytes;
- start with `a` through `z`;
- contain only lowercase ASCII letters, digits, `.`, `_`, `:`, `/`, or `-` after the first byte.

Examples of valid namespaces:

```text
query.plan/v1
task.spec/v1
assistant.tool/v1
cache.entity/v1
```

Part order is significant. Empty parts are valid. A frame with zero parts is valid. The frame format ensures that `("ab", "c")` and `("a", "bc")` are different byte sequences before hashing.

### Validation order

Portable implementations validate framed inputs in this order:

1. namespace syntax and length;
2. representable part count;
3. UTF-8 validity for text-oriented parts;
4. frame encoding and hashing.

When an input violates more than one condition, the first applicable error in this order is the observable result. Byte-oriented parts do not perform UTF-8 validation.

## Domain separation

Consumers should prefer framed fingerprints for semantic identities shared across subsystems.

The same parts under different namespaces intentionally produce different digests. Namespace versions should change when the consumer changes the canonical meaning or ordering of the supplied parts.

Raw `fingerprint.sha256.text` remains useful when the exact string identity is itself the contract. Arbitrary bytes use the byte-oriented implementation API and the same SHA-256 digest representation.

## Security and privacy

A fingerprint is not authentication, encryption, anonymization, or a password hash.

SHA-256 fingerprints of low-entropy or sensitive values can be guessed by enumeration. Do not treat a CPF, email address, phone number, token, or other sensitive input as protected merely because it was hashed.

Authentication/integrity with a secret requires a separate keyed construction such as HMAC and is outside this initial fingerprint contract.

## Future canonical structured values

The initial contract deliberately does not define canonical JSON. JSON canonicalization requires precise rules for duplicate object keys, property ordering, Unicode escaping, and number serialization across languages.

A future structured canonicalization contract may adopt an established cross-language standard or define a separately versioned representation. It must not silently change Core FingerPrint frame v1.
