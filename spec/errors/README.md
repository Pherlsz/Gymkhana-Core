# Error semantics

Core errors are stable, structured, and non-localized. Implementations may wrap them in idiomatic error types, exceptions, result values, or discriminated unions, but must preserve the semantic code required by the specification.

Common validation codes currently include:

- `empty`
- `invalid_format`
- `invalid_length`
- `invalid_checksum`
- `unknown_kind`

Domain-specific specifications may define additional stable codes when a generic code would lose useful semantics. Current examples include temporal codes such as `invalid_year`, `invalid_month`, `invalid_day`, `zero_value`, and `out_of_range`; normalization codes such as `invalid_area_code`, `invalid_number`, `missing_state`, `invalid_state`, and `ambiguous`; fingerprint codes such as `invalid_namespace`, `invalid_utf8`, and `too_many_parts`; and Assistant codes such as `invalid_role`, `invalid_content`, `invalid_tool_name`, `invalid_json`, `invalid_finish_reason`, `invalid_usage`, `invalid_capability`, `duplicate_capability`, `invalid_skill`, and `duplicate_skill`.

A code's meaning belongs to the specification area that defines it. New implementations must not compare localized message text to determine error behavior.

Errors must not embed sensitive original input values. In particular, Assistant errors must not include prompt text, tool arguments/results, media URIs, or skill payload data. Human-facing localization, HTTP status mapping, logging policy, and product copy belong to consuming applications.
