# 04 — Normalize Open Target identity

**What to build:** Give each Markdown file or directory a stable Open Target identity while preserving the user's selected path for display.

**Blocked by:** 02 — Add the application coordinator through Open.

**Status:** ready-for-agent

- [ ] Relative and absolute references to the same filesystem object resolve to one identity.
- [ ] A symbolic link and its resolved destination identify the same Open Target.
- [ ] The user-selected path remains available for display even when identity uses a resolved path.
- [ ] Markdown filename matching is case-insensitive and directories are accepted.
- [ ] Missing and unsupported targets are rejected before process launch with a meaningful outcome.
- [ ] Identity behavior is covered using temporary real files, directories, and symbolic links.
