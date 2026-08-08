# Keep one Preview Session per Open Target

After path normalization and symbolic-link resolution, an Open Target may have at most one Preview Session. GoGrip retains the user-selected path for display, but reopening any path that identifies the same target revisits its session instead of starting another one, preventing duplicate background work, conflicting addresses, and sessions that the Menu Bar Panel can no longer manage.
