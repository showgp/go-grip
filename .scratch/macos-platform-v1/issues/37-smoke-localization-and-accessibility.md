# 37 — Smoke-test localization and accessibility

**What to build:** Verify the localized and accessible V1 journeys in the same clean-installed immutable DMG candidate.

**Blocked by:** 36 — Smoke-test Finder and Preview Session lifecycle.

**Status:** ready-for-agent

- [ ] English and Simplified Chinese each cover Finder Command, panel, menus, lifecycle, recovery, history, confirmations, onboarding, and Settings without mixed-language gaps.
- [ ] Keyboard and VoiceOver complete the primary Open, revisit, stop, error recovery, history maintenance, confirmation, and Settings journeys.
- [ ] Light, dark, increased-contrast, Reduce Motion, and representative large-text configurations remain usable.
- [ ] Icon, count, unread error, lifecycle, unavailable state, and focus remain understandable without color alone.
- [ ] Bounded panel geometry does not clip critical translated or scaled actions.
- [ ] Results, screenshots, and accessibility evidence reference the same DMG digest used by prior smoke tests.
