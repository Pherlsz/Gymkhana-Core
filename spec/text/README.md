# Text specification

Text primitives are Unicode-aware and independent of any country or product.

## `text.display`

Produces a stable human-display representation by normalizing Unicode according to the implementation contract, trimming surrounding whitespace, and collapsing internal whitespace sequences without inventing semantic content.

## `text.search`

Produces a deterministic search representation suitable for equality/search preprocessing. It may fold case, accents, and punctuation according to versioned Core semantics.

Search representations are not display values and must not silently replace canonical persisted values unless a consumer explicitly chooses that behavior.

## Compatibility

Any semantic change that would make an existing input produce a different canonical/search output requires a specification version change and updated conformance vectors.
