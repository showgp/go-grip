# 32 — Support Reduce Motion and text scaling

**What to build:** Keep the Menu Bar Panel stable and usable when the user reduces motion or increases text size.

**Blocked by:** 23 — Integrate panel states and drag presentation; 28 — Localize content flows in Simplified Chinese; 31 — Make menu status and appearance visually accessible.

**Status:** ready-for-agent

- [ ] Reduce Motion suppresses nonessential panel, banner, row, and drag transitions while retaining state feedback.
- [ ] Larger accessibility text sizes keep critical labels and actions visible through wrapping or scrolling.
- [ ] Bounded panel geometry does not clip confirmations, errors, translated strings, or primary actions.
- [ ] Dynamic state changes do not cause disorienting focus or scroll jumps.
- [ ] English and Simplified Chinese are verified at representative large text sizes.
- [ ] Automated or repeatable visual checks cover normal/reduced motion and normal/large text.
