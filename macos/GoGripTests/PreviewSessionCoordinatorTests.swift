import Foundation
import XCTest

/// Coordinator lifecycle seam: a minimal controlled launch/exit seam makes
/// single-flight, generation isolation, reopen-while-stopping and stop-while-starting
/// deterministic. Real Go behavior through this coordinator is proven by the
/// temporary Foundation owner smoke, not by these fakes.
final class PreviewSessionCoordinatorTests: XCTestCase {
    private var tempRoot: URL!
    private let factory = ControlledFactory()

    private let urlA = URL(string: "http://127.0.0.1:49152/")!
    private let urlB = URL(string: "http://127.0.0.1:49153/")!

    override func setUpWithError() throws {
        tempRoot = FileManager.default.temporaryDirectory
            .appendingPathComponent("gogrip-coordinator-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: tempRoot, withIntermediateDirectories: true)
    }

    override func tearDownWithError() throws {
        if let tempRoot {
            try? FileManager.default.removeItem(at: tempRoot)
        }
    }

    // MARK: - Single flight and reuse

    @MainActor
    func testConcurrentRequestsForTheSameTargetShareOneLaunch() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let first = coordinator.open(directory)
        async let second = coordinator.open(directory)
        await waitUntil("the shared launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 101))

        let firstOutcome = await first
        let secondOutcome = await second
        let session = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(firstOutcome, .opened(session.id))
        XCTAssertEqual(secondOutcome, .opened(session.id))
        XCTAssertEqual(factory.count, 1, "concurrent requests must not start a second service")
        XCTAssertEqual(session.phase, .running)
        XCTAssertEqual(session.url, urlA)
        XCTAssertEqual(session.pid, 101)
    }

    @MainActor
    func testRunningTargetIsReusedWithoutStartingAnotherService() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let first = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 102))
        let firstOutcome = await first
        let id: PreviewSession.ID
        switch firstOutcome {
        case .opened(let opened): id = opened
        case .failed(let failure): XCTFail("unexpected open failure: \(failure)"); return
        }

        let secondOutcome = await coordinator.open(directory)
        let session = try XCTUnwrap(coordinator.session(id))
        XCTAssertEqual(secondOutcome, .opened(id))
        XCTAssertEqual(factory.count, 1, "a running target is reused, not restarted")
        XCTAssertEqual(session.url, urlA)
        XCTAssertEqual(session.pid, 102)
    }

    // MARK: - Generation isolation and stopping

    @MainActor
    func testReopenWhileStoppingWaitsForTheRealExit() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let first = coordinator.open(directory)
        await waitUntil("the first launch") { factory.count == 1 }
        let processA = try XCTUnwrap(factory.created.first)
        processA.completeStart(success(processA, url: urlA, pid: 103))
        let id = try await openedID(first)

        let stop = Task { await coordinator.stop(id) }
        await waitUntil("the stop to reach the adapter") { processA.stopCalls == 1 }

        async let reopen = coordinator.open(directory)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(factory.count, 1, "a reopen while stopping must wait for the old child")

        processA.completeStop(ManagedStopOutcome(generation: processA.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop.value
        await waitUntil("the new generation") { factory.count == 2 }
        let processB = try XCTUnwrap(factory.created.dropFirst().first)
        processB.completeStart(success(processB, url: urlB, pid: 104))

        let reopened = await reopen
        XCTAssertEqual(reopened, .opened(id))
        let session = try XCTUnwrap(coordinator.session(id))
        XCTAssertEqual(session.phase, .running)
        XCTAssertEqual(session.generation, processB.generation)
        XCTAssertEqual(session.url, urlB)
        XCTAssertEqual(session.pid, 104)
    }

    @MainActor
    func testStopDuringStartingWaitsForTheChildAndSuppressesTheLateResult() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the starting record") { factory.count == 1 && !coordinator.sessions.isEmpty }
        let process = try XCTUnwrap(factory.created.first)
        let id = try XCTUnwrap(coordinator.sessions.first?.id)
        XCTAssertEqual(coordinator.sessions.first?.phase, .starting)

        async let stopOne = coordinator.stop(id)
        async let stopTwo = coordinator.stop(id)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(process.stopCalls, 0, "no owned child exists yet; the stop must wait instead of reporting a no-child stop")
        XCTAssertFalse(process.isRunning)

        process.emit(.spawned(pid: 105))
        await waitUntil("the single adapter stop") { process.stopCalls == 1 }
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(process.stopCalls, 1, "repeated stops share one real finish")

        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .killedAfterGrace(exitCode: 9, reason: .uncaughtSignal)))
        _ = await stopOne
        _ = await stopTwo
        XCTAssertEqual(coordinator.sessions.first?.phase, .terminated)

        process.completeStart(success(process, url: urlA, pid: 105))
        try? await Task.sleep(nanoseconds: 50_000_000)
        let session = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(session.phase, .terminated, "a late startup result must not restore a stopped session")
        XCTAssertNil(session.url)
        XCTAssertEqual(factory.count, 1)

        let openOutcome = await open
        guard case .failed = openOutcome else { XCTFail("a stopped startup must not report success"); return }
    }

    @MainActor
    func testStaleGenerationEventsCannotOverwriteANewerSession() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let first = coordinator.open(directory)
        await waitUntil("the first launch") { factory.count == 1 }
        let processA = try XCTUnwrap(factory.created.first)
        processA.completeStart(success(processA, url: urlA, pid: 106))
        let id = try await openedID(first)

        let stop = Task { await coordinator.stop(id) }
        await waitUntil("the first stop") { processA.stopCalls == 1 }
        processA.completeStop(ManagedStopOutcome(generation: processA.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop.value

        async let reopen = coordinator.open(directory)
        await waitUntil("the second launch") { factory.count == 2 }
        let processB = try XCTUnwrap(factory.created.dropFirst().first)

        // Emitted on the main actor, these callbacks are applied inline, so the
        // session state below is observed after they were handled.
        processA.emit(.exited(exitCode: 9, reason: .exit))
        processA.emit(.targetStatus(ManagedTargetSnapshot(state: "unavailable", reason: "moved")))
        processA.emit(.fatal(code: "serve-failed", message: "late"))

        let starting = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(starting.generation, processB.generation)
        XCTAssertEqual(starting.phase, .starting)
        XCTAssertNil(starting.targetStatus)
        XCTAssertNil(starting.reloadStatus)
        XCTAssertNil(starting.lastFailure)

        processB.completeStart(success(processB, url: urlB, pid: 107))
        _ = await reopen
        let running = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(running.phase, .running)
        XCTAssertEqual(running.url, urlB)
        XCTAssertEqual(running.pid, 107)
    }

    // MARK: - Launch and exit failures

    @MainActor
    func testLaunchFailureEndsTheSessionWithTheAvailableReason() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(.failure(ManagedLaunchFailure(
            generation: process.generation,
            reason: .fatal(code: "target-unavailable", message: "the target is gone"),
            diagnostics: ""
        )))

        let openOutcome = await open
        guard case .failed(let failure) = openOutcome else { XCTFail("a failed launch must report failure"); return }
        XCTAssertTrue(failure.reason.contains("target-unavailable"))
        XCTAssertEqual(failure.displayPath, directory.standardizedFileURL.path)
        let session = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(session.phase, .terminated)
        XCTAssertNil(session.url)
        XCTAssertTrue(session.lastFailure?.contains("target-unavailable") == true)
    }

    @MainActor
    func testUnexpectedExitEndsRunningWithAReasonAndDoesNotRestart() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 108))
        let id = try await openedID(open)

        process.emit(.exited(exitCode: 7, reason: .exit))
        await waitUntil("the unexpected exit") { coordinator.session(id)?.phase == .terminated }
        let session = try XCTUnwrap(coordinator.session(id))
        XCTAssertNil(session.url)
        XCTAssertTrue(session.lastFailure?.contains("7") == true)

        try? await Task.sleep(nanoseconds: 100_000_000)
        XCTAssertEqual(factory.count, 1, "an unexpected exit must not restart the service")

        async let reopen = coordinator.open(directory)
        await waitUntil("the deliberate reopen") { factory.count == 2 }
        let reopened = try XCTUnwrap(factory.created.dropFirst().first)
        reopened.completeStart(success(reopened, url: urlB, pid: 109))
        let reopenOutcome = await reopen
        XCTAssertEqual(reopenOutcome, .opened(id))
        XCTAssertEqual(coordinator.session(id)?.phase, .running)
    }

    @MainActor
    func testExplicitStopIsNotReportedAsAnUnexpectedExit() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 110))
        let id = try await openedID(open)

        let stop = Task { await coordinator.stop(id) }
        await waitUntil("the stop") { process.stopCalls == 1 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop.value
        // The adapter reports the real exit for a requested stop as well.
        process.emit(.exited(exitCode: 0, reason: .exit))
        try? await Task.sleep(nanoseconds: 50_000_000)

        let session = try XCTUnwrap(coordinator.session(id))
        XCTAssertEqual(session.phase, .terminated)
        XCTAssertNil(session.url)
        XCTAssertNil(session.lastFailure, "an explicit stop is not an unexpected exit")
    }

    @MainActor
    func testStopFailureKeepsTheSessionAndAllowsARetryWithoutOverlap() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 111))
        let id = try await openedID(open)

        let firstStop = Task { await coordinator.stop(id) }
        await waitUntil("the first stop") { process.stopCalls == 1 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .failed("no owned child process")))
        await firstStop.value

        let failed = try XCTUnwrap(coordinator.session(id))
        XCTAssertEqual(failed.phase, .stopping, "an unconfirmed stop must not claim termination")
        XCTAssertTrue(failed.lastFailure?.contains("no owned child process") == true)

        // A reopen while the exit is unconfirmed may only retry the stop; it
        // must not start an overlapping generation. The outcome is observed
        // through a box so a regression (which would keep retrying) fails on a
        // deadline instead of hanging the test.
        let reopenBox = OutcomeBox()
        Task { reopenBox.value = await coordinator.open(directory) }
        await waitUntil("the stop retry from the reopen") { process.stopCalls == 2 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .failed("no owned child process")))
        await waitUntil("the reopen to report failure", timeout: 2) { reopenBox.value != nil }
        guard let reopenOutcome = reopenBox.value, case .failed(let reopenFailure) = reopenOutcome else {
            XCTFail("a reopen with an unconfirmed stop must not start a service")
            return
        }
        XCTAssertTrue(reopenFailure.reason.contains("no owned child process"))
        XCTAssertEqual(factory.count, 1, "no overlapping generation was started")

        let retry = Task { await coordinator.stop(id) }
        await waitUntil("the explicit stop retry") { process.stopCalls == 3 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .exited(exitCode: 0, reason: .exit)))
        await retry.value
        XCTAssertEqual(coordinator.session(id)?.phase, .terminated)
        XCTAssertNil(coordinator.session(id)?.url)
    }

    @MainActor
    func testExitBeforeStartupCommitStillAllowsADeliberateReopen() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        let id = try XCTUnwrap(coordinator.sessions.first?.id)

        // The child exits after the adapter committed readiness but before the
        // coordinator can commit it as running.
        process.emit(.exited(exitCode: 9, reason: .exit))
        await waitUntil("the early exit") { coordinator.session(id)?.phase == .terminated }
        process.completeStart(success(process, url: urlA, pid: 120))
        let firstOutcome = await open
        guard case .failed = firstOutcome else {
            XCTFail("a child that already exited must not commit a running session")
            return
        }
        XCTAssertNil(coordinator.session(id)?.url)

        async let reopen = coordinator.open(directory)
        await waitUntil("the deliberate reopen") { factory.count == 2 }
        let reopened = try XCTUnwrap(factory.created.dropFirst().first)
        reopened.completeStart(success(reopened, url: urlB, pid: 121))
        let reopenOutcome = await reopen
        XCTAssertEqual(reopenOutcome, .opened(id))
        XCTAssertEqual(coordinator.session(id)?.phase, .running)
        XCTAssertEqual(coordinator.session(id)?.url, urlB)
    }

    // MARK: - Observed target state and independent identities

    @MainActor
    func testTargetAndReloadStatesDoNotEndARunningSession() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let open = coordinator.open(directory)
        await waitUntil("the launch") { factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 112))
        let id = try await openedID(open)

        process.emit(.targetStatus(ManagedTargetSnapshot(state: "unavailable", reason: "moved")))
        process.emit(.reloadStatus(ManagedReloadSnapshot(state: "degraded", reason: "budget")))
        await waitUntil("the observed states") {
            coordinator.session(id)?.targetStatus != nil && coordinator.session(id)?.reloadStatus != nil
        }

        let session = try XCTUnwrap(coordinator.session(id))
        XCTAssertEqual(session.phase, .running)
        XCTAssertEqual(session.url, urlA)
        XCTAssertEqual(session.targetStatus?.state, "unavailable")
        XCTAssertEqual(session.reloadStatus?.state, "degraded")

        let stop = Task { await coordinator.stop(id) }
        await waitUntil("the stop") { process.stopCalls == 1 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop.value
        XCTAssertEqual(coordinator.session(id)?.phase, .terminated)
    }

    @MainActor
    func testParentChildAndFileTargetsStayIndependentSessions() async throws {
        let parent = try makeDirectory("parent")
        let child = try makeDirectory("parent/child")
        let file = try write("# note", to: "parent/child/note.md")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        let parentIdentity = try await identity(of: parent)
        let childIdentity = try await identity(of: child)
        let fileIdentity = try await identity(of: file)

        async let parentOpen = coordinator.open(parent)
        async let childOpen = coordinator.open(child)
        async let fileOpen = coordinator.open(file)
        await waitUntil("the three independent launches") { factory.count == 3 }
        var urls: [String: URL] = [:]
        for (index, process) in factory.created.enumerated() {
            let url = URL(string: "http://127.0.0.1:\(49152 + index)/")!
            urls[process.generation] = url
            process.completeStart(success(process, url: url, pid: Int32(200 + index)))
        }
        _ = await parentOpen
        _ = await childOpen
        _ = await fileOpen

        XCTAssertEqual(coordinator.sessions.count, 3)
        XCTAssertEqual(Set(coordinator.sessions.map(\.id)), [parentIdentity, childIdentity, fileIdentity])
        XCTAssertTrue(coordinator.sessions.allSatisfy { $0.phase == .running && $0.url != nil })
        XCTAssertEqual(Set(coordinator.sessions.compactMap(\.url)).count, 3)

        let launched = Dictionary(uniqueKeysWithValues: factory.created.map { ($0.target.path, $0) })
        XCTAssertEqual(launched[parentIdentity]?.mode, .directory)
        XCTAssertEqual(launched[childIdentity]?.mode, .directory)
        XCTAssertEqual(launched[fileIdentity]?.mode, .file)
        for session in coordinator.sessions {
            XCTAssertEqual(session.url, urls[session.generation])
        }

        let fileSession = try XCTUnwrap(coordinator.sessions.first { $0.id == fileIdentity })
        let stop = Task { await coordinator.stop(fileSession.id) }
        let fileProcess = try XCTUnwrap(factory.created.first { $0.generation == fileSession.generation })
        await waitUntil("the single stop") { fileProcess.stopCalls == 1 }
        fileProcess.completeStop(ManagedStopOutcome(generation: fileProcess.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop.value

        XCTAssertEqual(coordinator.session(fileIdentity)?.phase, .terminated)
        XCTAssertEqual(coordinator.session(parentIdentity)?.phase, .running)
        XCTAssertEqual(coordinator.session(childIdentity)?.phase, .running)
        let parentGeneration = try XCTUnwrap(coordinator.session(parentIdentity)?.generation)
        let childGeneration = try XCTUnwrap(coordinator.session(childIdentity)?.generation)
        XCTAssertEqual(coordinator.session(parentIdentity)?.url, urls[parentGeneration])
        XCTAssertEqual(coordinator.session(childIdentity)?.url, urls[childGeneration])
    }

    @MainActor
    func testStopAllStopsStartingAndRunningSessions() async throws {
        let directoryA = try makeDirectory("a")
        let directoryB = try makeDirectory("b")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        async let openA = coordinator.open(directoryA)
        await waitUntil("the starting session") { factory.count == 1 }
        let processA = try XCTUnwrap(factory.created.first)
        processA.emit(.spawned(pid: 113))
        async let openB = coordinator.open(directoryB)
        await waitUntil("the running session") { factory.count == 2 }
        let processB = try XCTUnwrap(factory.created.dropFirst().first)
        processB.completeStart(success(processB, url: urlB, pid: 114))
        _ = await openB
        XCTAssertEqual(coordinator.sessions.count, 2)

        let stopAll = Task { await coordinator.stopAll() }
        await waitUntil("both sessions to reach their adapter stop") {
            processA.stopCalls == 1 && processB.stopCalls == 1
        }
        processA.completeStop(ManagedStopOutcome(generation: processA.generation, result: .exited(exitCode: 0, reason: .exit)))
        processB.completeStop(ManagedStopOutcome(generation: processB.generation, result: .exited(exitCode: 0, reason: .exit)))
        processA.completeStart(.failure(ManagedLaunchFailure(
            generation: processA.generation,
            reason: .exitedEarly(status: 0),
            diagnostics: ""
        )))
        await stopAll.value

        XCTAssertTrue(coordinator.sessions.allSatisfy { $0.phase == .terminated })
        let openAOutcome = await openA
        guard case .failed = openAOutcome else { XCTFail("a session stopped while starting must not report success"); return }
    }

    @MainActor
    func testUnsupportedAndMissingTargetsNeverLaunchAService() async throws {
        let image = try write("png", to: "photo.png")
        let missing = tempRoot.appendingPathComponent("missing")
        let coordinator = PreviewSessionCoordinator(factory: factory)

        let unsupportedOutcome = await coordinator.open(image)
        guard case .failed(let unsupported) = unsupportedOutcome else {
            XCTFail("an unsupported file must not open"); return
        }
        XCTAssertEqual(unsupported.displayPath, image.standardizedFileURL.path)
        let missingOutcome = await coordinator.open(missing)
        guard case .failed = missingOutcome else {
            XCTFail("a missing target must not open"); return
        }
        XCTAssertEqual(factory.count, 0)
        XCTAssertTrue(coordinator.sessions.isEmpty)
    }

    // MARK: - Fixtures and helpers

    private func makeDirectory(_ relativePath: String) throws -> URL {
        let url = tempRoot.appendingPathComponent(relativePath)
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        return url
    }

    @discardableResult
    private func write(_ contents: String, to relativePath: String) throws -> URL {
        let url = tempRoot.appendingPathComponent(relativePath)
        try Data(contents.utf8).write(to: url)
        return url
    }

    private func identity(of url: URL) async throws -> PreviewSession.ID {
        switch await TargetPreparation.prepare(url) {
        case .success(let target): return target.identity
        case .failure(let failure): throw PreparationError.failed(failure)
        }
    }

    private func success(_ process: ControlledProcess, url: URL, pid: Int32 = 999) -> Result<ManagedLaunchSuccess, ManagedLaunchFailure> {
        .success(ManagedLaunchSuccess(
            generation: process.generation,
            url: url,
            reload: ManagedReloadSnapshot(state: "active", reason: nil),
            pid: pid
        ))
    }

    @MainActor
    private func openedID(_ outcome: OpenOutcome) throws -> PreviewSession.ID {
        switch outcome {
        case .opened(let id): return id
        case .failed(let failure): throw PreparationError.failed(.unavailable(failure.reason))
        }
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

    private enum PreparationError: Error {
        case failed(TargetPreparationFailure)
    }

    /// Main-actor holder that lets a test observe an open outcome without
    /// awaiting a task a regression could leave running.
    @MainActor
    private final class OutcomeBox {
        var value: OpenOutcome?
    }
}

/// Controlled launch seam: one scripted process per generation. Tests decide
/// when readiness, the child spawn, the real exit and stop outcomes happen.
private final class ControlledProcess: ManagedProcessLaunching, @unchecked Sendable {
    let generation: String
    let target: URL
    let mode: ManagedProcess.TargetMode

    private let events: (ManagedEvent) -> Void
    private let lock = NSLock()
    private var startResult: Result<ManagedLaunchSuccess, ManagedLaunchFailure>?
    private var startWaiters: [CheckedContinuation<Result<ManagedLaunchSuccess, ManagedLaunchFailure>, Never>] = []
    private var stopWaiters: [CheckedContinuation<ManagedStopOutcome, Never>] = []
    private var stopCallCount = 0
    private var running = false

    init(generation: String, target: URL, mode: ManagedProcess.TargetMode, events: @escaping (ManagedEvent) -> Void) {
        self.generation = generation
        self.target = target
        self.mode = mode
        self.events = events
    }

    var stopCalls: Int { withLock { stopCallCount } }
    var isRunning: Bool { withLock { running } }

    func start() async -> Result<ManagedLaunchSuccess, ManagedLaunchFailure> {
        await withCheckedContinuation { continuation in
            var immediate: Result<ManagedLaunchSuccess, ManagedLaunchFailure>?
            withLock {
                if let result = startResult {
                    immediate = result
                } else {
                    startWaiters.append(continuation)
                }
            }
            if let immediate {
                continuation.resume(returning: immediate)
            }
        }
    }

    func stop() async -> ManagedStopOutcome {
        // Each stop call waits for the next scripted completion, so a test can
        // model a failed stop followed by a successful retry.
        await withCheckedContinuation { continuation in
            withLock {
                stopCallCount += 1
                stopWaiters.append(continuation)
            }
        }
    }

    func completeStart(_ result: Result<ManagedLaunchSuccess, ManagedLaunchFailure>) {
        var waiters: [CheckedContinuation<Result<ManagedLaunchSuccess, ManagedLaunchFailure>, Never>] = []
        withLock {
            startResult = result
            if case .success = result {
                running = true
            }
            waiters = startWaiters
            startWaiters = []
        }
        for waiter in waiters {
            waiter.resume(returning: result)
        }
    }

    func completeStop(_ outcome: ManagedStopOutcome) {
        var waiters: [CheckedContinuation<ManagedStopOutcome, Never>] = []
        withLock {
            waiters = stopWaiters
            stopWaiters = []
        }
        for waiter in waiters {
            waiter.resume(returning: outcome)
        }
    }

    func emit(_ kind: ManagedEvent.Kind) {
        events(ManagedEvent(generation: generation, kind: kind))
    }

    private func withLock<T>(_ body: () -> T) -> T {
        lock.lock()
        defer { lock.unlock() }
        return body()
    }
}

/// Records every controlled process so tests can drive and count launches.
private final class ControlledFactory: ManagedProcessFactory, @unchecked Sendable {
    private let lock = NSLock()
    private var processes: [ControlledProcess] = []

    var created: [ControlledProcess] {
        lock.lock()
        defer { lock.unlock() }
        return processes
    }

    var count: Int { created.count }

    func makeProcess(
        generation: String,
        target: URL,
        mode: ManagedProcess.TargetMode,
        events: @escaping (ManagedEvent) -> Void
    ) -> ManagedProcessLaunching {
        let process = ControlledProcess(generation: generation, target: target, mode: mode, events: events)
        lock.lock()
        processes.append(process)
        lock.unlock()
        return process
    }
}
