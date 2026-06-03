import Foundation

struct HistoryEntry: Codable, Identifiable {
    let id: UUID
    let path: String
    let displayName: String
    let isDirectory: Bool
    var lastOpened: Date
    var accessCount: Int
}
