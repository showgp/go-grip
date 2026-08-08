# 05 — Enforce one Preview Session per Open Target

**What to build:** Ensure repeated and concurrent requests for one normalized Open Target join or revisit one Preview Session rather than launching duplicate background processes.

**Blocked by:** 03 — Migrate existing Open Target entry points; 04 — Normalize Open Target identity.

**Status:** ready-for-agent

- [ ] Reopening a running Open Target revisits its existing Preview Session and browser address.
- [ ] Requests arriving while the target is starting join the in-flight operation rather than launching again.
- [ ] Real-path and symbolic-link requests share one Preview Session.
- [ ] Different Open Targets can still run concurrently as separate Preview Sessions.
- [ ] Revisit activity is observable for later ordering without creating a second session.
- [ ] Concurrency tests prove that only one process launch occurs per normalized Open Target.
