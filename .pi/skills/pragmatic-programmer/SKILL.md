---
name: pragmatic-programmer
description: Applies pragmatic software-craft principles through evidence-backed consultation, codebase assessment, and engineering decisions. Use for best practices, pragmatic approaches, DRY, orthogonality, broken windows, tracer bullets, design by contract, technical debt, reversibility, build-vs-buy choices, estimation, or learning portfolios.
license: MIT; see LICENSE.upstream
metadata:
  author: agent-stack contributors
  version: "2.0.0"
  based_on: "wondelai/pragmatic-programmer 1.4.0"
---

# Pragmatic Programmer

Use this skill as executable interface to agent-stack's pragmatic guidance. Do not duplicate principle text already maintained in framework docs.

## Resolve framework docs

In bootstrapped repos, framework docs live under `docs/agents/`. In agent-stack source repo, same docs live under `core/`.

Resolve these before working:

- `pragmatic-principles.md` - authoritative principle spine, operational forms, and warn-and-cite protocol
- `complexity.md` - net-complexity-gain, deep modules, and build-vs-buy constraints
- `feature-protocol.md` - tracer-slice and principle-warning requirements
- `tdd.md` - contract-touch proof

Read only relevant headings. Do not load whole principle spine or 100-tip index when targeted search can find needed section.

## Choose mode

Infer mode from request. Ask only when intent remains ambiguous.

- Consult - explain applicable principle and give context-specific advice.
- Assess - score codebase or proposed change using `references/diagnostic.md`.
- Decide - compare build-vs-buy, reversible-vs-irreversible, estimation, or architecture choices.

Read `references/principle-crosswalk.md` when mapping request to framework sections or neighboring skills.

## Workflow

- Identify decision, change, or assessment boundary.
- Name applicable principle before recommending action.
- Inspect repository evidence. Use `file:line` references for codebase claims.
- Separate authoritative knowledge from coincidental code similarity when applying DRY.
- Prefer reversible experiments for uncertain decisions. Use thin tracer slice for uncertain integration paths.
- Apply YAGNI to speculative behavior, not automatically to interface depth.
- If proposed work violates framework principle, emit principle warning with principle, operational form, rationale, and cleanup plan.
- Route specialist work rather than reproducing specialist procedures.

## Specialist routing

- Building vertical slice - use `tracer-bullets`.
- Designing or deepening module seam - use `codebase-design` or `deep-module-engineer`.
- Clarifying overloaded requirements - use `lock-requirements`.
- Planning refactor - use `request-refactor-plan` when available.
- Debugging failure - use repository's diagnostic skill.

Keep this skill active for cross-cutting trade-offs and final pragmatic assessment.

## Response contract

For Consult or Decide, return:

- Principle - named framework section or mapped upstream concept
- Evidence - repository references, constraints, and unknowns
- Trade-off - benefit, cost, reversibility, and blast radius
- Action - smallest concrete next step
- Warning - required warn-and-cite block, or `None`

For Assess, follow `references/diagnostic.md`. Always state score, confidence, failing rows, and actions needed to reach 10/10.

## Provenance

Workflow adapts diagnostic and trigger ideas from wondelai's `pragmatic-programmer` v1.4.0. Framework docs remain authoritative. See `UPSTREAM.md` and `LICENSE.upstream`.
