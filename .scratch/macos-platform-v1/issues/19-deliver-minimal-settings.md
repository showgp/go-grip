# 19 — Deliver minimal Settings

**What to build:** Provide only the V1 settings needed to control launch behavior, diagnose Finder Command availability, and identify the installed build.

**Blocked by:** 18 — Add first-run and Finder Command guidance.

**Status:** ready-for-agent

- [ ] Launch at login is disabled by default and changes only after explicit user action.
- [ ] Registration and unregistration failures show a meaningful error and restore the toggle to actual system state.
- [ ] Settings expose Finder Command status and the system Services settings action established by first-run guidance.
- [ ] Version and build information match the installed application metadata.
- [ ] Settings do not expose language, port, recursion, theme, or session-restoration controls.
- [ ] Tests use a controlled login-item adapter and never modify the real user's login items.
