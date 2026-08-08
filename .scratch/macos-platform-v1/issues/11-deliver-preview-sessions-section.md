# 11 — Deliver the Preview Sessions section

**What to build:** Give the Menu Bar Panel a dedicated Preview Sessions section that accurately presents active lifecycle state and lets the user revisit or stop each preview.

**Blocked by:** 07 — Complete Preview Session stop and termination.

**Status:** ready-for-agent

- [ ] Starting, Running, and Stopping Preview Sessions appear in a dedicated section.
- [ ] Starting and Stopping rows communicate progress and disable invalid actions.
- [ ] A Running row can revisit the browser preview and explicitly stop the session.
- [ ] The section updates immediately when lifecycle state changes or a process exits.
- [ ] Active entries are ordered by the most recent start or revisit.
- [ ] Empty active state does not reserve a misleading row or report a stale count.
- [ ] UI and coordinator tests cover each state and row action.
