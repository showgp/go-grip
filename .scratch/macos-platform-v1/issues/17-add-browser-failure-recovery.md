# 17 — Add browser failure recovery

**What to build:** Preserve a healthy Preview Session when macOS cannot open its browser address and give the user direct recovery actions.

**Blocked by:** 16 — Add startup and exit error recovery.

**Status:** ready-for-agent

- [ ] Browser opening is treated as a separate effect after Preview Session startup.
- [ ] A browser-open failure leaves the Preview Session Running.
- [ ] The error banner offers Open Again, Copy Address, and Stop.
- [ ] Open Again uses the existing Preview Session address and does not launch another process.
- [ ] Copy Address places the exact preview address on the pasteboard with accessible confirmation.
- [ ] Stop follows the normal Preview Session stopping lifecycle.
- [ ] Tests use a controlled browser adapter and never launch the real browser.
