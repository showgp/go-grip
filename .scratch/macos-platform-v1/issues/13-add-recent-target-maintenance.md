# 13 — Add Recent Target maintenance

**What to build:** Let users understand, retry, remove, and clear Recent Targets without accidentally affecting active Preview Sessions.

**Blocked by:** 12 — Deliver Recent Targets core semantics.

**Status:** ready-for-agent

- [ ] A missing, inaccessible, or temporarily offline Recent Target remains visible with an unavailable state.
- [ ] Retrying an available target uses the application coordinator and updates history only after success.
- [ ] Removing one Recent Target updates the panel and persistence immediately.
- [ ] Clearing Recent Targets updates the panel and persistence immediately.
- [ ] Remove and Clear never stop or hide an active Preview Session.
- [ ] A target on removable storage can become available again without recreating its record.
- [ ] Tests cover unavailable, retry, remove, clear, and active-state independence.
