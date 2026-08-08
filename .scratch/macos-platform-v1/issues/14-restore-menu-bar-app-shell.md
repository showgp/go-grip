# 14 — Restore the menu bar application shell

**What to build:** Make GoGrip a true menu-bar-only application whose primary panel opens and dismisses like a native transient macOS control.

**Blocked by:** 03 — Migrate existing Open Target entry points.

**Status:** ready-for-agent

- [ ] GoGrip runs as an `LSUIElement` application with no Dock or application-switcher presence.
- [ ] Left-clicking the menu bar item toggles the Menu Bar Panel.
- [ ] Clicking outside the panel closes it during normal use.
- [ ] The panel remains available while a supported drag is active or a system Open dialog is presented, then returns to transient behavior.
- [ ] Drag cancellation and unsupported drops cannot leave the panel permanently non-transient.
- [ ] First launch and relaunch preserve the menu-bar-only lifecycle.
- [ ] Build-product and UI tests verify Dock absence, panel toggle, dismissal, and drag transitions.
