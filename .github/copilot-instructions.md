# Copilot instructions — Morris Family Recipe Tracker

This file governs **how Copilot talks to the maintainer**. It does not govern
engineering behaviour — that lives in [`AGENTS.md`](../AGENTS.md),
[`CONTEXT.md`](../CONTEXT.md), and [`docs/agents/`](../docs/agents/), which are
the authority for code changes, commit shape, testing, and domain vocabulary.
When a communication rule here and an engineering rule there appear to
conflict, the engineering rule wins and this file's channel preference is
simply dropped.

## Communication channel

1. **Prefer `get_multiline_input` when it is available and the session is
   interactive.** If the tool exists in the current session, use it for
   questions, confirmations, and clarifications that genuinely need the
   maintainer's input.
2. **Degrade gracefully.** If `get_multiline_input` is not exposed (headless
   run, CI, Copilot coding agent, a different host, or the MCP server is
   down): do not stall, do not retry the missing tool, and do not claim to
   have asked a question you could not ask. Answer inline in the normal
   response channel and state plainly that the interactive channel was
   unavailable.
3. **Never block on input you can resolve yourself.** If the answer is
   already in the repo — README, `CONTEXT.md`, `AGENTS.md`, `config.json.example`,
   git history — read it instead of asking. Only ask when the decision is
   genuinely the maintainer's (scope, product intent, irreversible ops).

## Session endings

4. **End every response with a concrete next step or a stated stop.** One
   closing question is fine when it moves work forward. Do not append a
   ritual "do you need any further assistance?" footer to every reply — it
   trains the maintainer to ignore it and adds nothing on a reply that already
   ends in a decision.
5. **`end` / `stop` means stop.** Those two words end the session. So does an
   unanswered question after one re-prompt, a dismissed dialog, or an
   interrupted run: report what was completed, what was not, and halt.
   **Never loop a prompt.** A cancelled or dismissed dialog is a stop signal,
   not a prompt to re-ask. Maximum one re-prompt per question, ever.

## Scope

6. These rules apply only to conversation mechanics. They never override:
   - the vault-before-database law and the three-way `Recipe` contract in
     [`CONTEXT.md`](../CONTEXT.md),
   - `web/js/elm.js` being a rebuilt-never-edited artifact,
   - Conventional Commits and the `[Unreleased]` changelog discipline in
     [`docs/agents/commit-and-branch.md`](../docs/agents/commit-and-branch.md),
   - any instruction the maintainer gives in-session, which outranks this file.
