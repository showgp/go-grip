# 34 — Build a notarized and validated DMG

**What to build:** Turn the completed signed archive into one immutable notarized, stapled, Gatekeeper-approved DMG candidate.

**Blocked by:** 33 — Build a signed universal archive.

**Status:** ready-for-agent

- [ ] Notarization uses protected credentials and waits for a successful result.
- [ ] Failure preserves useful diagnostics and stops the release.
- [ ] The ticket is stapled and validated before DMG construction.
- [ ] DMG construction does not rebuild or mutate the signed application afterward.
- [ ] The DMG and contained application pass signature, staple, Gatekeeper, architecture, version, and minimum-system checks.
- [ ] A cryptographic digest identifies the immutable DMG candidate for downstream tests and publication.
- [ ] The candidate is retained internally but is not uploaded publicly by this ticket.
