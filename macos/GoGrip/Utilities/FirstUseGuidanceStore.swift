import Foundation

/// Minimal first-use state for the Finder/Services guidance: one dedicated
/// flag with no other behavior attached. It is independent from the recent
/// targets key and from the legacy `go-grip-history` data, which are never
/// read, migrated, written or cleared here.
struct FirstUseGuidanceStore {
    /// Independent key; neither the recent targets key nor the legacy history
    /// key is touched.
    static let defaultKey = "go-grip-first-use-guidance-shown"

    private let defaults: UserDefaults

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    /// True while the guidance has never been presented automatically. A user
    /// who only revisits the guidance through the panel does not change this.
    func shouldPresent() -> Bool {
        defaults.object(forKey: Self.defaultKey) == nil
    }

    /// Records that the first-use guidance was presented so later launches
    /// stay silent until the user opens the guidance explicitly.
    func markPresented() {
        defaults.set(true, forKey: Self.defaultKey)
    }
}
