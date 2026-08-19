# Postal specification

Postal identifiers are jurisdiction-sensitive. There is no single global postal-code grammar.

Operations must therefore carry explicit jurisdiction context, for example:

- `postal.br.cep.canonicalize`;
- future `postal.us.zip.canonicalize`;
- future `postal.gb.postcode.canonicalize`.

Canonicalization preserves semantically significant leading zeroes and does not infer locale, language, currency, or timezone from the postal jurisdiction.

Presentation formatting is separate from canonical storage representation.
