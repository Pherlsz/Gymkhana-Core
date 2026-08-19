# Contact specification

Contact primitives model communication identifiers independently from application persistence.

## Email

`contact.email.canonicalize` produces the Core canonical email representation. The current semantics preserve the local part and normalize the domain according to the versioned implementation contract.

## Telephone

Telephone semantics distinguish a global representation from jurisdiction-specific validation rules.

E.164 is the canonical international representation for validated telephone numbers when the operation supports it. Country-specific acceptance and numbering-plan validation must use explicit jurisdiction context, for example `contact.phone.br.canonicalize`.

Core must not guess a country code from an ambiguous local number without explicit context.
