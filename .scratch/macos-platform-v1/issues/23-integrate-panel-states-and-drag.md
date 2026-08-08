# 23 — Integrate panel states and drag presentation

**What to build:** Make every Menu Bar Panel state visually coherent and complete the drag-and-drop presentation within the native panel structure.

**Blocked by:** 14 — Restore the menu bar application shell; 22 — Compose the native Menu Bar Panel structure.

**Status:** ready-for-agent

- [ ] Empty, first-run, Starting, Running, Stopping, recent, unavailable, error, and settings states fit the shared panel structure.
- [ ] The full panel accepts a supported Open Target through drag-and-drop.
- [ ] Drag entry distinguishes valid and invalid input without hiding essential context.
- [ ] Drag exit, cancellation, rejection, and completion restore normal transient panel behavior.
- [ ] State transitions do not produce clipped, overlapping, or unreachable controls.
- [ ] Focus remains in a predictable area when banners or list sections appear and disappear.
- [ ] UI tests cover each state plus valid, invalid, cancelled, and completed drag paths.
