import Foundation
import XCTest

/// Adapter seam: parameter-array launch, v1 feedback validation, failure
/// cleanup and stopping. Controlled fake tools script v1 feedback real Go
/// cannot produce; ownership and success evidence comes from the real-Go
/// integration tests instead.
final class ManagedProcessTests: XCTestCase {
    private var tempDirs: [URL] = []

    override func tearDown() {
        StubURLProtocol.handler = nil
        for dir in tempDirs { try? FileManager.default.removeItem(at: dir) }
        tempDirs = []
        super.tearDown()
    }

    // MARK: - Helpers

    private func makeTempDir() throws -> URL {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("gogrip-managed-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        tempDirs.append(url)
        return url
    }

    private func makeFakeTool(_ script: String) throws -> URL {
        let dir = try makeTempDir()
        let url = dir.appendingPathComponent("fake-go-grip")
        try ("#!/bin/sh\n" + script + "\n").write(to: url, atomically: true, encoding: .utf8)
        try FileManager.default.setAttributes([.posixPermissions: 0o755], ofItemAtPath: url.path)
        return url
    }

    private let testGeneration = "gen-host-1"
    private let quickTiming = ManagedProcessTiming(startupDeadline: 5, terminationGrace: 0.5)

    private func start(
        tool: URL,
        target: URL = URL(fileURLWithPath: "/tmp/gogrip-target"),
        mode: ManagedProcess.TargetMode = .directory,
        timing: ManagedProcessTiming? = nil,
        sessionConfiguration: URLSessionConfiguration? = nil,
        events: @escaping (ManagedEvent) -> Void = { _ in }
    ) async -> (ManagedProcess, Result<ManagedLaunchSuccess, ManagedLaunchFailure>) {
        let adapter = ManagedProcess(
            executableURL: tool,
            generation: testGeneration,
            target: target,
            mode: mode,
            timing: timing ?? quickTiming,
            headSessionConfiguration: sessionConfiguration,
            events: events
        )
        let result = await adapter.start()
        return (adapter, result)
    }

    private func failure(
        _ result: Result<ManagedLaunchSuccess, ManagedLaunchFailure>,
        file: StaticString = #filePath,
        line: UInt = #line
    ) -> ManagedLaunchFailure? {
        switch result {
        case .success(let success):
            XCTFail("expected failure, got success \(success)", file: file, line: line)
            return nil
        case .failure(let failure):
            XCTAssertEqual(failure.generation, testGeneration, file: file, line: line)
            return failure
        }
    }

    private func readyLine(url: String = "http://127.0.0.1:54321/") -> String {
        #"{"version":1,"event":"ready","generation":"\#(testGeneration)","url":"\#(url)","reload":{"state":"active"}}"#
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

    // MARK: - Successful startup through the same validation path

    func testSpawnedEventIsDeliveredBeforeReadinessValidation() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let lock = NSLock()
        var observed: [ManagedEvent] = []
        var spawnedSeen = false
        var headSawSpawned: Bool?
        StubURLProtocol.handler = { request in
            lock.lock()
            headSawSpawned = spawnedSeen
            lock.unlock()
            let response = HTTPURLResponse(
                url: request.url!, statusCode: 204,
                httpVersion: "HTTP/1.1",
                headerFields: ["X-GoGrip-Generation": "gen-host-1"]
            )!
            return (response, nil)
        }
        let (adapter, result) = await start(
            tool: tool,
            sessionConfiguration: StubURLProtocol.configuration()
        ) { event in
            lock.lock()
            observed.append(event)
            if case .spawned = event.kind { spawnedSeen = true }
            lock.unlock()
        }
        addTeardownBlock { _ = await adapter.stop() }

        guard case .success(let success) = result else {
            return XCTFail("expected success, got \(String(describing: result))")
        }
        XCTAssertEqual(success.generation, testGeneration)
        XCTAssertEqual(success.url.absoluteString, "http://127.0.0.1:54321/")
        XCTAssertEqual(success.reload, ManagedReloadSnapshot(state: "active", reason: nil))
        XCTAssertTrue(success.pid > 0)
        lock.lock()
        let events = observed
        let orderingObserved = headSawSpawned
        lock.unlock()
        XCTAssertEqual(events.first?.generation, testGeneration)
        XCTAssertEqual(events.first?.kind, .spawned(pid: success.pid))
        XCTAssertEqual(
            orderingObserved, true,
            "the readiness HEAD must not be issued before the spawned event"
        )
    }

    func testReportsReadinessFailureWhenHeaderGenerationDiffers() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(
            tool: tool,
            sessionConfiguration: StubURLProtocol.configuration(generation: "another-generation", status: 204)
        )

        guard case .readinessFailed(let detail)? = failure(result)?.reason else {
            return XCTFail("expected readinessFailed, got \(String(describing: result))")
        }
        XCTAssertFalse(detail.isEmpty)
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testRefusesReadinessRedirectToAnotherOrigin() async throws {
        // The redirect destination is a second reachable loopback origin (a
        // different port) that really answers 204 with the matching
        // generation, so following the redirect would contact it and succeed
        // absent the origin check.
        let destination = try LocalHTTPServer { _ in
            "HTTP/1.1 204 No Content\r\nX-GoGrip-Generation: gen-host-1\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
        }
        addTeardownBlock { destination.stop() }
        let origin = try LocalHTTPServer { head in
            if head.hasPrefix("HEAD /__gogrip/ready") {
                let location = "http://127.0.0.1:\(destination.port)/__gogrip/ready"
                return "HTTP/1.1 302 Found\r\nLocation: \(location)\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
            }
            return "HTTP/1.1 204 No Content\r\nX-GoGrip-Generation: gen-host-1\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
        }
        addTeardownBlock { origin.stop() }

        // Control: the destination really would answer 204 with the matching
        // generation if the redirect were followed.
        var controlRequest = URLRequest(url: URL(string: "http://127.0.0.1:\(destination.port)/__gogrip/ready")!)
        controlRequest.httpMethod = "HEAD"
        let (_, controlResponse) = try await URLSession.shared.data(for: controlRequest)
        XCTAssertEqual((controlResponse as? HTTPURLResponse)?.statusCode, 204)
        XCTAssertEqual(destination.requests.count, 1, "the control request must reach the destination")

        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:\(origin.port)/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(tool: tool)

        guard case .readinessFailed? = failure(result)?.reason else {
            return XCTFail("expected readinessFailed, got \(String(describing: result))")
        }
        let heads = origin.requests
        XCTAssertEqual(heads.count, 1, "exactly one readiness HEAD may reach the origin")
        XCTAssertTrue(heads.first?.contains("HEAD /__gogrip/ready") ?? false)
        XCTAssertEqual(
            destination.requests.count, 1,
            "the readiness check must not follow the redirect to another origin"
        )
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testReportsReadinessTransportFailure() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        StubURLProtocol.handler = { _ in throw URLError(.cannotConnectToHost) }
        let (adapter, result) = await start(
            tool: tool,
            sessionConfiguration: StubURLProtocol.configuration()
        )

        guard case .readinessFailed? = failure(result)?.reason else {
            return XCTFail("expected readinessFailed, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testReadinessFailsWhenChildDiesDuringConfirmation() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let lock = NSLock()
        var childPID: Int32?
        let process = ManagedProcess(
            executableURL: tool,
            generation: testGeneration,
            target: URL(fileURLWithPath: "/tmp/gogrip-target"),
            mode: .directory,
            timing: quickTiming,
            headSessionConfiguration: StubURLProtocol.configuration()
        ) { event in
            lock.lock()
            if case .spawned(let pid) = event.kind { childPID = pid }
            lock.unlock()
        }
        StubURLProtocol.handler = { request in
            lock.lock()
            let pid = childPID
            lock.unlock()
            if let pid {
                kill(pid, SIGKILL)
                // Wait until the adapter itself has observed the exit, so the
                // late 204 can only race an already-recorded failure.
                let deadline = Date().addingTimeInterval(3)
                while process.isRunning && Date() < deadline { usleep(2_000) }
            }
            let response = HTTPURLResponse(
                url: request.url!, statusCode: 204,
                httpVersion: "HTTP/1.1",
                headerFields: ["X-GoGrip-Generation": "gen-host-1"]
            )!
            return (response, nil)
        }
        let result = await process.start()

        guard case .readinessFailed? = failure(result)?.reason else {
            return XCTFail("expected readinessFailed, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(process.pid))
    }

    func testFailureDuringReadinessConfirmationCannotBecomeSuccess() async throws {
        let trigger = try makeTempDir().appendingPathComponent("trigger")
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        while [ ! -f '\(trigger.path)' ]; do sleep 0.01; done
        printf 'not a v1 frame\\n'
        exec /bin/sleep 300
        """)
        let process = ManagedProcess(
            executableURL: tool,
            generation: testGeneration,
            target: URL(fileURLWithPath: "/tmp/gogrip-target"),
            mode: .directory,
            timing: quickTiming,
            headSessionConfiguration: StubURLProtocol.configuration()
        ) { _ in }
        StubURLProtocol.handler = { request in
            FileManager.default.createFile(atPath: trigger.path, contents: nil)
            // The invalid frame fails startup and terminates the child; only
            // then answer the HEAD, so a success commit would be a real race.
            let deadline = Date().addingTimeInterval(3)
            while process.isRunning && Date() < deadline { usleep(2_000) }
            let response = HTTPURLResponse(
                url: request.url!, statusCode: 204,
                httpVersion: "HTTP/1.1",
                headerFields: ["X-GoGrip-Generation": "gen-host-1"]
            )!
            return (response, nil)
        }
        let result = await process.start()

        guard case .invalidFeedback? = failure(result)?.reason else {
            return XCTFail("expected invalidFeedback, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(process.pid))
    }

    // MARK: - Invalid feedback

    func testReadyURLMustBeHTTPLoopback() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://192.0.2.10:9/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(tool: tool)

        guard case .invalidFeedback? = failure(result)?.reason else {
            return XCTFail("expected invalidFeedback, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testUnsupportedVersionFailsIntegration() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":2,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(tool: tool)

        guard case .invalidFeedback? = failure(result)?.reason else {
            return XCTFail("expected invalidFeedback, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testGenerationMismatchFailsIntegration() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"other","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(tool: tool)

        guard case .invalidFeedback? = failure(result)?.reason else {
            return XCTFail("expected invalidFeedback, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testOversizedResidualFailsIntegration() async throws {
        let tool = try makeFakeTool("""
        dd if=/dev/zero bs=1024 count=70 2>/dev/null | tr '\\0' 'x'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(tool: tool)

        guard case .invalidFeedback? = failure(result)?.reason else {
            return XCTFail("expected invalidFeedback, got \(String(describing: result))")
        }
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    // MARK: - Fatal, early exit, timeout

    func testFatalFrameSurfacesCodeAndCleansUp() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"fatal","generation":"gen-host-1","code":"target-unavailable","message":"resolve target: no such file or directory"}'
        exit 1
        """)
        let (adapter, result) = await start(tool: tool)

        guard case .fatal(let code, let message)? = failure(result)?.reason else {
            return XCTFail("expected fatal, got \(String(describing: result))")
        }
        XCTAssertEqual(code, "target-unavailable")
        XCTAssertEqual(message, "resolve target: no such file or directory")
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testEarlyExitWithoutFeedbackFails() async throws {
        let tool = try makeFakeTool("exit 3")
        let (adapter, result) = await start(tool: tool)

        guard case .exitedEarly(let status)? = failure(result)?.reason else {
            return XCTFail("expected exitedEarly, got \(String(describing: result))")
        }
        XCTAssertEqual(status, 3)
        XCTAssertTrue(waitUntilGone(adapter.pid), "failed startup must not leave the child running")
    }

    func testTimeoutFailsAndStopsOwnedChild() async throws {
        let tool = try makeFakeTool("exec /bin/sleep 300")
        let started = Date()
        let (adapter, result) = await start(
            tool: tool,
            timing: ManagedProcessTiming(startupDeadline: 0.8, terminationGrace: 0.5)
        )

        guard case .timedOut? = failure(result)?.reason else {
            return XCTFail("expected timedOut, got \(String(describing: result))")
        }
        XCTAssertGreaterThanOrEqual(Date().timeIntervalSince(started), 0.8)
        XCTAssertLessThan(Date().timeIntervalSince(started), 3)
        XCTAssertTrue(waitUntilGone(adapter.pid), "timed-out startup must stop the owned child")
    }

    // MARK: - Spawn failure and runtime protocol violations

    func testSpawnFailureReportedWithoutChild() async throws {
        let directory = try makeTempDir() // not executable
        let (adapter, result) = await start(tool: directory)

        guard case .spawnFailed(let detail)? = failure(result)?.reason else {
            return XCTFail("expected spawnFailed, got \(String(describing: result))")
        }
        XCTAssertFalse(detail.isEmpty)
        XCTAssertNil(adapter.pid)
    }

    func testPostStartupProtocolViolationTerminatesOwnedChild() async throws {
        let trigger = try makeTempDir().appendingPathComponent("trigger")
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        while [ ! -f '\(trigger.path)' ]; do sleep 0.01; done
        printf 'not a v1 frame\\n'
        exec /bin/sleep 300
        """)
        let lock = NSLock()
        var observed: [ManagedEvent] = []
        let (adapter, result) = await start(
            tool: tool,
            sessionConfiguration: StubURLProtocol.configuration(generation: testGeneration, status: 204)
        ) { event in
            lock.lock()
            observed.append(event)
            lock.unlock()
        }
        guard case .success(let success) = result else {
            return XCTFail("expected success, got \(String(describing: result))")
        }

        try FileManager.default.createFile(atPath: trigger.path, contents: nil)
        let deadline = Date().addingTimeInterval(5)
        while Date() < deadline {
            lock.lock()
            let reported = observed.contains { if case .protocolViolation = $0.kind { return true }; return false }
            let exited = observed.contains { if case .exited = $0.kind { return true }; return false }
            lock.unlock()
            if reported && exited { break }
            usleep(20_000)
        }
        lock.lock()
        let violations = observed.compactMap { event -> String? in
            if case .protocolViolation(let detail) = event.kind { return detail }
            return nil
        }
        let exits = observed.filter { if case .exited = $0.kind { return true }; return false }
        lock.unlock()

        XCTAssertEqual(violations.count, 1, "a runtime violation must be reported exactly once")
        XCTAssertFalse(violations.first?.isEmpty ?? true)
        XCTAssertEqual(exits.count, 1, "the terminated child must still report its real exit")
        XCTAssertTrue(waitUntilGone(success.pid), "the owned child must not survive a protocol violation")
    }

    // MARK: - Stopping

    func testStopSIGKILLsAfterProductionTerminationGrace() async throws {
        let tool = try makeFakeTool("""
        printf '%s\\n' '{"version":1,"event":"ready","generation":"gen-host-1","url":"http://127.0.0.1:54321/","reload":{"state":"active"}}'
        exec /bin/sleep 300
        """)
        let (adapter, result) = await start(
            tool: tool,
            timing: ManagedProcessTiming(),
            sessionConfiguration: StubURLProtocol.configuration(generation: testGeneration, status: 204)
        )
        guard case .success = result else {
            return XCTFail("expected success, got \(String(describing: result))")
        }
        let pid = adapter.pid
        let started = Date()
        let outcome = await adapter.stop()

        XCTAssertGreaterThanOrEqual(Date().timeIntervalSince(started), 3.8, "SIGKILL must not preempt the 4s grace")
        guard case .killedAfterGrace(let status, let reason) = outcome.result else {
            return XCTFail("expected killedAfterGrace, got \(String(describing: outcome))")
        }
        XCTAssertEqual(status, 9)
        XCTAssertEqual(reason, .uncaughtSignal)
        XCTAssertTrue(waitUntilGone(pid), "SIGKILL fallback must confirm the owned child exited")
    }
}

/// Deterministic coverage for the bounded stderr tail itself. Reader
/// scheduling in a real launch decides which exact bytes a failure snapshot
/// holds, so the cap and newest-suffix semantics are pinned here.
final class ManagedDiagnosticsTailTests: XCTestCase {
    func testKeepsOnlyTheNewestCapacityBytes() {
        var tail = ManagedDiagnosticsTail()
        tail.append(Data("DIAG-BEGIN\n".utf8))
        tail.append(Data(repeating: 0x78, count: 100 * 1024))
        tail.append(Data("\nDIAG-END\n".utf8))

        XCTAssertEqual(tail.text.utf8.count, ManagedDiagnosticsTail.capacity)
        XCTAssertTrue(tail.text.hasSuffix("\nDIAG-END\n"), "the newest bytes must be retained")
        XCTAssertFalse(tail.text.contains("DIAG-BEGIN"), "bytes beyond the cap must be dropped")
    }

    func testKeepsEverythingUnderCapacity() {
        var tail = ManagedDiagnosticsTail()
        tail.append(Data("first\n".utf8))
        tail.append(Data("second\n".utf8))

        XCTAssertEqual(tail.text, "first\nsecond\n")
    }
}

/// URLProtocol stub for deterministic readiness-response shapes (redirect,
/// status, generation header) that a real Go child cannot produce.
final class StubURLProtocol: URLProtocol {
    static var handler: ((URLRequest) throws -> (HTTPURLResponse, Data?))?

    /// Configuration with the stub installed; the caller assigns the handler.
    static func configuration() -> URLSessionConfiguration {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [StubURLProtocol.self]
        return configuration
    }

    static func configuration(generation: String, status: Int) -> URLSessionConfiguration {
        handler = { request in
            let response = HTTPURLResponse(
                url: request.url!, statusCode: status,
                httpVersion: "HTTP/1.1",
                headerFields: ["X-GoGrip-Generation": generation]
            )!
            return (response, nil)
        }
        return configuration()
    }

    override class func canInit(with request: URLRequest) -> Bool { true }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = Self.handler else {
            client?.urlProtocol(self, didFailWithError: URLError(.unsupportedURL))
            return
        }
        do {
            let (response, data) = try handler(request)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            if let data { client?.urlProtocol(self, didLoad: data) }
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}

/// Minimal loopback HTTP server for the redirect regression: the adapter's
/// real HEAD must be answered by a real 302, which a URLProtocol stub cannot
/// deliver through URLSession's redirect machinery.
final class LocalHTTPServer {
    private(set) var port = 0

    private let listenSocket: Int32
    private let lock = NSLock()
    private var recordedRequests: [String] = []
    private var stopped = false

    /// The responder receives the raw request head and returns the raw HTTP
    /// response bytes.
    init(responder: @escaping (String) -> String) throws {
        listenSocket = socket(AF_INET, SOCK_STREAM, 0)
        guard listenSocket >= 0 else {
            throw URLError(.cannotConnectToHost)
        }
        var reuse: Int32 = 1
        setsockopt(listenSocket, SOL_SOCKET, SO_REUSEADDR, &reuse, socklen_t(MemoryLayout<Int32>.size))
        var address = sockaddr_in()
        address.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
        address.sin_family = sa_family_t(AF_INET)
        address.sin_port = 0
        inet_pton(AF_INET, "127.0.0.1", &address.sin_addr)
        let bound = withUnsafePointer(to: &address) { pointer in
            pointer.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(listenSocket, $0, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        guard bound == 0, listen(listenSocket, 8) == 0 else {
            close(listenSocket)
            throw URLError(.cannotConnectToHost)
        }
        var assigned = sockaddr_in()
        var length = socklen_t(MemoryLayout<sockaddr_in>.size)
        _ = withUnsafeMutablePointer(to: &assigned) { pointer in
            pointer.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                getsockname(listenSocket, $0, &length)
            }
        }
        port = Int(UInt16(bigEndian: assigned.sin_port))
        DispatchQueue(label: "com.showgp.go-grip.test-http").async { [weak self] in
            self?.serve(responder: responder)
        }
    }

    deinit {
        stop()
    }

    var requests: [String] {
        lock.lock()
        defer { lock.unlock() }
        return recordedRequests
    }

    func stop() {
        lock.lock()
        let alreadyStopped = stopped
        stopped = true
        lock.unlock()
        if !alreadyStopped {
            close(listenSocket)
        }
    }

    private func serve(responder: @escaping (String) -> String) {
        while true {
            var clientAddress = sockaddr()
            var clientLength = socklen_t(MemoryLayout<sockaddr>.size)
            let client = accept(listenSocket, &clientAddress, &clientLength)
            guard client >= 0 else { return } // socket closed by stop()
            defer { close(client) }

            var head = Data()
            var buffer = [UInt8](repeating: 0, count: 2048)
            let terminator = Data("\r\n\r\n".utf8)
            while head.range(of: terminator) == nil {
                let count = read(client, &buffer, buffer.count)
                guard count > 0 else { break }
                head.append(contentsOf: buffer[0..<count])
                if head.count > 16 * 1024 { break }
            }
            let headText = String(decoding: head, as: UTF8.self)
            lock.lock()
            recordedRequests.append(headText)
            lock.unlock()
            let response = Data(responder(headText).utf8)
            response.withUnsafeBytes { raw in
                guard let base = raw.baseAddress else { return }
                _ = write(client, base, raw.count)
            }
        }
    }
}
