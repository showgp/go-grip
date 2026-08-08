# Use macOS Services for the Finder Command

The Finder Command will use macOS Services rather than Finder Sync, and the macOS product will not ship a Finder Sync extension. A Service fits the general-purpose “Open with GoGrip” action and avoids Finder Sync's synchronization-specific model, monitored-directory boundary, and separate extension enablement; we accept that macOS controls the command's position and ordering in Finder's Services menus.
