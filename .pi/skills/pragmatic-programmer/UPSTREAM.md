# Upstream Provenance

This framework skill adapts design ideas from locally reviewed upstream skill:

- Name: `pragmatic-programmer`
- Author: wondelai
- Version reviewed: 1.4.0
- License: MIT
- Upstream body checksum: `3f9c43eafc60bd142c6d6f4d5f254d295060d4b1b89a79878b2da3273fdeb22f`

## Retained ideas

- Broad craftsmanship trigger vocabulary.
- Consultative use across design, architecture, engineering culture, build-vs-buy, reversibility, and estimation.
- Evidence-oriented quick diagnostic.
- Normalized 10-point score.
- Progressive disclosure into focused references.
- Explicit routing to neighboring specialist skills.

## Framework adaptations

- Existing `core/pragmatic-principles.md` remains authoritative instead of vendoring duplicate principle prose.
- Diagnostic uses Enforced, Partial, Gap, N/A, and Unverified evidence handling.
- Scores exclude non-applicable and unverified rows; confidence exposes evidence coverage.
- Framework's 18-principle spine replaces upstream's seven-group knowledge body.
- Warn-and-cite protocol, contract touch, and specialist routing are integrated.
- Missing sibling references to unavailable skills were removed.

## Update procedure

When reviewing newer upstream release:

- Compare trigger description, diagnostic rows, and decision workflows.
- Port useful mechanics, not duplicated explanatory prose.
- Update `metadata.based_on`, this file, and `LICENSE.upstream` when required.
- Recompute package checksum in `skills/SKILLS.md`.
- Run manifest, bootstrap, idempotency, and uninit checks.

Do not overwrite framework operational guidance with upstream generic examples.
