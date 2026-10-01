---
name: setup-openspec-tickets
description: "Configure a repo for the OpenSpec ticket flow after installing skills: confirm config rules and agent instructions before writing them. Run once per repo."
disable-model-invocation: true
---

# Setup OpenSpec Tickets Flow

Configure the repo after `openspec init` and skill installation. Explore, present
findings, confirm the proposed changes with the user, then write.

## Process

### 1. Explore

Look at the current repo to understand its starting state. Read whatever exists;
don't assume:

- `openspec/config.yaml`: present? What do its `context`, `rules`, and
  `operations` sections already say? Never overwrite user content there.
- Root agent instructions: `AGENTS.md`, else `CLAUDE.md`. If neither exists,
  propose creating `AGENTS.md` with the approved section. If a flow section
  exists, edit it rather than adding another.
- Check `openspec --version` and `openspec list --json`; inspect the CLI's
  status for any selected active change, if one exists.
- Upstream skills present: `grilling`, `tdd`, `to-tickets` under
  `.agents/skills/` (or the project's skill directory). Missing ones mean the
  flow's step 2 (grilling before artifacts) and step 4 (scoped TDD) have no
  backing skill; report, and offer to install with
  `npx skills@latest add mattpocock/skills -s <skill> -y`.
- Which quality rules the repo already defines (linters, size limits, test
  policies). The flow skills read "applicable repository policies" at runtime;
  record where they live so reviewers can find them, but do not invent new
  quality rules here.

### 2. Present findings and confirm

Summarise: openspec CLI state, config.yaml current contents, agent-instruction
file choice (or proposed new `AGENTS.md`), present/missing upstream skills,
and existing repo policy locations. Write nothing yet. Take the sections in
order; one answer per section, recommended answer first.

**Section A — OpenSpec config rules.** Recommended: add these rules under
`rules.proposal` and `rules.tasks` in `openspec/config.yaml`, respectively:

```yaml
rules:
  proposal:
    - Include an explicit Non-goals section listing behaviors considered and excluded.
  tasks:
    - Put each checkbox, complete task description, and verification criterion on one physical line; openspec instructions apply --json exposes only that line as the task description.
```

If `rules` or `rules.tasks` already exist, merge rather than replace. Skip any
rule whose intent is already present. Explain in one line why the tasks rule
matters (apply JSON truncates to one line); if the user declines it, the
`tasks-to-tickets` skill will surface split lines as a repair question every
time, so get a decision either way.

**Section B — Agent-instruction section.** Add an `## OpenSpec ticket flow`
section to `AGENTS.md` (or `CLAUDE.md` if that alone exists). If neither exists,
ask to create `AGENTS.md` with this section. Draft from this template, trimming
to what the repo lacks:

```markdown
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
```

Preserve the repo's own heading style. If the file already says any of this,
edit in place rather than duplicating.

**Section C — Upstream skill gaps.** Only if skills from the exploration are
missing: offer to run the installer commands for the missing upstream skills
and stop for the user's confirmation. Never fetch a different skill as a
substitute.

### 3. Write and verify

After the user decides both sections (and whether to install any missing skills), apply only the approved changes:

- Edit `openspec/config.yaml` only if Section A was approved (merge, never clobber).
- Add or update the agent-instruction section only if Section B was approved.
- Re-read changed files and confirm the approved rules and section are present, without duplicate headings or discarded existing content. If neither section was approved, leave the repo unchanged. Do not create a sample change; this skill only writes configuration.

### 4. Done

Report exactly what was written, and the first command to run next:
discuss a real request with `grilling`, then create one OpenSpec change and
walk the approval chain. Mention that per-project quality rules (linters, size
limits) stay the repo's own responsibility and are read by the flow skills
from this file.
