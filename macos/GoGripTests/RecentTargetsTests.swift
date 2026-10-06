import XCTest
@testable import GoGrip

/// Recent-target persistence at its consumer seam: an isolated, disposable
/// UserDefaults suite, so the real user domain is never read or written. The
/// panel and the model consume these exact identities and orderings.
final class RecentTargetsTests: XCTestCase {
    private var suiteName: String!
    private var defaults: UserDefaults!

    override func setUpWithError() throws {
        suiteName = "gogrip-recent-tests-\(UUID().uuidString)"
        defaults = try XCTUnwrap(UserDefaults(suiteName: suiteName))
    }

    override func tearDownWithError() throws {
        defaults.removePersistentDomain(forName: suiteName)
        defaults = nil
        suiteName = nil
    }

    func testKeepingTwentyOneDistinctTargetsEvictsTheOldestAndReopeningAnOlderTargetRefreshesTheOrder() {
        let store = RecentTargetsStore(defaults: defaults)

        for index in 0...20 {
            store.record(
                identity: "/targets/\(index)",
                displayPath: "/selected/\(index)",
                mode: index == 3 ? .file : .directory,
                at: Date(timeIntervalSinceReferenceDate: Double(index))
            )
        }

        XCTAssertEqual(
            store.load().map(\.identity),
            (1...20).reversed().map { "/targets/\($0)" },
            "the oldest of 21 distinct identities must be evicted and the rest kept most-recent-first"
        )
        XCTAssertFalse(store.load().contains { $0.identity == "/targets/0" })

        let refreshed = store.record(
            identity: "/targets/5",
            displayPath: "/selected/5-again",
            mode: .directory,
            at: Date(timeIntervalSinceReferenceDate: 100)
        )

        XCTAssertEqual(
            refreshed.map(\.identity),
            ["/targets/5"] + (6...20).reversed().map { "/targets/\($0)" } + (1...4).reversed().map { "/targets/\($0)" },
            "reopening an older identity must move that exact entry to the front without duplicating it"
        )
        XCTAssertEqual(refreshed.count, RecentTargetsStore.maximumCount)
        XCTAssertEqual(refreshed.filter { $0.identity == "/targets/5" }.count, 1)
        XCTAssertEqual(
            refreshed.first?.displayPath,
            "/selected/5-again",
            "the refreshed record must keep the latest user-selected path"
        )
        XCTAssertEqual(store.load().first?.displayPath, "/selected/5-again")
    }

    func testRecreatingTheStoreRestoresTheSameOrder() {
        let store = RecentTargetsStore(defaults: defaults)
        for index in 0..<3 {
            store.record(
                identity: "/targets/\(index)",
                displayPath: "/selected/\(index)",
                mode: .directory,
                at: Date(timeIntervalSinceReferenceDate: Double(index))
            )
        }

        let recreated = RecentTargetsStore(defaults: defaults)

        XCTAssertEqual(
            recreated.load().map(\.identity),
            ["/targets/2", "/targets/1", "/targets/0"],
            "a recreated store must load the same position and order it persisted"
        )
        XCTAssertEqual(recreated.load().map(\.displayPath), ["/selected/2", "/selected/1", "/selected/0"])
    }

    func testClearRemovesOnlyTheRecentKeyAndKeepsOtherStoredValues() {
        let store = RecentTargetsStore(defaults: defaults)
        store.record(identity: "/targets/a", displayPath: "/selected/a", mode: .directory)
        store.record(identity: "/targets/b", displayPath: "/selected/b", mode: .file)
        defaults.set("sentinel", forKey: "gogrip-recent-tests-sentinel")
        defaults.set(Data("legacy".utf8), forKey: "go-grip-history")

        store.clear()

        XCTAssertTrue(store.load().isEmpty)
        XCTAssertTrue(RecentTargetsStore(defaults: defaults).load().isEmpty, "a later launch must stay empty")
        XCTAssertEqual(defaults.string(forKey: "gogrip-recent-tests-sentinel"), "sentinel")
        XCTAssertEqual(
            defaults.data(forKey: "go-grip-history"),
            Data("legacy".utf8),
            "clearing recents must not touch the legacy history key or any other preference"
        )
    }
}
