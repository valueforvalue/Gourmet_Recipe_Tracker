# Principle Crosswalk

This file maps upstream pragmatic-programmer concepts to agent-stack authority. Use it to locate guidance, not as competing explanation.

## Authority boundary

- `pragmatic-principles.md` owns principle definitions, operational forms, accepted exceptions, and 100-tip state.
- `feature-protocol.md` owns vertical-slice and principle-warning mechanics.
- `tdd.md` owns contract-touch proof.
- `complexity.md` owns net-complexity-gain and deep-module decisions.
- This skill owns mode selection, diagnostic scoring, and response shape.

When docs disagree, current bootstrapped framework docs win. Record drift in this crosswalk rather than copying corrected prose into SKILL.md.

## Upstream concepts

### DRY

- Primary: `pragmatic-principles.md` §1.1.
- Operational: `glossary-discipline.md`, `complexity.md` rule of three, schema or API generators.
- Guard: deduplicate knowledge, not coincidentally similar syntax.

### Orthogonality

- Primary: `pragmatic-principles.md` §1.2 and §1.7.
- Operational: `complexity.md` boundary enforcement and per-stack architecture tests.
- Specialist: `codebase-design`, `deep-module-engineer`, or `design-an-interface`.

### Reversibility

- Primary: `pragmatic-principles.md` §1.3.
- Operational: `commit-and-branch.md`, migration history, feature flags, adapters, rollback procedures.
- Decision test: optimize flexibility only where concrete uncertainty or replacement cost justifies it.

### Tracer bullets and prototypes

- Primary: `pragmatic-principles.md` §1.4 and §1.10.
- Operational: `feature-protocol.md` slice-size and end-to-end rules.
- Specialist: `tracer-bullets` for implementation.
- Guard: tracer code is retained production code; prototype is explicitly disposable.

### Design by contract and assertive programming

- Primary: `pragmatic-principles.md` §1.5 and §1.6.
- Operational: `tdd.md` Contract touch and `feature-protocol.md` Contract touch.
- Guard: enforce at types, validation, constraints, typed errors, and tests before generic runtime-contract helpers.

### Broken windows

- Primary: `pragmatic-principles.md` §6 Tip #5 and Tip #6.
- Operational: `laws.md`, `bug-patterns.md`, changelog, tracked principle warnings.
- Guard: temporary divergence needs rationale, owner, cleanup path, and regression net.

### Estimation and knowledge portfolio

- Primary: `pragmatic-principles.md` §6 Tip #9, Tip #22, Tip #23, and §1.12.
- Operational: project learning index when present, plans, estimate
  calibration, incident-derived laws, retrospectives, and changelog.
- Guard: use ranges and confidence; time-box exploration when uncertainty dominates estimate.

## Framework extensions

Agent-stack expands upstream's seven-group model with these explicit surfaces:

- Decoupling and Law of Demeter - §1.7.
- Metaprogramming and metadata - §1.8.
- Temporal coupling and concurrency - §1.9.
- Program deliberately - §1.11.
- Algorithm speed - §1.12.
- Refactoring - §1.13.
- Testability - §1.14.
- Ubiquitous automation - §1.15.
- Writing and communication - §1.16.
- Pragmatic projects - §1.17.
- Before project work - §1.18.

Use these sections directly when request is more specific than upstream grouping.
