# GoGrip macOS Platform V1

Status: ready-for-agent

## Problem Statement

GoGrip already has a working Go preview backend and the beginnings of a macOS menu bar application, but the macOS experience is not yet reliable or distributable. The current Finder integration uses Finder Sync for a general-purpose action, does not provide a functioning end-to-end launch path, and can create Preview Sessions that the user cannot reliably revisit or stop. The Menu Bar Panel mixes Recent Targets with active state, contains broken lifecycle and error paths, and lacks a coherent native visual language. The application also has no complete signed, notarized, clean-install release gate, so a successful build does not demonstrate that the product is usable after installation.

## Solution

Deliver GoGrip as a persistent, native macOS 13+ menu bar application that uses a macOS Service as its Finder Command. A user can select one Markdown file or directory in Finder and open it directly in GoGrip, or use the Menu Bar Panel's Open command and drag-and-drop surface. GoGrip normalizes the Open Target, reuses an existing Preview Session when present, and opens the preview in the user's browser.

The Menu Bar Panel separately presents active Preview Sessions and Recent Targets, provides explicit revisit and stop controls, and communicates failures without creating false running state. The application uses a compact, localized, accessible macOS appearance with a purpose-built monochrome menu bar icon. Public releases are universal, Developer ID-signed, notarized, validated from the final DMG, and blocked unless the full clean-install workflow succeeds.

## User Stories

