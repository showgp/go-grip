# 03 — Migrate existing Open Target entry points

**What to build:** Route drag-and-drop, Recent Target activation, Preview Session revisit, and the existing external-open provider through the application coordinator so every entry point follows one lifecycle.

**Blocked by:** 02 — Add the application coordinator through Open.

**Status:** ready-for-agent

- [ ] A supported drag-and-drop request produces the same observable result as Open.
- [ ] Activating a Recent Target and revisiting a Preview Session use the same coordinator request path.
- [ ] The existing external-open provider delegates to the coordinator instead of implementing its own launch sequence.
- [ ] Unsupported entry input produces one consistent error outcome.
- [ ] Duplicate opening logic and independently owned session/history state are removed without changing supported behavior.
- [ ] Contract tests prove that every adapter produces the same Open Target request semantics.
