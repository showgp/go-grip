# 09 — Add directory Finder Command and input hardening

**What to build:** Extend the Finder Command to one directory while rejecting ambiguous or unsupported selections and preserving the same recursive preview behavior across locations.

**Blocked by:** 08 — Add the Markdown Finder Command Service.

**Status:** ready-for-agent

- [ ] Finder offers the command for one directory and does not offer it for unrelated file types.
- [ ] A directory invocation requests recursive document discovery from the existing backend.
- [ ] Exactly one supported Open Target is required; empty and multiple selections fail clearly without partial processing.
- [ ] Markdown validation remains case-insensitive at runtime.
- [ ] Accessible targets outside the home directory, including temporary external-volume fixtures, are not rejected by an artificial managed-directory boundary.
- [ ] Unsupported, missing, and inaccessible targets do not launch a preview process.
- [ ] File and directory Service paths share the same coordinator behavior tests.