1. As a macOS user, I want GoGrip to live in the menu bar without a Dock icon, so that it remains available without behaving like a conventional document-window application.
2. As a macOS user, I want to select one Markdown file in Finder and invoke Open with GoGrip, so that I can preview it without using Terminal.
3. As a macOS user, I want to select one directory in Finder and invoke Open with GoGrip, so that I can browse the directory as a recursive documentation workspace.
4. As a macOS user, I want the Finder Command to be absent for unsupported file types, so that I never choose an action that silently does nothing.
5. As a macOS user, I want Markdown extension matching to be case-insensitive, so that valid Markdown files are not rejected because of filename casing.
6. As a macOS user, I want the Finder Command to work outside my home directory when I have access, so that projects on external volumes or shared locations behave consistently.
7. As a macOS user, I want the Finder Command to launch GoGrip when it is not running, so that I do not need to start the application first.
8. As a macOS user, I want the Finder Command to immediately open the preview in my browser, so that it completes the requested action without another confirmation step.
9. As a macOS user, I want macOS to manage the Finder Command's Services menu placement, so that GoGrip uses the platform-supported integration model.
10. As a macOS user, I want Settings to tell me whether the Finder Command is available, so that I can diagnose a missing Services entry.
11. As a macOS user, I want a shortcut from Settings to the system Services configuration, so that I can enable, disable, or assign a keyboard shortcut to the Finder Command.
12. As a macOS user, I want Finder Sync removed from the distributed application, so that I do not need to enable an unrelated Finder extension.
13. As a macOS user, I want the Menu Bar Panel to provide an Open command, so that I can choose an Open Target without Finder's context menu.
14. As a macOS user, I want to drop a Markdown file or directory anywhere on the Menu Bar Panel, so that frequent opening is fast.
15. As a macOS user, I want an unsupported drop to produce clear feedback, so that I know why it was rejected.
16. As a macOS user, I want opening a directory from any macOS entry point to use recursive discovery, so that the same Open Target always produces the same workspace.
17. As a macOS user, I want a symbolic link and its resolved destination to identify the same Open Target, so that I do not accidentally create duplicate previews.
18. As a macOS user, I want GoGrip to preserve the path I selected for display, so that the interface reflects how I navigated to the Open Target.
19. As a macOS user, I want an Open Target to have at most one Preview Session, so that repeated opening does not leave duplicate background processes.
20. As a macOS user, I want reopening an active Open Target to revisit its existing browser preview, so that the action is quick and deterministic.
21. As a macOS user, I want concurrent requests for the same Open Target to collapse into one start operation, so that rapid repeated actions do not race.
22. As a macOS user, I want a starting Preview Session to show progress, so that I know GoGrip received my request.
23. As a macOS user, I want controls that are invalid while a Preview Session is starting or stopping to be disabled, so that I cannot create contradictory operations.
24. As a macOS user, I want a running Preview Session to provide revisit and stop actions, so that I remain in control of its lifecycle.
25. As a macOS user, I want closing the browser window not to stop the Preview Session, so that I can revisit it later from the Menu Bar Panel.
26. As a macOS user, I want startup timeout, early process exit, or invalid startup information to be treated as failure, so that GoGrip never displays a false Preview Session.
27. As a macOS user, I want GoGrip not to guess a default address after startup failure, so that it never opens an unrelated local service.
28. As a macOS user, I want browser-launch failure to preserve a healthy Preview Session, so that I can retry opening or copy its address.
29. As a macOS user, I want an unexpectedly terminated Preview Session removed from active state, so that the Menu Bar Panel never claims a dead preview is running.
30. As a macOS user, I want an unexpected termination to leave its Recent Target intact, so that I can try again.
31. As a macOS user, I want an unseen failure indicated in the menu bar, so that background failure is discoverable without notification permission.
32. As a macOS user, I want opening the Menu Bar Panel to show and acknowledge the latest error, so that the warning indicator clears predictably.
33. As a macOS user, I want a failed Finder Command to open the Menu Bar Panel and show a dismissible error banner with a retry action, so that recovery is immediate and stays close to session management.
34. As a macOS user, I want successful openings recorded as Recent Targets, so that I can quickly return to previous work.
35. As a macOS user, I want failed opening attempts excluded from Recent Targets, so that recent history remains trustworthy.
36. As a macOS user, I want Preview Sessions and Recent Targets shown in separate sections, so that active state is not confused with history.
37. As a macOS user, I want an active Open Target hidden from the Recent Targets section, so that the same target is not duplicated in the panel.
38. As a macOS user, I want a stopped target to reappear in Recent Targets, so that stopping it does not erase history.
39. As a macOS user, I want active and recent entries ordered by most recent start or revisit, so that current work stays near the top.
40. As a macOS user, I want a missing or temporarily unavailable Recent Target retained and marked unavailable, so that an offline volume does not erase my history.
41. As a macOS user, I want to remove one Recent Target without stopping an active Preview Session, so that history and runtime state remain independent.
42. As a macOS user, I want clearing Recent Targets to immediately update the panel without affecting Preview Sessions, so that the action has visible and bounded consequences.
43. As a macOS user, I want Stop All to ask for confirmation, so that I do not terminate all previews accidentally.
44. As a macOS user, I want quitting with active Preview Sessions to explain that they will stop and ask for confirmation, so that application ownership is explicit.
45. As a macOS user, I want quitting with no active Preview Sessions to happen immediately, so that an unnecessary confirmation does not slow me down.
46. As a macOS user, I want GoGrip to stop all owned Preview Sessions when it exits, so that it leaves no manageable work behind.
47. As a macOS user, I want relaunching GoGrip to preserve Recent Targets but not restart Preview Sessions, so that startup remains quiet and predictable.
48. As a macOS user, I want launch at login disabled by default, so that GoGrip does not add itself to login items without consent.
49. As a macOS user, I want to enable or disable launch at login from Settings, so that I control whether GoGrip starts automatically.
50. As a first-time user, I want the Menu Bar Panel to open once with concise guidance, so that I can discover the icon, Finder Command, Open command, and drag-and-drop behavior.
51. As a returning user, I want first-run guidance not to reappear, so that routine use remains unobtrusive.
52. As a macOS user, I want left-clicking the menu bar item to toggle the Menu Bar Panel, so that its primary interaction is fast.
53. As a macOS user, I want right-clicking the menu bar item to show Open, Stop All, Settings, About, and Quit, so that essential commands remain available outside the panel.
54. As a macOS user, I want clicking outside the panel to close it except during an active drag or system dialog, so that it behaves like a native transient panel.
55. As a macOS user, I want a compact session count next to the icon only when sessions are active, so that status is legible without a clipped badge.
56. As a macOS user, I want an unread error represented by an exclamation indicator rather than color alone, so that failure remains accessible.
57. As a macOS user, I want a distinctive monochrome GoGrip menu bar icon, so that I can identify the application while preserving native light/dark adaptation.
58. As a macOS user, I want the icon to use a simplified angular M derived from the existing brand rather than the full shield, so that it remains clear at menu bar size.
59. As a macOS user, I want the panel to use system typography, spacing, materials, and SF Symbols without emoji, so that it feels native and consistent.
60. As a macOS user, I want the panel to remain approximately 360 points wide and grow only to a bounded height with internal scrolling, so that its layout is stable.
61. As a Chinese-speaking user, I want GoGrip localized in Simplified Chinese, so that Finder, panel, settings, and error language is consistent.
62. As an English-speaking user, I want GoGrip localized in English, so that the application follows my system language.
63. As a keyboard user, I want every panel action reachable with visible focus, so that a pointer is not required.
64. As a VoiceOver user, I want meaningful labels, values, and actions for every icon and session control, so that runtime state is understandable.
65. As a user with Reduce Motion enabled, I want nonessential panel animations suppressed, so that GoGrip respects my system preference.
66. As a user with larger text settings, I want critical labels and actions to remain readable and operable, so that scaling does not hide functionality.
67. As a user, I want all status indicators to communicate meaning without relying only on color, so that state remains distinguishable.
68. As a user, I want the application to adapt correctly to light and dark menu bar appearances, so that icon and panel content remain legible.
69. As a person downloading GoGrip, I want a Developer ID-signed and notarized application, so that macOS can verify its publisher and integrity.
70. As a person downloading GoGrip, I want the final DMG validated by Gatekeeper before publication, so that the distributed artifact matches the tested artifact.
71. As a person installing GoGrip, I want the bundled preview executable to support both Apple Silicon and Intel Macs, so that the universal application works on supported hardware.
72. As a maintainer, I want release publication blocked when automated tests fail, so that known regressions do not ship.
73. As a maintainer, I want signing and notarization verification to fail the release when incomplete, so that an untrusted DMG cannot be uploaded accidentally.
74. As a maintainer, I want a clean-install smoke test to run against the final DMG, so that packaging, Services discovery, menu bar behavior, and process cleanup are validated together.
75. As a maintainer, I want test data, UserDefaults, browser opening, and login-item changes isolated from the real user account, so that the test suite is safe to run repeatedly.
76. As a maintainer, I want release credentials supplied through protected CI secrets and temporary signing infrastructure, so that private credentials never enter the repository.

