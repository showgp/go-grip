# AGENTS.md

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

> For the Project Architecture, please see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

## Agent skills

### Issue tracker

Issues are tracked as local Markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Triage uses the five default canonical labels. See `docs/agents/triage-labels.md`.

### Domain docs

Domain documentation uses the single-context layout. See `docs/agents/domain.md`.

## OpenSpec ticket flow

- For a new change, run a grilling session with the user to confirm the complete
  requirements and non-goals. Do not create planning artifacts or code before
  the relevant approval.
- OpenSpec is the sole source of behavioral requirements. Read the selected
  change's current CLI status and instructions, and the artifact paths they
  return; do not substitute agent instructions, README, or tests for its specs.
- Draft the proposal, spec, design, and tasks separately. Obtain approval before
  writing each artifact, then stop for review. Do not use a one-shot workflow
  that generates all artifacts without these checkpoints.
- YAGNI is a scope gate: each new behavioral guarantee, defensive abstraction,
  and permanent test must trace to an approved OpenSpec requirement or an
  explicit repository quality rule. A review finding or a constructed edge case
  is evidence, not a blocking criterion; discuss credible trigger, user impact,
  simpler alternative, and cost with the user before adding handling or tests.
- In `tasks.md`, put each checkbox, complete task description, and verification
  criterion on one physical line; after writing or editing tasks, inspect
  `openspec instructions apply --json` output for counts and descriptions.
- Publish tickets one per review cycle with `tasks-to-tickets`; implement one
  approved ticket at a time with `implement-openspec-ticket` (scoped TDD at
  seams agreed with the user; real checks; observed CLI smoke), then stop and
  wait for permission. Reporting progress is not permission to continue.
- `review-openspec-ticket` performs the independent Standards and Spec review
  before a ticket can close. A test's passing or a review's zero findings never
  promote unapproved behavior into the spec.
- Syncing durable specs (`openspec/specs/`) and archiving a change each require
  separate explicit approval; do not commit on the user's behalf without
  approval.
