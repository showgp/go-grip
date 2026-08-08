# 06 — Make Preview Session startup trustworthy

**What to build:** Represent startup explicitly and declare a Preview Session running only after valid startup information is received from a viable preview process.

**Blocked by:** 05 — Enforce one Preview Session per Open Target.

**Status:** ready-for-agent

- [ ] A new request becomes Starting before process launch and Running only after valid machine-readable startup information.
- [ ] Starting controls cannot issue a second start or an invalid revisit.
- [ ] Timeout, end-of-stream, malformed startup information, invalid address data, and early process exit all fail the request.
- [ ] Startup failure never creates a Running Preview Session.
- [ ] No failure path guesses or falls back to port 6419.
- [ ] Process diagnostics needed for a later user-facing error are retained without writing sensitive data.
- [ ] Tests cover every success and failure transition with a controlled preview executable.
