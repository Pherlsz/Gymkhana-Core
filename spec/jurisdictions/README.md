# Jurisdiction modules

Jurisdiction-specific rules are isolated from global Core primitives.

A jurisdiction module may define identifiers, postal rules, administrative subdivisions, or other legally/locality-dependent semantics. No jurisdiction is the default for Core as a whole.

Examples:

```text
jurisdictions/
├── br/
├── us/
├── pt/
├── gb/
├── ca/
└── ...
```

Country, jurisdiction, language, locale, currency, and timezone are distinct concepts. Implementations must not derive one from another unless an API explicitly documents that mapping.

Application-specific aliases and dirty-data heuristics do not become jurisdictional standards merely because they are useful to one consumer. Such catalogs should remain configurable or consumer-owned unless they represent a stable reusable dataset.
