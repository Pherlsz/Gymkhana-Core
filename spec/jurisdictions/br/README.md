# Brazil jurisdiction module

Brazil-specific behavior is a jurisdictional extension of the language-neutral Core specification. It must not define defaults for unrelated countries or services.

Initial implemented semantics include:

- `identity.br.cpf.canonicalize` — accepts the supported CPF presentation separators, validates both check digits, and returns exactly 11 ASCII digits;
- `identity.br.cnpj.canonicalize` — validates numeric and Receita Federal alphanumeric CNPJ forms and returns the 14-character uppercase canonical value;
- `identity.br.document.canonicalize` — receives an explicit document `kind` plus value and returns the validated canonical representation for a supported Brazilian document kind;
- `identity.br.document.format` — receives an explicit document `kind` plus value and returns the canonical display representation after validation;
- `postal.br.cep.canonicalize` — returns exactly 8 ASCII digits and preserves leading zeroes;
- `contact.phone.br.canonicalize` — returns the validated Brazilian telephone number in E.164 form.

The current Go document implementation includes CPF, CNPJ, RG/identity, voter ID, CNH, CTPS, PIS, passport, SUS card, selected professional registrations, and selected loose document identifiers. Not every kind has the same validation strength: weak identifiers are not silently inferred merely because their shape can accept a value.

State-dependent professional identifiers require explicit valid UF context when the canonical contract requires it. Canonicalization and display formatting remain separate operations.

Administrative-subdivision and geography helpers may be provided when they represent reusable Brazilian semantics. Consumer-specific city aliases, import cleanup rules, and product vocabulary are not automatically part of this jurisdiction module.

Future jurisdiction modules must follow the same separation rather than copying Brazilian assumptions into generic APIs.
