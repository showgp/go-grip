import AppKit

/// Production browser seam: the system default browser via NSWorkspace.
/// `true` only means the system accepted the request; it is not proof that a
/// page rendered.
struct WorkspaceBrowserOpening: BrowserOpening {
    @discardableResult
    func open(_ url: URL) -> Bool {
        NSWorkspace.shared.open(url)
    }
}
