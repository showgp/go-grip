# 36 — Smoke-test Finder and Preview Session lifecycle

**What to build:** Exercise the core Finder Command, Preview Session, Recent Target, browser, and shutdown journey using the clean-installed DMG candidate.

**Blocked by:** 35 — Verify clean installation and trust.

**Status:** ready-for-agent

- [ ] With GoGrip stopped, Finder opens one Markdown file and one directory through the Service.
- [ ] Each request starts the installed application, opens a real responding local preview, and records the successful Recent Target.
- [ ] Repeat and symbolic-link requests reuse one Preview Session and do not create another process.
- [ ] Revisit, Stop, Stop All, browser-close-without-stop, browser recovery, and unexpected-exit recovery behave as specified.
- [ ] Quit leaves no orphan preview process, and relaunch restores Recent Targets without restoring Preview Sessions.
- [ ] Missing, unsupported, and unavailable targets produce the specified recovery behavior.
- [ ] Logs and process evidence identify the exact DMG digest under test.
