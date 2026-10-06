import Foundation

/// Versioned, bounded persistence for recent targets: at most 20 distinct
/// normalized identities, most recently used first. It persists the location
/// and the display information only; loading decodes the list without
/// probing, visiting or starting any target. The independent key replaces the
/// example app's `go-grip-history` (50 raw-path entries), whose data is
/// neither read nor migrated nor deleted.
struct RecentTargetsStore {
    /// Distinct normalized identities kept across launches.
    static let maximumCount = 20
    /// Independent, versioned key. The legacy key is not touched.
    static let defaultKey = "go-grip-recent-targets-v1"

    private let defaults: UserDefaults

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    /// The persisted targets, most recently used first. A missing or
    /// unreadable value is an empty list.
    func load() -> [RecentTarget] {
        guard let data = defaults.data(forKey: Self.defaultKey),
              let targets = try? JSONDecoder().decode([RecentTarget].self, from: data) else {
            return []
        }
        return Array(targets.prefix(Self.maximumCount))
    }

    /// Records one successfully opened target: an existing identity is
    /// refreshed with the latest user-selected path and kind and moved to the
    /// front instead of being duplicated; the oldest entry beyond the limit is
    /// dropped. Returns the new published list.
    @discardableResult
    func record(
        identity: String,
        displayPath: String,
        mode: ManagedProcess.TargetMode,
        at date: Date = Date()
    ) -> [RecentTarget] {
        var targets = load().filter { $0.identity != identity }
        targets.insert(
            RecentTarget(identity: identity, displayPath: displayPath, mode: mode, lastUsed: date),
            at: 0
        )
        targets = Array(targets.prefix(Self.maximumCount))
        if let data = try? JSONEncoder().encode(targets) {
            defaults.set(data, forKey: Self.defaultKey)
        }
        return targets
    }

    /// Clears only this key. Running sessions, their management entries and
    /// every other stored preference stay untouched, and later launches stay
    /// empty until the user explicitly opens a target again.
    func clear() {
        defaults.removeObject(forKey: Self.defaultKey)
    }
}
