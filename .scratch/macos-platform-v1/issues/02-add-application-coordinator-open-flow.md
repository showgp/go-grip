# 02 — Add the application coordinator through Open

**What to build:** Introduce one application-level coordination seam and route the Menu Bar Panel's Open command through it, so one observable state owns opening, browser launch, Preview Session state, Recent Target persistence, and user-facing errors.

**Blocked by:** 01 — Wire Swift tests and CI.

**Status:** ready-for-agent

- [ ] Choosing one supported Open Target through Open completes the existing preview-and-browser workflow through the coordinator.
- [ ] The coordinator exposes application state that the Menu Bar Panel can observe without duplicating process or persistence state.
- [ ] Process launch, browser opening, persistence, filesystem lookup, and time are controllable in tests.
- [ ] Success and failure are verified through coordinator inputs and observable effects rather than private call counts.
- [ ] Existing Open behavior remains usable while the new seam is introduced.
