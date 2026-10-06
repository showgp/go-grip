import AppKit
import Foundation
import ServiceManagement
import XCTest

/// The Finder Services entry seam: a request delivered to the production
/// provider while no panel exists must turn every selected file URL into the
/// same production batch the manual entry uses, and the request value must be
/// captured before the selector returns.
///
/// The real Finder cold start, the actual Finder multi-selection handoff, the
/// default browser page and the panel actions are the development app smoke,
/// not this file.
final class FinderServiceProviderTests: XCTestCase {
    private var tempRoot: URL!
    private let factory = ScriptedFactory()
    private let browser = RecordingBrowser()
    private let loginItem = ControlledLoginItem()
    private var recentSuiteName: String!
    private var recentDefaults: UserDefaults!
    private var recentStore: RecentTargetsStore!

    override func setUpWithError() throws {
        tempRoot = FileManager.default.temporaryDirectory
            .appendingPathComponent("gogrip-services-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: tempRoot, withIntermediateDirectories: true)
        recentSuiteName = "gogrip-services-recent-\(UUID().uuidString)"
        recentDefaults = try XCTUnwrap(UserDefaults(suiteName: recentSuiteName))
        recentStore = RecentTargetsStore(defaults: recentDefaults)
    }

    override func tearDownWithError() throws {
        if let tempRoot {
            try? FileManager.default.removeItem(at: tempRoot)
        }
        if let recentSuiteName {
            recentDefaults.removePersistentDomain(forName: recentSuiteName)
        }
    }

    @MainActor
    func testServiceRequestOpensEverySelectedTargetAsOneManageableBatch() async throws {
        let folder = try makeDirectory("文档 资料")
        let other = try makeDirectory("other")
        let article = try makeMarkdown("说明 文章.MD")
        let pasteboard = makePasteboard()
        pasteboard.writeObjects([folder as NSURL, article as NSURL, other as NSURL])

        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)
        let provider = FinderServiceProvider(model: model)

        var error: NSString?
        provider.openWithGoGrip(pasteboard, userData: nil, error: &error)
        // The selector captured the request values: the request ending right
        // after it returned cannot change the outcome.
        pasteboard.clearContents()

        XCTAssertNil(error, "a readable request is not an immediate error")
        await waitUntil("every selected target has its own running session and browser request") {
            coordinator.sessions.count == 3
                && coordinator.sessions.allSatisfy { $0.phase == .running }
                && browser.requests.count == 3
        }
        XCTAssertEqual(
            Set(coordinator.sessions.map(\.displayPath)),
            Set([folder.path, article.path, other.path]),
            "all three selected targets become manageable sessions, not only the first"
        )
        XCTAssertEqual(browser.requests.count, 3, "every running target requests the default browser once")
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
    }

    @MainActor
    func testUnreadableServiceRequestReportsBeforeReturnAndStartsNothing() async throws {
        let pasteboard = makePasteboard()
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)
        let provider = FinderServiceProvider(model: model)

        var error: NSString?
        provider.openWithGoGrip(pasteboard, userData: nil, error: &error)

        XCTAssertNotNil(error, "an empty request is reported before the selector returns")
        XCTAssertTrue(coordinator.sessions.isEmpty)
        XCTAssertEqual(factory.count, 0)
        XCTAssertTrue(browser.requests.isEmpty)
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
    }

    private func makePasteboard() -> NSPasteboard {
        let pasteboard = NSPasteboard(name: NSPasteboard.Name("gogrip-services-\(UUID().uuidString)"))
        pasteboard.clearContents()
        return pasteboard
    }

    private func makeDirectory(_ name: String) throws -> URL {
        let url = tempRoot.appendingPathComponent(name)
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        return url
    }

    private func makeMarkdown(_ name: String) throws -> URL {
        let url = tempRoot.appendingPathComponent(name)
        try "# \(name)\n".write(to: url, atomically: true, encoding: .utf8)
        return url
    }

    @MainActor
    private func waitUntil(
        _ description: String,
        timeout: TimeInterval = 2,
        line: UInt = #line,
        _ condition: () -> Bool
    ) async {
        let deadline = Date().addingTimeInterval(timeout)
        while !condition() {
            if Date() >= deadline {
                XCTFail("timed out waiting for \(description)", line: line)
                return
            }
            try? await Task.sleep(nanoseconds: 2_000_000)
        }
    }
}

/// Scripted launch seam: every created launch completes immediately with its
/// own verified URL, so a request runs to completion without parking.
private final class ScriptedFactory: ManagedProcessFactory, @unchecked Sendable {
    private let lock = NSLock()
    private var processes: [ScriptedProcess] = []

    var count: Int {
        lock.lock()
        defer { lock.unlock() }
        return processes.count
    }

    func makeProcess(
        generation: String,
        target: URL,
        mode: ManagedProcess.TargetMode,
        events: @escaping (ManagedEvent) -> Void
    ) -> ManagedProcessLaunching {
        lock.lock()
        let process = ScriptedProcess(generation: generation, target: target, index: processes.count)
        processes.append(process)
        lock.unlock()
        return process
    }
}

/// One scripted owned child: a verified startup on the first `start()`.
private final class ScriptedProcess: ManagedProcessLaunching, @unchecked Sendable {
    let generation: String
    let target: URL
    private let index: Int
    private let lock = NSLock()
    private var running = false

    init(generation: String, target: URL, index: Int) {
        self.generation = generation
        self.target = target
        self.index = index
    }

    var isRunning: Bool {
        lock.lock()
        defer { lock.unlock() }
        return running
    }

    func start() async -> Result<ManagedLaunchSuccess, ManagedLaunchFailure> {
        lock.lock()
        running = true
        lock.unlock()
        return .success(ManagedLaunchSuccess(
            generation: generation,
            url: URL(string: "http://127.0.0.1:\(49_300 + index)/")!,
            reload: ManagedReloadSnapshot(state: "active", reason: nil),
            pid: Int32(700 + index)
        ))
    }

    func stop() async -> ManagedStopOutcome {
        lock.lock()
        running = false
        lock.unlock()
        return ManagedStopOutcome(generation: generation, result: .exited(exitCode: 0, reason: .exit))
    }
}

/// Controlled browser seam: records every requested URL and accepts it.
private final class RecordingBrowser: BrowserOpening, @unchecked Sendable {
    private let lock = NSLock()
    private var recorded: [URL] = []

    var requests: [URL] {
        lock.lock()
        defer { lock.unlock() }
        return recorded
    }

    @discardableResult
    func open(_ url: URL) -> Bool {
        lock.lock()
        recorded.append(url)
        lock.unlock()
        return true
    }
}
