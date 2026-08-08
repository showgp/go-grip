# 10 — Remove Finder Sync from the product

**What to build:** Complete the Finder integration migration by removing Finder Sync, its custom-URL path, and its packaged extension while keeping the macOS Service fully functional.

**Blocked by:** 09 — Add directory Finder Command and input hardening.

**Status:** ready-for-agent

- [ ] The macOS build graph no longer defines, embeds, signs, or archives a Finder Sync extension.
- [ ] The distributed application contains no Finder Sync extension bundle.
- [ ] The old custom-URL launch path and extension-specific configuration are removed.
- [ ] Users are not asked to enable a Finder extension during onboarding or diagnosis.
- [ ] The Markdown and directory Finder Commands continue to work after the old path is removed.
- [ ] Build-product tests fail if Finder Sync is accidentally reintroduced.
