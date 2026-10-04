import Foundation
import XCTest

/// Real-Go integration for the production Foundation launch path: actual
/// managed children, actual loopback URLs, real content and real exits.
///
/// `make macos-test` builds the tool and stages a copy into an isolated
/// temporary bundle on the system volume before running the tests. xctest
/// processes on a checkout stored on a removable volume cannot open files
/// from that volume (open blocks in dyld), and xcodebuild does not forward
/// custom environment variables, so the staged path is fixed. Without it the
/// repo build under macos/.build/test is used.
final class ManagedProcessGoIntegrationTests: XCTestCase {
    private static let stagedTool = URL(
        fileURLWithPath: "/tmp/gogrip-managed-test-tool/GoGrip.app/Contents/MacOS/go-grip"
    )

    private var tempDirs: [URL] = []

    private var repositoryRoot: URL {
        URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent() // GoGripTests
            .deletingLastPathComponent() // macos
            .deletingLastPathComponent() // repository root
    }

    private var goBinary: URL {
        if FileManager.default.isExecutableFile(atPath: Self.stagedTool.path) {
            return Self.stagedTool
        }
        return repositoryRoot.appendingPathComponent("macos/.build/test/go-grip")
    }

    private var goBinaryExists: Bool {
        FileManager.default.isExecutableFile(atPath: goBinary.path)
    }

    override func tearDown() {
        for dir in tempDirs { try? FileManager.default.removeItem(at: dir) }
        tempDirs = []
        super.tearDown()
    }

    // MARK: - Fixtures and helpers

    private func makeTargetsDirectory() throws -> URL {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("gogrip-go-targets-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        tempDirs.append(url)
        return url
    }

    @discardableResult
    private func write(_ contents: String, to url: URL) throws -> URL {
        try FileManager.default.createDirectory(at: url.deletingLastPathComponent(), withIntermediateDirectories: true)
        try contents.write(to: url, atomically: true, encoding: .utf8)
        return url
    }

    private func start(
        target: URL,
        mode: ManagedProcess.TargetMode,
        generation: String = UUID().uuidString
    ) async -> (ManagedProcess, Result<ManagedLaunchSuccess, ManagedLaunchFailure>) {
        let adapter = ManagedProcess(
            executableURL: goBinary,
            generation: generation,
            target: target,
            mode: mode,
            events: { _ in }
        )
        let result = await adapter.start()
        return (adapter, result)
    }

    private func get(_ url: URL) async throws -> (status: Int, body: String) {
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        let (data, response) = try await URLSession.shared.data(for: request)
        let http = try XCTUnwrap(response as? HTTPURLResponse)
        return (http.statusCode, String(decoding: data, as: UTF8.self))
    }

    @discardableResult
    private func waitUntilGone(_ pid: Int32?, timeout: TimeInterval = 3) -> Bool {
        guard let pid else { return true }
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline {
            if kill(pid, 0) == -1 && errno == ESRCH { return true }
            usleep(20_000)
        }
        return kill(pid, 0) == -1 && errno == ESRCH
    }

    /// FIFO descriptors held by this test process, counted through our own
    /// `/dev/fd` entries. kqueue descriptors also report as FIFOs, but a
    /// terminated Process adds none.
    private func pipeDescriptorCount() -> Int {
        let names = (try? FileManager.default.contentsOfDirectory(atPath: "/dev/fd")) ?? []
        guard !names.isEmpty else { return .max }
        var count = 0
        for name in names {
            guard let descriptor = Int32(name) else { continue }
            var info = stat()
            guard fstat(descriptor, &info) == 0 else { continue }
            if info.st_mode & mode_t(S_IFMT) == mode_t(S_IFIFO) { count += 1 }
        }
        return count
    }

    private func waitForPipeDescriptors(atMost limit: Int, timeout: TimeInterval = 3) -> Bool {
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline {
            if pipeDescriptorCount() <= limit { return true }
            usleep(20_000)
        }
        return pipeDescriptorCount() <= limit
    }

    // MARK: - Ready URL and real content

    func testDirectoryTargetReachesRealReadyURLAndNestedContent() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let root = try makeTargetsDirectory()
        try write("# Root\n", to: root.appendingPathComponent("README.md"))
        try write("# Nested\n", to: root.appendingPathComponent("guide/deep/nested.md"))
        let (adapter, result) = await start(target: root, mode: .directory)
        addTeardownBlock { _ = await adapter.stop() }

        guard case .success(let success) = result else {
            return XCTFail("expected success, got \(String(describing: result))")
        }
        XCTAssertEqual(success.url.host, "127.0.0.1", "the hosted preview must be a real loopback URL")
        let port = try XCTUnwrap(success.url.port)
        XCTAssertGreaterThan(port, 0)
        XCTAssertEqual(success.pid, adapter.pid)
        XCTAssertEqual(kill(success.pid, 0), 0, "the validated child must still be running")

        let rootPage = try await get(success.url)
        XCTAssertEqual(rootPage.status, 200)

        let nested = URL(string: "http://127.0.0.1:\(port)/guide/deep/nested.md")!
        let nestedPage = try await get(nested)
        XCTAssertEqual(nestedPage.status, 200)
        XCTAssertTrue(nestedPage.body.contains("Nested"), "nested document content must be served")

        let outcome = await adapter.stop()
        guard case .exited(let status, _) = outcome.result else {
            return XCTFail("expected normal exit, got \(String(describing: outcome))")
        }
        XCTAssertEqual(status, 0)
        XCTAssertTrue(waitUntilGone(success.pid), "stop must confirm the real exit")

        var unreachable = false
        do {
            _ = try await get(success.url)
        } catch {
            unreachable = true
        }
        XCTAssertTrue(unreachable, "the released preview port must no longer serve")
    }

