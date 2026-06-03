import Foundation

struct Storage {
    private let key = "go-grip-history"
    private let maxEntries = 50

    func load() -> [HistoryEntry] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let entries = try? JSONDecoder().decode([HistoryEntry].self, from: data) else {
            return []
        }
        return entries
    }

    func save(_ entries: [HistoryEntry]) {
        guard let data = try? JSONEncoder().encode(entries) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }

    func addToHistory(path: String) {
        var entries = load()
        var isDir: ObjCBool = false
        FileManager.default.fileExists(atPath: path, isDirectory: &isDir)
        let displayName = URL(fileURLWithPath: path).lastPathComponent

        if let index = entries.firstIndex(where: { $0.path == path }) {
            var existing = entries.remove(at: index)
            existing.lastOpened = Date()
            existing.accessCount += 1
            entries.insert(existing, at: 0)
        } else {
            let entry = HistoryEntry(
                id: UUID(),
                path: path,
                displayName: displayName,
                isDirectory: isDir.boolValue,
                lastOpened: Date(),
                accessCount: 1
            )
            entries.insert(entry, at: 0)
        }

        if entries.count > maxEntries {
            entries = Array(entries.prefix(maxEntries))
        }

        save(entries)
    }

    func removeFromHistory(id: UUID) {
        var entries = load()
        entries.removeAll { $0.id == id }
        save(entries)
    }

    func clearHistory() {
        UserDefaults.standard.removeObject(forKey: key)
    }
}
