# CI policy

Use local verification during development:

```bash
scripts/verify-local.sh quick
scripts/verify-local.sh full
scripts/verify-local.sh security
```

Keep pull requests as drafts while pushing iterative changes. The automatic CI runs only when a pull request is opened as ready, reopened, or moved from draft to ready for review. It does not run after every synchronization push.

After changing a ready pull request, return it to draft, finish and validate locally, then mark it ready once to request a final hosted gate.

Security scans run monthly and remain available through manual dispatch. Release automation runs only when `VERSION` changes on `main` or when manually requested.
