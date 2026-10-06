import Foundation

/// One persisted recent target: the normalized identity deduplicates and
/// orders the list, while the user-selected display path and kind stay for
/// display. It is display data only — never a running fact, a bookmark or a
/// promise that the target is still reachable.
struct RecentTarget: Codable, Equatable, Identifiable {
    var id: String { identity }
    let identity: String
    let displayPath: String
    let mode: ManagedProcess.TargetMode
    let lastUsed: Date
}
