# 01 — Wire Swift tests and CI

**What to build:** Make the existing macOS test suite a real, isolated build target that runs locally and in pull-request CI, providing a green foundation for the remaining macOS V1 work.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] The macOS project exposes a shared test action that runs through the standard Xcode test command.
- [ ] Existing storage, process-management, startup-reader, and integration tests are discovered and executed.
- [ ] Tests use isolated preferences and temporary fixtures rather than the real user's Recent Targets or files.
- [ ] Pull-request CI runs the macOS tests and fails when a test fails.
- [ ] The existing Go test suite and macOS application build remain green.
