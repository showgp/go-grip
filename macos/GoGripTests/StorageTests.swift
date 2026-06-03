import XCTest
@testable import GoGrip

final class StorageTests: XCTestCase {
    var storage: Storage!

    override func setUp() {
        super.setUp()
        storage = Storage()
        storage.clearHistory()
    }

    override func tearDown() {
        storage.clearHistory()
        super.tearDown()
    }

    func testLoadEmpty() {
        let entries = storage.load()
        XCTAssertTrue(entries.isEmpty)
    }

    func testAddAndLoad() {
        storage.addToHistory(path: "/tmp")
        let entries = storage.load()
        XCTAssertEqual(entries.count, 1)
        XCTAssertEqual(entries.first?.path, "/tmp")
    }

    func testAddToHistoryDedup() {
        storage.addToHistory(path: "/tmp")
        storage.addToHistory(path: "/tmp")
        let entries = storage.load()
        XCTAssertEqual(entries.count, 1)
        XCTAssertEqual(entries.first?.accessCount, 2)
    }

    func testAddToHistoryMaxEntries() {
        for i in 0..<51 {
            storage.addToHistory(path: "/tmp/test_\(i)")
        }
        let entries = storage.load()
        XCTAssertEqual(entries.count, 50)
        let paths = entries.map(\.path)
        XCTAssertFalse(paths.contains("/tmp/test_0"))
    }

    func testRemoveFromHistory() {
        storage.addToHistory(path: "/tmp")
        var entries = storage.load()
        guard let entry = entries.first else {
            XCTFail("Expected at least one entry")
            return
        }
        storage.removeFromHistory(id: entry.id)
        entries = storage.load()
        XCTAssertTrue(entries.isEmpty)
    }

    func testClearHistory() {
        storage.addToHistory(path: "/tmp/a")
        storage.addToHistory(path: "/tmp/b")
        storage.clearHistory()
        let entries = storage.load()
        XCTAssertTrue(entries.isEmpty)
    }

    func testHistoryOrdering() {
        storage.addToHistory(path: "/tmp/a")
        storage.addToHistory(path: "/tmp/b")
        storage.addToHistory(path: "/tmp/c")
        let entries = storage.load()
        XCTAssertEqual(entries.count, 3)
        XCTAssertEqual(entries[0].path, "/tmp/c")
        XCTAssertEqual(entries[1].path, "/tmp/b")
        XCTAssertEqual(entries[2].path, "/tmp/a")
    }

    // MARK: - M3: Storage boundary tests

    func testUnicodePath() {
        let path = "中文/目录/文件.md"
        storage.addToHistory(path: path)
        let entries = storage.load()
        XCTAssertEqual(entries.count, 1)
        XCTAssertEqual(entries.first?.path, path)
        XCTAssertEqual(entries.first?.displayName, "文件.md")
    }

    func testEmptyPath() {
        storage.addToHistory(path: "")
        let entries = storage.load()
        XCTAssertEqual(entries.count, 1)
    }

    func testCorruptedData() {
        UserDefaults.standard.set("not valid json", forKey: "go-grip-history")
        let entries = storage.load()
        XCTAssertTrue(entries.isEmpty)
    }
}
