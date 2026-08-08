# 08 — Add the Markdown Finder Command Service

**What to build:** Let a user invoke Open with GoGrip for one Markdown file through macOS Services, automatically launching GoGrip when needed and completing the preview in the browser.

**Blocked by:** 06 — Make Preview Session startup trustworthy.

**Status:** ready-for-agent

- [ ] The application declares a Finder-scoped Service for Markdown input with localized menu metadata.
- [ ] Selecting one Markdown file in Finder invokes the application coordinator with one Open Target.
- [ ] The Service works when GoGrip is not already running.
- [ ] A successful invocation starts or revisits the Preview Session and opens the browser directly.
- [ ] The Service receives its input through the system pasteboard contract and validates it again at runtime.
- [ ] Service adapter tests use controlled pasteboard input and do not change the user's Services preferences.
- [ ] Real installation verification confirms the command appears in Finder's system-managed Services location.
