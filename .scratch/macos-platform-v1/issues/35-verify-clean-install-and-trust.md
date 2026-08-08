# 35 — Verify clean installation and trust

**What to build:** Install only the immutable DMG candidate in a clean macOS account or VM and prove that packaging, trust, application shape, and Service discovery are correct.

**Blocked by:** 34 — Build a notarized and validated DMG.

**Status:** ready-for-agent

- [ ] The test mounts the identified DMG, installs its application, and launches only that installed copy.
- [ ] Gatekeeper, signature, staple, minimum system version, universal architectures, and bundled preview executable are verified again after installation.
- [ ] First launch shows the menu bar item without a Dock or application-switcher presence.
- [ ] The installed bundle declares `LSUIElement`, contains the Finder Command Service, and contains no Finder Sync extension.
- [ ] Finder discovers the Service after installation without relying on a developer-machine registration.
- [ ] First-run guidance appears once and relaunch does not repeat it.
- [ ] Failures retain logs and screenshots and clean up only the isolated environment.
