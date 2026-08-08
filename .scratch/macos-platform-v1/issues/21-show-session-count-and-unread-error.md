# 21 — Show session count and unread error state

**What to build:** Make the menu bar item communicate active Preview Session count and unseen failure without an overlaid, clipped, or color-only badge.

**Blocked by:** 15 — Add menu bar commands and quit semantics; 16 — Add startup and exit error recovery; 20 — Create the GoGrip menu bar template icon.

**Status:** ready-for-agent

- [ ] Zero active Preview Sessions show only the GoGrip template icon.
- [ ] One or more active Preview Sessions show a compact adjacent count that updates after lifecycle changes.
- [ ] An unread error shows an explicit exclamation indicator and remains understandable without color.
- [ ] Opening the panel acknowledges the error indicator only after the error is available to the user.
- [ ] Count and error presentation coexist without clipping or obscuring the template icon.
- [ ] Light, dark, highlighted, and increased-contrast appearances remain legible.
- [ ] Tests verify updates after start, revisit, stop, Stop All, failure, acknowledgement, and unexpected exit.
