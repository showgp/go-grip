# 33 — Build a signed universal archive

**What to build:** Produce a Developer ID-signed archive from the completed macOS V1 product, valid on Apple Silicon and Intel Macs.

**Blocked by:** 29 — Make Open, Preview Session, and errors keyboard and VoiceOver accessible; 30 — Make Recent Targets, Settings, and menus keyboard and VoiceOver accessible; 32 — Support Reduce Motion and text scaling.

**Status:** ready-for-agent

- [ ] Release builds enable Hardened Runtime and use a configured Developer ID Application identity and Team ID.
- [ ] The application and every nested executable are signed in the correct order with a consistent identity.
- [ ] The bundled preview executable contains arm64 and x86_64 architectures.
- [ ] The archive contains the completed V1 UI, both localizations, accessibility behavior, macOS Service, `LSUIElement`, and no Finder Sync extension.
- [ ] Signature verification checks deep structure, strict validity, Team identity, and Hardened Runtime.
- [ ] Credentials use protected CI secrets and a temporary keychain and are removed after the job.
- [ ] The job fails before packaging when product tests, archive validation, or signing validation fail.
