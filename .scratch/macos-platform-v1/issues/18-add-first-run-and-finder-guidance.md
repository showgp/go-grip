# 18 — Add first-run and Finder Command guidance

**What to build:** Help a first-time user discover the Menu Bar Panel, Finder Command, Open action, and drag-and-drop without introducing a separate welcome application.

**Blocked by:** 10 — Remove Finder Sync from the product; 14 — Restore the menu bar application shell.

**Status:** ready-for-agent

- [ ] The Menu Bar Panel opens automatically once on first launch with concise inline guidance.
- [ ] Guidance explains the menu bar item, Finder Command, Open, and drag-and-drop.
- [ ] Returning users do not see first-run guidance again unless test state is explicitly reset.
- [ ] The application refreshes macOS Services registration on launch.
- [ ] The panel or Settings reports the Finder Command's expected availability without promising a top-level menu position.
- [ ] A user can open the relevant system Services settings from GoGrip.
- [ ] No guidance references Finder Sync or asks the user to enable a Finder extension.
- [ ] Tests isolate first-run and Service-registration state from the real user account.