## Implementation Decisions

- The Go preview backend and its browser-based rendering behavior are treated as an existing dependency. This effort changes the macOS host, integration, lifecycle, presentation, tests, packaging, and release pipeline; it does not redesign the backend.
- The macOS product is a persistent `LSUIElement` menu bar application with no Dock or application-switcher presence. It supports macOS 13 and later and is distributed as a universal application.
- The Finder Command is implemented as a macOS Service. Finder Sync is removed from the product, build graph, package, onboarding, and release validation.
- The Service advertises only Markdown and directory inputs and is constrained to Finder. The application still validates that exactly one supported Open Target was delivered because service input is external input.
- The Finder Command's name is localized as “Open with GoGrip” and its Simplified Chinese equivalent. macOS owns its menu position, ordering, enablement, and optional keyboard shortcut.
- Service registration is refreshed at application launch. Settings report its availability and link to the system Services configuration when the user needs to enable or inspect it.
- All entry points produce the same application-level Open Target request. Finder Command, Open panel, drag-and-drop, Recent Target activation, and Preview Session revisit do not implement separate lifecycle flows.
- A new application coordination boundary is the primary behavioral and testing seam. It accepts Open Target requests and user commands, owns the visible application state, and coordinates path identity, Preview Session lifecycle, Recent Target persistence, browser opening, and error presentation.
- Open Target identity uses standardized absolute paths with symbolic-link resolution. The resolved identity is used for equality and session ownership; the user-selected path is retained for display.
- Directory Open Targets always request recursive document discovery from the existing backend.
- One normalized Open Target may have at most one Preview Session. A request for a target that is starting or running joins or revisits the existing operation instead of launching another process.
- Preview Session state is explicit: Starting, Running, and Stopping. Failed startup is an error outcome, not a Preview Session state.
- Startup succeeds only after valid machine-readable startup information is received and the child process remains viable. Timeout, end-of-stream, malformed information, or early exit fails the request; there is no default-port fallback.
- Browser opening is a separate effect after Preview Session startup. Browser failure leaves a healthy Preview Session intact and exposes Open Again, Copy Address, and Stop actions.
- Unexpected child termination removes the Preview Session, retains any existing Recent Target, and produces an unread error. Opening the Menu Bar Panel acknowledges the unread indicator after presenting the error.
- A Finder Command failure automatically opens the Menu Bar Panel and presents its error banner; failures from other entry points use the same banner without requiring system notification permission.
- Closing a browser has no effect on Preview Session ownership. Sessions stop only through explicit Stop, Stop All, application shutdown, or unexpected process termination.
- Clean application termination stops every owned Preview Session. Relaunch restores Recent Targets but never restores Preview Sessions automatically.
- A Recent Target is created or moved to the top only after a successful opening or revisit. Failed requests are not persisted as recent history.
- Preview Sessions and Recent Targets are independent collections in application state and separate sections in the panel. An active target is suppressed from the visible Recent Targets section without removing its persisted record.
- Active entries are ordered by most recent start or revisit. Recent entries are ordered by most recent successful opening or revisit.
- An unavailable Recent Target remains visible with unavailable status and retry/remove actions. It is not automatically deleted because removable media may return.
- Removing or clearing Recent Targets never stops Preview Sessions. The visible state updates immediately after persistence changes.
- The Menu Bar Panel contains a header with product identity, active count, and Settings; an optional dismissible error banner; Preview Sessions; Recent Targets; and a primary Open command. The full panel accepts supported drag-and-drop input.
- Left click toggles the Menu Bar Panel. Right click presents a native menu containing Open, Stop All, Settings, About, and Quit.
- The panel behaves transiently and closes when focus moves outside it, except while processing a drag or presenting a system dialog.
- Stop All requires confirmation. Quit requires confirmation only when Preview Sessions are active and always explains that those sessions will stop.
- First launch opens the panel once with concise inline guidance. Launch at login is disabled by default and changed only by explicit user action.
- Settings are limited to launch at login, Finder Command status/system-settings access, and version/build information. Language, ports, recursive behavior, theme, and session restoration are not user-configurable.
- The menu bar icon is a purpose-built monochrome template asset: a simplified angular M derived from the existing GoGrip logo and optically adjusted for 16–18 point rendering. The full shield is not scaled into the menu bar.
- Active count is rendered as compact adjacent text rather than an overlaid red badge. An unread error uses a visible exclamation indicator and does not depend on color alone.
- The panel uses native macOS typography, spacing, materials, controls, and SF Symbols. It avoids emoji and mixed-language strings, remains approximately 360 points wide, grows with content to about 520 points, and scrolls internally beyond that height.
- User-visible strings are localized in English and Simplified Chinese and follow the system language. Accessibility includes keyboard operation, visible focus, VoiceOver semantics, Reduce Motion, scalable text, light/dark adaptation, and non-color-only state communication.
- Public distribution uses Developer ID signing, Hardened Runtime, Apple notarization, stapling, and a universal bundled preview executable. The Mac App Store and unsigned public distribution are not targets.
- Release credentials are injected through protected secrets and a temporary CI keychain. Certificates, private keys, passwords, and notarization credentials are never stored in the repository.
- Release publication is a hard dependency chain: tests, signed archive, signature verification, notarization and staple, final DMG construction, clean-install smoke test from that DMG, Gatekeeper/staple validation, then upload. A failed or skipped stage prevents publication.

