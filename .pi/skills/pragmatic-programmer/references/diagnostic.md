# Pragmatic Diagnostic

Use for Assess mode. Score only declared scope. Never treat absent evidence as failure.

## Evidence states

- Enforced - repository rule plus code, test, automation, or repeated practice proves row. Value 1.
- Partial - intent exists, but enforcement or coverage has material gap. Value 0.5.
- Gap - applicable row lacks reliable implementation. Value 0.
- N/A - row does not apply to assessed scope. Exclude from score.
- Unverified - evidence was unavailable or investigation was too shallow. Exclude from score and reduce confidence.

Calculate:

```text
score = 10 * (enforced + 0.5 * partial) / applicable
confidence = evidenced applicable rows / all applicable-or-unverified rows
```

Round score to one decimal. Report confidence as percentage. Never claim 10/10 while any row is Partial, Gap, or Unverified.

## Diagnostic rows

### Knowledge authority

Question: Is each business rule, contract, schema fact, and configuration decision represented by one authoritative source?

Evidence: duplicated validators, copied constants, schema/type generation, comments restating behavior, glossary drift.

Gap action: name authoritative source, route consumers through it, and remove competing representations.

### Orthogonal change

Question: Can persistence, vendor, UI, or infrastructure choices change without forcing unrelated business-logic edits?

Evidence: forbidden-import tests, adapters, dependency direction, global state, change history.

Gap action: localize coupling behind smallest stable interface and add boundary regression test.

### Working tracer path

Question: Does one thin production path connect user-visible entry point through real boundaries and verification?

Evidence: vertical-slice tests, deployable walking skeleton, end-to-end acceptance path.

Gap action: build one narrow end-to-end slice before expanding horizontal layers.

### Explicit contracts

Question: Do changed public seams state caller obligations, guarantees, failures, and state effects, with tests proving relevant claims?

Evidence: types, validation, constraints, typed errors, doc comments, contract-touch RED tests.

Gap action: define contract at strongest enforcement site and add smallest proof.

### Repaired windows

Question: Are known defects, hacks, stale warnings, and temporary exceptions fixed or boarded up with owner and cleanup path?

Evidence: TODOs, ignored tests, disabled checks, stale incidents, principle warnings, linked issues.

Gap action: repair now or record explicit scope, owner, deadline, and regression net.

### Reversible choices

Question: Can risky deployment, vendor, data, and architecture decisions be rolled back or replaced at proportionate cost?

Evidence: feature flags, migration classification, adapters, rollback procedure, deployment strategy.

Gap action: isolate irreversible edge, add rollback path, or time-box experiment before commitment.

### Honest estimates

Question: Do estimates use decomposed ranges, assumptions, confidence, and later calibration rather than single-point promises?

Evidence: plans, issue estimates, spike results, estimate-versus-actual records.

Gap action: state optimistic, likely, and pessimistic range with assumptions and confidence.

### Active knowledge portfolio

Question: Does project convert incidents and new knowledge into durable docs, tests, automation, or scheduled learning?

Evidence: learning index, retrospectives, laws earned from bugs, changelog, training cadence.

Gap action: record lesson in smallest durable surface and schedule next learning investment.

## Output

Return:

```text
Pragmatic score: N.N/10
Confidence: NN%
Scope: <assessed boundary>

Enforced: <rows>
Partial: <rows with file:line evidence and fix>
Gaps: <rows with file:line evidence and fix>
Unverified: <rows and missing evidence>
N/A: <rows and reason>

Path to 10/10:
- <smallest ordered actions>
```

Order actions by change leverage, then reversibility, then effort. Do not recommend cleanup sprint when one localized repair suffices.
