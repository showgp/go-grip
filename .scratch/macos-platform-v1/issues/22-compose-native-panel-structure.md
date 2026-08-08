# 22 — Compose the native Menu Bar Panel structure

**What to build:** Assemble the completed workflows into a compact native panel hierarchy with stable sizing and scrolling, without yet taking on drag-state polish or localization.

**Blocked by:** 13 — Add Recent Target maintenance; 17 — Add browser failure recovery; 19 — Deliver minimal Settings.

**Status:** ready-for-agent

- [ ] The hierarchy is header, optional error banner, Preview Sessions, Recent Targets, and primary Open action.
- [ ] The header contains GoGrip product identity, the active Preview Session count, and direct access to Settings.
- [ ] Width remains approximately 360 points and height grows with content to approximately 520 points.
- [ ] Content beyond the maximum height scrolls internally while the header and primary action remain predictably reachable.
- [ ] System typography, spacing, materials, controls, and SF Symbols are used consistently.
- [ ] Emoji headings, oversized empty artwork, mixed icon styles, and user-resizable geometry are absent.
- [ ] Empty, short, and long-list fixtures verify stable size and scroll behavior.