## Testing Decisions

- Tests assert externally visible behavior and domain invariants rather than private SwiftUI layout structure or implementation call counts.
- The primary test seam is the application coordination boundary. Tests send Open Target requests and user commands through this boundary and observe application state plus controlled process, persistence, browser, clock, and filesystem effects.
- Finder Command, Open panel, drag-and-drop, Recent Target selection, and revisit tests reuse the same coordinator behavior suite. Entry adapters receive focused contract tests for input decoding and validation rather than duplicating lifecycle tests.
- Existing Storage, ProcessManager, startup-information reader, and integration test cases are retained as prior art, connected to a real unit-test target, and updated where their old expectations conflict with this specification.
- Open Target tests cover absolute-path standardization, symbolic-link resolution, display-path retention, case-insensitive Markdown acceptance, directory acceptance, unsupported inputs, missing targets, and removable targets becoming available again.
- Preview Session tests cover Starting/Running/Stopping transitions, one-session identity, concurrent duplicate requests, revisit ordering, stop, Stop All, quit cleanup, unexpected exit, startup timeout, malformed startup information, end-of-stream, early exit, and the prohibition on default-port fallback.
- Recent Target tests cover creation only after success, deduplication, ordering, active suppression, reappearance after stop, removal, clearing, unavailable state, persistence across relaunch, and independence from Preview Session lifecycle.
- Error tests cover Finder input rejection with automatic panel presentation, process launch failure, startup protocol failure, browser failure with retained session, unexpected exit, unread-error indication, acknowledgement, dismissal, retry, and Settings errors rolling back to actual system state.
- Menu Bar Panel UI tests cover empty, starting, active, recent, unavailable, error, and first-run states; left/right menu bar interaction; revisit/stop/remove/clear flows; confirmations; transient dismissal; drag behavior; bounded scrolling; and settings presentation.
- UI tests run with isolated preferences, temporary Open Targets, a controlled fake preview executable, and a browser adapter that never launches the real browser. They do not modify the real login item, Services preferences, user history, or running GoGrip processes.
- Localization tests verify that English and Simplified Chinese resource sets contain the same keys and that Finder Command, panel, settings, confirmation, and error strings do not mix languages.
- Accessibility tests verify labels, values, actions, keyboard traversal, focus visibility, non-color status, text scaling, Reduce Motion behavior, and light/dark icon rendering at supported menu bar sizes.
- Build-product tests inspect the final application bundle for `LSUIElement`, the macOS Service declaration, absence of Finder Sync, correct minimum system version, universal architectures, bundled preview executable, and consistent version metadata.
- Release verification checks every nested executable's signature, Hardened Runtime, Team identity, notarization ticket, staple, Gatekeeper assessment, and final DMG integrity.
- A clean-install smoke test uses only the final DMG in a clean macOS account or VM. It mounts and installs the application, launches it, verifies no Dock icon, discovers the Finder Command, opens a temporary Markdown file and directory, confirms a real HTTP preview, reopens and resolves a symbolic link without creating another process, stops and quits without orphan processes, and confirms Recent Targets persist without session restoration.
- Existing Go tests remain green to demonstrate that the unchanged backend contract has not regressed.
- The release workflow orders gates so that no DMG is uploaded until all automated and clean-install validation has passed.

