# 38 — Publish the verified DMG

**What to build:** Make public release upload a small immutable-artifact gate that can publish only the exact DMG that passed every clean-install test.

**Blocked by:** 37 — Smoke-test localization and accessibility.

**Status:** ready-for-agent

- [ ] Publication requires successful product tests, signing, notarization, trust verification, core smoke, localization smoke, and accessibility smoke.
- [ ] The candidate digest is checked immediately before upload and matches every downstream test record.
- [ ] Gatekeeper assessment and notarization-staple validation run again against the unchanged candidate immediately before upload.
- [ ] No rebuild, repackaging, restapling, or metadata mutation occurs between successful tests and upload.
- [ ] Release version and DMG version match the tag.
- [ ] Failed or missing evidence prevents upload and leaves diagnostic artifacts available.
- [ ] The uploaded asset is the exact immutable DMG verified by Tickets 35–37.
