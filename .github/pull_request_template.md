## Pull Request Review Checklist

As required by Ghost Ops standards (and specified in `ai-skills.json`), all AI-synthesized code or system changes must undergo mandatory human review.

Please confirm the following:
- [ ] This PR has been manually reviewed for AI-generated code.
- [ ] Code passes all linting, test, and security gates (e.g. `make audit`, `gosec`, `govulncheck`).
- [ ] Architecture decision records (ADRs) have been updated if applicable.
- [ ] No secrets or sensitive data are exposed.
