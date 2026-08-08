# 15 — Add menu bar commands and quit semantics

**What to build:** Give the menu bar item a native right-click command menu with safe global lifecycle actions.

**Blocked by:** 07 — Complete Preview Session stop and termination; 14 — Restore the menu bar application shell.

**Status:** ready-for-agent

- [ ] Right-clicking the menu bar item offers Open, Stop All, Settings, About, and Quit.
- [ ] Open uses the same application coordinator path as the panel's Open action.
- [ ] Stop All is disabled when no Preview Sessions exist and asks for confirmation when active.
- [ ] Confirming Stop All terminates every owned Preview Session and leaves Recent Targets intact.
- [ ] Quit is immediate when no Preview Sessions exist.
- [ ] Quit with active sessions explains that they will stop, asks for confirmation, and leaves no owned process behind.
- [ ] Cancelling either confirmation makes no lifecycle change.
- [ ] Relaunch restores Recent Targets but never restarts Preview Sessions.