    func testSingleFileWithSpacesAndNonASCIIStaysOnThatFile() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let root = try makeTargetsDirectory()
        let file = root.appendingPathComponent("中文 说明.md")
        try write("# 唯一标记\n", to: file)
        try write("# Other\n", to: root.appendingPathComponent("other.md"))
        let (adapter, result) = await start(target: file, mode: .file)
        addTeardownBlock { _ = await adapter.stop() }

        guard case .success(let success) = result else {
            return XCTFail("expected success, got \(String(describing: result))")
        }
        XCTAssertEqual(success.url.host, "127.0.0.1")
        XCTAssertEqual(
            success.url.path.removingPercentEncoding,
            "/中文 说明.md",
            "the URL must point at the selected file, not its parent directory"
        )
        let page = try await get(success.url)
        XCTAssertEqual(page.status, 200)
        XCTAssertTrue(page.body.contains("唯一标记"), "the selected file's content must be served")

        let sibling = URL(string: "http://127.0.0.1:\(try XCTUnwrap(success.url.port))/other.md")!
        let siblingPage = try await get(sibling)
        XCTAssertEqual(siblingPage.status, 404, "single-file mode must not fall back to the parent directory")

        _ = await adapter.stop()
    }

    func testEmptyDirectoryKeepsSession() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let root = try makeTargetsDirectory()
        let (adapter, result) = await start(target: root, mode: .directory)
        addTeardownBlock { _ = await adapter.stop() }

        guard case .success(let success) = result else {
            return XCTFail("expected success, got \(String(describing: result))")
        }
        let page = try await get(success.url)
        XCTAssertEqual(page.status, 200, "an accessible empty directory stays a session")
        XCTAssertTrue(
            page.body.contains("No Markdown files"),
            "the empty-state page must be served, not a generic success page"
        )
        _ = await adapter.stop()
    }

    func testNonexistentTargetFailsWithGoFatalReason() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let root = try makeTargetsDirectory()
        let missing = root.appendingPathComponent("missing")
        let (adapter, result) = await start(target: missing, mode: .directory)

        guard case .failure(let failure) = result else {
            return XCTFail("expected failure, got \(String(describing: result))")
        }
        guard case .fatal(let code, let message) = failure.reason else {
            return XCTFail("expected fatal, got \(failure.reason)")
        }
        XCTAssertEqual(code, "target-unavailable")
        XCTAssertFalse(message.isEmpty)
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the real Go child running")
    }

    // MARK: - Stopping and isolation

    func testStopOneOfTwoSessionsExitsOnlyThatOne() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let first = try makeTargetsDirectory()
        let second = try makeTargetsDirectory()
        try write("# First\n", to: first.appendingPathComponent("one.md"))
        try write("# Second\n", to: second.appendingPathComponent("two.md"))
        let (adapterA, resultA) = await start(target: first, mode: .directory)
        addTeardownBlock { _ = await adapterA.stop() }
        let (adapterB, resultB) = await start(target: second, mode: .directory)
        addTeardownBlock { _ = await adapterB.stop() }

        guard case .success(let successA) = resultA, case .success(let successB) = resultB else {
            return XCTFail("expected both sessions to start")
        }
        let pageA = try await get(successA.url)
        XCTAssertEqual(pageA.status, 200)
        let pageB = try await get(successB.url)
        XCTAssertEqual(pageB.status, 200)

        let outcomeA = await adapterA.stop()
        guard case .exited(let statusA, _) = outcomeA.result else {
            return XCTFail("expected normal exit for A, got \(String(describing: outcomeA))")
        }
        XCTAssertEqual(statusA, 0)
        XCTAssertTrue(
            waitUntilGone(successA.pid),
            "A must really exit: a sibling that inherited A's writer would keep it alive"
        )

        let stillServing = try await get(successB.url)
        XCTAssertEqual(stillServing.status, 200, "B must keep serving after A stops")
        XCTAssertEqual(kill(successB.pid, 0), 0, "B's real Go child must still be running")

        _ = await adapterB.stop()
    }

    func testStartupFailureReleasesPipeDescriptors() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let root = try makeTargetsDirectory()
        let missing = root.appendingPathComponent("gone")

        // Warm-up failure: one-time runtime descriptors settle before the
        // baseline, and its own release is the first sample.
        let beforeWarmup = pipeDescriptorCount()
        let (warmup, warmupResult) = await start(target: missing, mode: .directory)
        guard case .failure = warmupResult else {
            return XCTFail("expected the first target to fail")
        }
        XCTAssertTrue(waitUntilGone(warmup.pid))
        XCTAssertTrue(
            waitForPipeDescriptors(atMost: beforeWarmup + 2),
            "a failed launch must release its ownership/feedback/diagnostics descriptors"
        )
        let baseline = pipeDescriptorCount()

        // Retain every failed instance: the descriptors must be released even
        // while the adapters stay alive holding their real exit results, so
        // this cannot pass through deallocation.
        var retained: [ManagedProcess] = []
        for _ in 0..<6 {
            let (adapter, result) = await start(target: missing, mode: .directory)
            guard case .failure = result else {
                return XCTFail("expected the target to keep failing")
            }
            XCTAssertTrue(waitUntilGone(adapter.pid))
            retained.append(adapter)
        }
        XCTAssertTrue(
            waitForPipeDescriptors(atMost: baseline + 2),
            "failed launches must not accumulate pipe descriptors (retained: \(retained.count))"
        )

        // A later real session must still start and serve after the failures;
        // no failed launch's writer may be held against it.
        try write("# Later\n", to: root.appendingPathComponent("later.md"))
        let (later, laterResult) = await start(target: root, mode: .directory)
        addTeardownBlock { _ = await later.stop() }
        guard case .success(let success) = laterResult else {
            return XCTFail("expected the later session to start, got \(String(describing: laterResult))")
        }
        let page = try await get(success.url)
        XCTAssertEqual(page.status, 200)

        let outcome = await later.stop()
        guard case .exited(let status, _) = outcome.result else {
            return XCTFail("expected normal exit, got \(String(describing: outcome))")
        }
        XCTAssertEqual(status, 0)
    }

    func testIndependentCLIKeepsServingAfterSessionStop() async throws {
        try XCTSkipUnless(goBinaryExists, "go-grip test binary missing; run make macos-test")
        let cliRoot = try makeTargetsDirectory()
        try write("# CLI target\n", to: cliRoot.appendingPathComponent("cli.md"))
        let sessionRoot = try makeTargetsDirectory()
        try write("# Session target\n", to: sessionRoot.appendingPathComponent("session.md"))

        // Standalone CLI: no ownership stdin, JSON on stdout, existing network policy.
        let cli = Process()
        cli.executableURL = goBinary
        cli.arguments = ["--json", "--browser=false", "--no-reload", "--", cliRoot.path]
        let cliOutput = Pipe()
        cli.standardOutput = cliOutput
        cli.standardError = FileHandle.nullDevice
        addTeardownBlock {
            if cli.isRunning {
                kill(cli.processIdentifier, SIGKILL)
            }
            cli.waitUntilExit()
        }
        try cli.run()
        let jsonLine = try XCTUnwrap(readJSONLine(from: cliOutput.fileHandleForReading, timeout: 10))
        let json = try XCTUnwrap(
            try JSONSerialization.jsonObject(with: Data(jsonLine.utf8)) as? [String: Any]
        )
        let cliURL = try XCTUnwrap((json["url"] as? String).flatMap(URL.init(string:)))
        let cliStatus = try await get(cliURL).status
        XCTAssertEqual(cliStatus, 200)

        let (adapter, result) = await start(target: sessionRoot, mode: .directory)
        guard case .success = result else {
            return XCTFail("expected the managed session to start")
        }
        _ = await adapter.stop()

        XCTAssertTrue(cli.isRunning, "stopping a managed session must not terminate the standalone CLI")
        let after = try await get(cliURL)
        XCTAssertEqual(after.status, 200, "the standalone CLI must keep serving its own target")
    }

    private func readJSONLine(from handle: FileHandle, timeout: TimeInterval) -> String? {
        let deadline = Date().addingTimeInterval(timeout)
        var buffer = Data()
        while Date() < deadline {
            let chunk = handle.availableData
            if chunk.isEmpty { return nil }
            buffer.append(chunk)
            if let newline = buffer.firstIndex(of: UInt8(ascii: "\n")) {
                return String(decoding: buffer[..<newline], as: UTF8.self)
            }
        }
        return nil
    }
}
