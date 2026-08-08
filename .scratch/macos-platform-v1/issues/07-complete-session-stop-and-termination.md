# 07 — Complete Preview Session stop and termination

**What to build:** Give Preview Sessions a reliable Stopping path, clean application ownership, and correct handling when the preview process exits unexpectedly.

**Blocked by:** 06 — Make Preview Session startup trustworthy.

**Status:** ready-for-agent

- [ ] Stop transitions a Running Preview Session through Stopping and removes it only after termination is handled.
- [ ] Repeated Stop requests are idempotent and controls are disabled while stopping.
- [ ] Unexpected process exit removes active state without racing a later startup-state update.
- [ ] Application shutdown stops every owned Preview Session and waits or escalates within a bounded interval.
- [ ] Closing the browser window has no effect on a Running Preview Session, which remains available to revisit or stop.
- [ ] No stopped, failed, or replaced process is left without an owner that can terminate it.
- [ ] Relaunch begins with no restored Preview Sessions.
- [ ] Tests cover browser closure, clean stop, forced termination, unexpected exit, concurrent stop, and shutdown cleanup.
