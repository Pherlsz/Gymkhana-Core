# Error semantics

Core errors are stable, structured, and non-localized. Implementations may wrap them in idiomatic error types, exceptions, results, or discriminated unions, but must preserve the semantic code.

Initial validation codes:

- `empty`
- `invalid_format`
- `invalid_length`
- `invalid_checksum`
- `unknown_kind`

Errors must not embed sensitive original input values. Human-facing localization belongs to consuming applications.
