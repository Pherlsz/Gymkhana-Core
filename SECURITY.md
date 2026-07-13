# Security Policy

Gymkhana Core is a private project.

## Reporting a vulnerability

Report vulnerabilities privately to the repository owner through an appropriate private GitHub channel. Do not open a public issue containing exploit details, credentials, tokens, personal data, or other sensitive information.

Include:

- affected code or version;
- reproduction steps;
- security impact;
- suggested mitigation when available.

## Supported versions

During initial development, `main` is the supported development line. After releases begin, the latest tagged version and `main` are reviewed for security fixes unless stated otherwise in a release notice.

## Scope

This repository must remain independent from credentials, databases, HTTP sessions, providers, and product data. A change that introduces those concerns is also an architectural security issue.
