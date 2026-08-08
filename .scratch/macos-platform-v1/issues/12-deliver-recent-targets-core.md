# 12 — Deliver Recent Targets core semantics

**What to build:** Present a trustworthy Recent Targets section that records successful work, remains independent from Preview Sessions, and avoids duplicate active rows.

**Blocked by:** 05 — Enforce one Preview Session per Open Target; 11 — Deliver the Preview Sessions section.

**Status:** ready-for-agent

- [ ] A Recent Target is created or moved to the top only after a successful start or revisit.
- [ ] Failed opening requests never enter recent history.
- [ ] Recent Targets are deduplicated by normalized Open Target identity.
- [ ] An active target is suppressed from the visible Recent Targets section without deleting its persisted record.
- [ ] Stopping a Preview Session makes its Recent Target visible again in the correct order.
- [ ] Activating a Recent Target starts or revisits through the application coordinator.
- [ ] Recent Targets persist across application relaunch while Preview Sessions do not.