## Out of Scope

- Changes to the Go Markdown parser, browser renderer, editing workflow, export behavior, or preview protocol beyond consuming its existing machine-readable startup contract.
- An embedded browser, native Markdown reader, or replacement for the existing browser preview.
- Finder Sync, Finder badges, Finder toolbar items, or a guaranteed top-level Finder context-menu position.
- Opening multiple Open Targets in one command or batch session management.
- Automatic restoration of Preview Sessions after relaunch.
- Idle timeout, browser-close detection, or automatic session reclamation.
- A global keyboard shortcut for opening the Menu Bar Panel.
- An automatic-update framework.
- A Mac App Store build or sandbox-specific product variant.
- Manual language, theme, port, recursion, or session-restoration preferences.
- Rich branding, color artwork, emoji headings, or a resizable Menu Bar Panel.
- Supporting macOS releases earlier than macOS 13.

## Further Notes

- The canonical vocabulary is Open Target, Finder Command, Preview Session, Recent Target, and Menu Bar Panel. New implementation and test names should use these terms rather than the avoided synonyms in the domain glossary.
- The Finder Command follows the macOS Services model. Its menu placement and ordering are platform behavior, not a GoGrip acceptance criterion.
- ADR-0001 requires macOS Services and forbids shipping Finder Sync. ADR-0002 requires a persistent signed and notarized menu bar application plus verified-DMG release gating. ADR-0003 defines Open Target identity and one-session ownership. This specification must not be implemented in a way that contradicts them.
- Existing macOS planning and handover documents describe historical implementation intent and are not authoritative where they conflict with this specification or the ADRs.
- Developer ID signing and notarization require maintainer-provided Apple credentials in the release environment. Lack of credentials may block publication but does not justify weakening release validation.
- V1 is complete only when the final DMG passes the clean-install workflow: trusted installation; no Dock icon; discoverable Finder Command for one Markdown file or directory; automatic application launch; direct browser opening; one Preview Session across repeat and symbolic-link requests; correct Menu Bar Panel management; no orphan process after stop or quit; Recent Target persistence without session restoration; English and Simplified Chinese consistency; and baseline keyboard, VoiceOver, light/dark, Reduce Motion, and text-scaling support.
