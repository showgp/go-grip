# GoGrip

GoGrip previews local Markdown through a browser and provides macOS entry points for opening and managing previews.

## Language

**Open Target**:
A local Markdown file or directory that the user asks GoGrip to preview. A symbolic link and its resolved destination identify the same Open Target.
_Avoid_: Path, item, document

**Finder Command**:
The Finder action that asks GoGrip to open one selected Open Target.
_Avoid_: Finder extension, right-click handler

**Preview Session**:
An active preview of one Open Target that the user can revisit or stop.
_Avoid_: Instance, server, task

**Recent Target**:
A persisted record of a successfully opened Open Target, whether or not it currently has a Preview Session.
_Avoid_: History item, recent file

**Menu Bar Panel**:
The macOS control surface for opening Open Targets and managing Recent Targets and Preview Sessions.
_Avoid_: Status bar window, popover UI
