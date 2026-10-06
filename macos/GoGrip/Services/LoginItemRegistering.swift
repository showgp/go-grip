import ServiceManagement

/// Reads and changes the one login-start registration that belongs to this
/// app: `SMAppService.mainApp` of the running bundle. The real system status
/// is the only login-start truth the panel consumes; no desired value is
/// persisted, and nothing here registers or unregisters without an explicit
/// user action.
protocol LoginItemRegistering: AnyObject {
    /// The current registration as the system reports it.
    var status: SMAppService.Status { get }
    /// Registers the main app to launch at login. Throws when the system
    /// refuses; the caller re-reads `status` afterwards instead of assuming
    /// the request took effect.
    func register() throws
    /// Removes this app's login-start registration.
    func unregister() throws
    /// Opens the system Login Items settings so the user can approve or review
    /// the item; only ever called from an explicit user action.
    func openLoginItemsSettings()
}

/// Production adapter over the main app's own service. It deliberately does
/// not wrap `SMAppService` in a scheduler, retry loop or persisted state: the
/// SDK's synchronous completion is the boundary, and the next `status` read
/// is the result.
final class MainAppLoginItem: LoginItemRegistering {
    var status: SMAppService.Status {
        SMAppService.mainApp.status
    }

    func register() throws {
        try SMAppService.mainApp.register()
    }

    func unregister() throws {
        try SMAppService.mainApp.unregister()
    }

    func openLoginItemsSettings() {
        SMAppService.openSystemSettingsLoginItems()
    }
}
