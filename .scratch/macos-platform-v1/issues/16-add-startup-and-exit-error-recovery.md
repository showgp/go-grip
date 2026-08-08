# 16 — Add startup and exit error recovery

**What to build:** Turn preview startup failures and unexpected exits into visible, recoverable application outcomes instead of false or silently disappearing state.

**Blocked by:** 07 — Complete Preview Session stop and termination; 08 — Add the Markdown Finder Command Service; 11 — Deliver the Preview Sessions section; 14 — Restore the menu bar application shell.

**Status:** ready-for-agent

- [ ] Process launch failure, startup timeout, malformed startup information, early exit, and unexpected running exit produce meaningful error outcomes.
- [ ] A failed start leaves no Preview Session and does not create a new Recent Target.
- [ ] An unexpected exit removes the Preview Session and retains an existing Recent Target.
- [ ] Finder Command failures automatically open the Menu Bar Panel and present the error rather than failing silently.
- [ ] The panel shows a dismissible error banner with an appropriate Retry action.
- [ ] An error that occurs while the panel is closed creates one unread-error state.
- [ ] Opening the panel presents the error and predictably acknowledges the unread state.
- [ ] Consuming or dismissing an error cannot recursively publish the same error.
