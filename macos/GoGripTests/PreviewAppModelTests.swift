import Combine
import Foundation
import ServiceManagement
import XCTest

/// Host seams for the menu bar application: the production `PreviewAppModel`
/// with controlled default-browser results, controlled processes and a
/// controlled preparation for the quit admission window.
///
/// Real NSOpenPanel selection, panel actions, the browser page and a real
/// application quit are the development-app smoke, not this file.
final class PreviewAppModelTests: XCTestCase {
    private var tempRoot: URL!
    private let factory = ControlledFactory()
    private let browser = ControlledBrowser()
    private let loginItem = ControlledLoginItem()

    private let urlA = URL(string: "http://127.0.0.1:49152/")!
    private let urlB = URL(string: "http://127.0.0.1:49153/")!
    private var recentSuiteName: String!
    private var recentDefaults: UserDefaults!
    private var recentStore: RecentTargetsStore!

    override func setUpWithError() throws {
        tempRoot = FileManager.default.temporaryDirectory
            .appendingPathComponent("gogrip-model-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: tempRoot, withIntermediateDirectories: true)
        recentSuiteName = "gogrip-model-recent-\(UUID().uuidString)"
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

    // MARK: - Default browser result and recovery

    @MainActor
    func testRejectedBrowserRequestKeepsTheRunningSessionAndRecoversOnRetry() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        browser.result = false
        async let open = model.openBatch([directory])
        await waitUntil("the launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 201))
        await open

        let session = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(session.phase, .running)
        XCTAssertEqual(session.url, urlA)
        XCTAssertEqual(session.pid, 201)
        XCTAssertEqual(browser.requests, [urlA], "the browser must be asked for the verified URL")
        XCTAssertNotNil(session.browserFailure, "a rejected request is recorded without touching the phase")
        XCTAssertEqual(model.lastOperationFailures.map(\.displayPath), [session.displayPath])

        browser.result = true
        model.openBrowser(session.id)
        let recovered = try XCTUnwrap(coordinator.session(session.id))
        XCTAssertNil(recovered.browserFailure, "a successful request clears the recorded failure")
        XCTAssertEqual(recovered.phase, .running)
        XCTAssertEqual(recovered.url, urlA)
        XCTAssertEqual(browser.requests, [urlA, urlA])
        XCTAssertEqual(factory.count, 1, "recovery must not restart the service")
    }

    @MainActor
    func testFailedOpenNeverRequestsTheBrowserOrCreatesAFakeSession() async throws {
        let missing = tempRoot.appendingPathComponent("missing")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        await model.openBatch([missing])

        XCTAssertEqual(factory.count, 0)
        XCTAssertTrue(browser.requests.isEmpty)
        XCTAssertTrue(coordinator.sessions.isEmpty)
        XCTAssertEqual(model.lastOperationFailures.map(\.displayPath), [missing.standardizedFileURL.path])
    }

    @MainActor
    func testSuccessfulOpenDoesNotProduceAFailureReport() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        async let open = model.openBatch([directory])
        await waitUntil("the launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 202))
        await open

        let session = try XCTUnwrap(coordinator.sessions.first)
        XCTAssertEqual(session.phase, .running)
        XCTAssertNil(session.browserFailure)
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
        XCTAssertEqual(browser.requests, [urlA])
    }

    // MARK: - Quit admission

    @MainActor
    func testTerminationAdmissionRefusesLatePreparationAndLateStartup() async throws {
        let directoryA = try makeDirectory("a")
        let directoryB = try makeDirectory("b")
        let preparation = ControlledPreparation()
        let coordinator = PreviewSessionCoordinator(
            factory: factory,
            prepare: { await preparation.prepare($0) }
        )
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        // A starts before the quit request and is still starting.
        async let openA = model.openBatch([directoryA])
        await waitUntil("A's preparation") { preparation.pendingCount == 1 }
        preparation.complete(directoryA, result: await preparedResult(directoryA))
        await waitUntil("A's launch") { self.factory.count == 1 }
        let processA = try XCTUnwrap(factory.created.first)
        processA.emit(.spawned(pid: 301))

        // B's preparation is parked when the quit request closes admission.
        async let openB = model.openBatch([directoryB])
        await waitUntil("B's preparation") { preparation.pendingCount == 1 }
        let mayQuitImmediately = model.beginTermination()
        XCTAssertFalse(mayQuitImmediately, "a starting session still needs its real exit")
        preparation.complete(directoryB, result: await preparedResult(directoryB))
        await openB

        XCTAssertEqual(factory.count, 1, "no new child may start after the quit admission closed")
        XCTAssertEqual(coordinator.sessions.count, 1, "the late target must not become a session")
        XCTAssertEqual(coordinator.sessions.first?.displayPath, directoryA.standardizedFileURL.path)
        XCTAssertTrue(browser.requests.isEmpty)

        let termination = Task { await model.completeTermination() }
        await waitUntil("A's stop") { processA.stopCalls == 1 }
        processA.completeStart(success(processA, url: urlA, pid: 301))
        processA.completeStop(ManagedStopOutcome(generation: processA.generation, result: .exited(exitCode: 0, reason: .exit)))
        let mayExit = await termination.value

        XCTAssertTrue(mayExit, "every recorded service confirmed its real exit")
        _ = await openA
        XCTAssertTrue(coordinator.sessions.allSatisfy { $0.phase == .terminated })
        XCTAssertNil(coordinator.sessions.first?.url)
        XCTAssertEqual(processA.stopCalls, 1)
        XCTAssertTrue(browser.requests.isEmpty, "the browser must never open after the quit admission closed")
        XCTAssertTrue(model.lastOperationFailures.isEmpty, "a quit is not a failed user operation")
    }

    @MainActor
    func testTerminationAdmissionRefusesAQueuedLaunch() async throws {
        let directory = try makeDirectory("docs")
        let preparation = ControlledPreparation()
        let coordinator = PreviewSessionCoordinator(
            factory: factory,
            prepare: { await preparation.prepare($0) }
        )
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)
        // A launch that crosses the closed admission is the defect under test;
        // fail it so the regression reports the extra child instead of parking
        // the open behind it.
        factory.onCreate = { late in
            late.completeStart(.failure(ManagedLaunchFailure(
                generation: late.generation,
                reason: .spawnFailed("unexpected launch after the quit admission closed"),
                diagnostics: ""
            )))
        }

        async let open = model.openBatch([directory])
        await waitUntil("the preparation") { preparation.pendingCount == 1 }
        preparation.complete(directory, result: await preparedResult(directory))
        // Hand the main actor to the open: it passes both admission checks and
        // queues its launch, which stays pending behind this test.
        await Task.yield()
        XCTAssertEqual(coordinator.sessions.first?.phase, .starting, "the queued launch must still be starting")
        XCTAssertEqual(factory.count, 0, "the queued launch must not have reached the adapter yet")

        model.beginTermination()
        await open

        XCTAssertEqual(factory.count, 0, "a launch queued when the quit admission closed must not create a child")
        XCTAssertEqual(coordinator.sessions.first?.phase, .terminated, "the cancelled start must not stay starting")
        XCTAssertTrue(browser.requests.isEmpty)
    }

    @MainActor
    func testUnconfirmedStopRefusesTerminationAndKeepsTheSession() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        async let open = model.openBatch([directory])
        await waitUntil("the launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlB, pid: 302))
        await open
        let id = try XCTUnwrap(coordinator.sessions.first?.id)
        XCTAssertFalse(model.beginTermination())

        let termination = Task { await model.completeTermination() }
        await waitUntil("the stop") { process.stopCalls == 1 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .failed("no owned child process")))
        let mayExit = await termination.value

        XCTAssertFalse(mayExit, "an exit must not be allowed while a stop is unconfirmed")
        let session = try XCTUnwrap(coordinator.session(id))
        XCTAssertEqual(session.phase, .stopping, "the unconfirmed session keeps its management entry")
        XCTAssertEqual(session.url, urlB)
        XCTAssertFalse(model.lastOperationFailures.isEmpty)
        XCTAssertEqual(browser.requests, [urlB], "the quit wait must not add another browser request")
    }

    @MainActor
    func testTerminationAdmissionRefusesAReopenWaitingBehindAStop() async throws {
        let directory = try makeDirectory("docs")
        let preparation = ControlledPreparation()
        let admissionClosed = OneShotFlag()
        let coordinator = PreviewSessionCoordinator(
            factory: factory,
            prepare: { await preparation.prepare($0) }
        )
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)
        // The coordinator acknowledges the stop wait; closing the quit
        // admission there is deterministic: the reopen has passed both
        // admission checks and is parked for the old generation's stop.
        coordinator.willWaitForStop = { [weak model] in
            model?.beginTermination()
            admissionClosed.set()
        }

        async let open = model.openBatch([directory])
        await waitUntil("the preparation") { preparation.pendingCount == 1 }
        preparation.complete(directory, result: await preparedResult(directory))
        await waitUntil("the launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 401))
        await open
        let id = try XCTUnwrap(coordinator.sessions.first?.id)
        // A launch created after the quit admission closed is the defect under
        // test; fail it so the regression reports the extra child instead of
        // leaving a parked task behind.
        factory.onCreate = { late in
            late.completeStart(.failure(ManagedLaunchFailure(
                generation: late.generation,
                reason: .spawnFailed("unexpected launch after the quit admission closed"),
                diagnostics: ""
            )))
        }

        // A manual stop puts the session in .stopping; the reopen below parks
        // behind it after passing the admission checks.
        let stop = Task { await model.stop(id) }
        await waitUntil("the stop") { process.stopCalls == 1 }
        async let reopen = model.openBatch([directory])
        await waitUntil("the reopen preparation") { preparation.pendingCount == 1 }
        preparation.complete(directory, result: await preparedResult(directory))
        await waitUntil("the quit admission to close behind the parked reopen") { admissionClosed.isSet }

        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop.value
        await reopen

        XCTAssertEqual(factory.count, 1, "a reopen parked behind a stop must not start a child after the quit admission closed")
        XCTAssertEqual(coordinator.session(id)?.phase, .terminated)
        XCTAssertEqual(browser.requests, [urlA], "no browser request may follow the closed admission")
        let mayExit = await model.completeTermination()
        XCTAssertTrue(mayExit, "nothing owned remains to confirm")
    }

    @MainActor
    func testTerminationWithoutOwnedSessionsProceedsImmediately() async throws {
        let directory = try makeDirectory("docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        XCTAssertTrue(model.beginTermination(), "nothing owned means the app may quit right away")
        let mayExit = await model.completeTermination()
        XCTAssertTrue(mayExit)

        await model.openBatch([directory])
        XCTAssertEqual(factory.count, 0, "an open after the quit request must not start a service")
        XCTAssertTrue(coordinator.sessions.isEmpty)
        XCTAssertTrue(browser.requests.isEmpty)
    }

    // MARK: - Unified batch: preparation, quantity and cancellation

    @MainActor
    func testFiveDistinctTargetsOpenDirectlyWithoutQuantityConfirmation() async throws {
        let directories = try (0..<5).map { try makeDirectory("direct\($0)") }
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)

        await model.openBatch(directories)

        XCTAssertTrue(confirmation.requests.isEmpty, "five distinct valid targets must open without a quantity confirmation")
        XCTAssertEqual(factory.count, 5)
        XCTAssertEqual(coordinator.sessions.count, 5)
        XCTAssertTrue(coordinator.sessions.allSatisfy { $0.phase == .running })
        XCTAssertEqual(browser.requests, (0..<5).map { launches.url($0) }, "every opened target must get its own verified URL")
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
    }

    @MainActor
    func testSixDistinctTargetsAskForTheActualCountAndOpenAfterApproval() async throws {
        let directories = try (0..<6).map { try makeDirectory("confirm\($0)") }
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)

        await model.openBatch(directories)

        XCTAssertEqual(confirmation.requests, [6], "the confirmation must carry the number of distinct valid targets")
        XCTAssertEqual(factory.count, 6, "the approved batch must be processed")
        XCTAssertEqual(coordinator.sessions.count, 6)
        XCTAssertEqual(browser.requests.count, 6)
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
    }

    @MainActor
    func testCancelledQuantityConfirmationStartsNothingAndKeepsExistingSessions() async throws {
        let directories = try (0..<6).map { try makeDirectory("cancel\($0)") }
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)

        // One of the six selected targets is already running when the batch is cancelled.
        await model.openBatch([directories[0]])
        let existing = try XCTUnwrap(coordinator.sessions.first)

        confirmation.result = false
        await model.openBatch(directories)

        XCTAssertEqual(confirmation.requests, [6], "the running target still counts as one of the six distinct valid targets")
        XCTAssertEqual(factory.count, 1, "a cancelled batch must not start a service")
        XCTAssertEqual(browser.requests, [launches.url(0)], "a cancelled batch must not reopen the running target in the browser")
        XCTAssertEqual(coordinator.sessions.count, 1)
        let kept = try XCTUnwrap(coordinator.session(existing.id))
        XCTAssertEqual(kept.phase, .running, "cancelling must not stop an existing session")
        XCTAssertEqual(kept.generation, existing.generation)
        XCTAssertEqual(kept.url, existing.url)
        XCTAssertEqual(kept.pid, existing.pid)
        XCTAssertTrue(model.lastOperationFailures.isEmpty, "cancelling is not a failed operation")
    }

    @MainActor
    func testSixPathsThatResolveToFiveIdentitiesSkipTheQuantityConfirmation() async throws {
        let directories = try (0..<5).map { try makeDirectory("alias\($0)") }
        let alias = tempRoot.appendingPathComponent("alias-link")
        try FileManager.default.createSymbolicLink(at: alias, withDestinationURL: directories[0])
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)

        await model.openBatch(directories + [alias])

        XCTAssertTrue(confirmation.requests.isEmpty, "six paths with five identities must not ask for confirmation")
        XCTAssertEqual(factory.count, 5)
        XCTAssertEqual(coordinator.sessions.count, 5)
        XCTAssertEqual(
            coordinator.sessions.filter { $0.id == resolvedPath(directories[0]) }.count, 1,
            "an alias must not create a second service"
        )
        XCTAssertEqual(browser.requests.count, 5)
    }

    @MainActor
    func testUnsupportedInputDoesNotCountTowardsTheQuantityConfirmation() async throws {
        let directories = try (0..<5).map { try makeDirectory("counted\($0)") }
        let image = tempRoot.appendingPathComponent("photo.png")
        try Data("not an image".utf8).write(to: image)
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)

        await model.openBatch(directories + [image])

        XCTAssertTrue(confirmation.requests.isEmpty, "an unsupported file must not count as a valid target")
        XCTAssertEqual(factory.count, 5)
        XCTAssertEqual(coordinator.sessions.count, 5)
        XCTAssertEqual(browser.requests.count, 5)
        XCTAssertEqual(model.lastOperationFailures.map(\.displayPath), [image.standardizedFileURL.path])
    }

    // MARK: - Mixed batches: continue, one consolidated report, recovery

    @MainActor
    func testMixedBatchContinuesPastFailuresAndReportsThemOnce() async throws {
        let directory = try makeDirectory("nested/docs")
        let article = tempRoot.appendingPathComponent("说明 文章.md")
        try Data("# 你好".utf8).write(to: article)
        let empty = try makeDirectory("empty")
        let image = tempRoot.appendingPathComponent("photo.png")
        try Data("not an image".utf8).write(to: image)
        let missing = tempRoot.appendingPathComponent("missing")
        let failingLaunch = try makeDirectory("launch-failure")
        let rejectedBrowser = try makeDirectory("browser-rejected")
        let selected = [directory, article, empty, image, missing, failingLaunch, rejectedBrowser]

        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.failing = [resolvedPath(failingLaunch)]
        launches.install(on: factory)
        browser.rejectURLs = [launches.url(4)]
        let reports = ReportRecorder(model: model)

        await model.openBatch(selected)

        // The two invalid inputs never reached a launch; every other target did,
        // including the ones after the failed launch.
        XCTAssertEqual(factory.count, 5)
        XCTAssertTrue(confirmation.requests.isEmpty, "the distinct valid count stays at five without the invalid inputs")
        let running = coordinator.sessions.filter { $0.phase == .running }
        XCTAssertEqual(running.count, 4, "a failed launch must not stop the remaining targets")
        let articleSession = try XCTUnwrap(coordinator.session(resolvedPath(article)))
        XCTAssertEqual(articleSession.displayPath, article.standardizedFileURL.path)
        XCTAssertEqual(articleSession.url, launches.url(1))
        XCTAssertEqual(coordinator.session(resolvedPath(empty))?.url, launches.url(2))
        XCTAssertEqual(
            coordinator.session(resolvedPath(failingLaunch))?.phase, .terminated,
            "a failed launch must not be reported as running"
        )
        XCTAssertNil(coordinator.session(resolvedPath(failingLaunch))?.url)
        XCTAssertEqual(
            browser.requests, [launches.url(0), launches.url(1), launches.url(2), launches.url(4)],
            "the browser must be asked for each verified URL, including the rejected one"
        )

        // One consolidated report: unsupported, preparation, launch and browser
        // failures with their targets and available reasons.
        XCTAssertEqual(
            model.lastOperationFailures.map(\.displayPath),
            [image, missing, failingLaunch, rejectedBrowser].map { $0.standardizedFileURL.path }
        )
        XCTAssertTrue(model.lastOperationFailures.allSatisfy { !$0.reason.isEmpty })
        XCTAssertEqual(
            Set(model.lastOperationFailures.map(\.reason)).count, model.lastOperationFailures.count,
            "each failure must carry the reason available for its own cause instead of one boilerplate line"
        )
        let launchFailure = try XCTUnwrap(model.lastOperationFailures.first { $0.displayPath == failingLaunch.standardizedFileURL.path })
        XCTAssertTrue(
            launchFailure.reason.contains("refused to start"),
            "the report must carry the reason the launch was available: \(launchFailure.reason)"
        )
        XCTAssertEqual(reports.count, 1, "one failed batch must publish one consolidated report")

        // A rejected browser request keeps the running session and its URL, and
        // a later accepted request recovers it without a new service.
        let rejected = try XCTUnwrap(coordinator.session(resolvedPath(rejectedBrowser)))
        XCTAssertEqual(rejected.url, launches.url(4))
        XCTAssertNotNil(rejected.browserFailure)
        browser.rejectURLs = []
        model.openBrowser(rejected.id)
        XCTAssertNil(coordinator.session(rejected.id)?.browserFailure)
        XCTAssertEqual(coordinator.session(rejected.id)?.phase, .running)
        XCTAssertEqual(coordinator.session(rejected.id)?.url, launches.url(4))
        XCTAssertEqual(factory.count, 5, "recovering the browser request must not restart the service")
    }

    /// One consolidated report carries the applicable check path for failures
    /// that stopped the target from being read — both the preparation-stage
    /// failure and the service-reported access failure — while a cause that is
    /// not about access keeps its own reason and is not pointed at
    /// authorization settings.
    @MainActor
    func testAccessFailuresCarryConditionalGuidanceWithoutCoveringOtherCauses() async throws {
        let missing = tempRoot.appendingPathComponent("missing")
        let denied = try makeDirectory("denied")
        let image = tempRoot.appendingPathComponent("photo.png")
        try Data("not an image".utf8).write(to: image)

        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)
        let deniedIdentity = resolvedPath(denied)
        factory.onCreate = { process in
            guard process.target.path == deniedIdentity else { return }
            process.completeStart(.failure(ManagedLaunchFailure(
                generation: process.generation,
                reason: .fatal(
                    code: "target-unavailable",
                    message: "access target: open \(deniedIdentity): operation not permitted"
                ),
                diagnostics: ""
            )))
        }

        await model.openBatch([missing, denied, image])

        XCTAssertEqual(factory.count, 1, "only the denied target may reach a launch")
        XCTAssertTrue(browser.requests.isEmpty)
        XCTAssertEqual(model.lastOperationFailures.count, 3)
        let reasons = Dictionary(uniqueKeysWithValues: model.lastOperationFailures.map {
            ($0.displayPath, $0.reason)
        })
        let missingReason = try XCTUnwrap(reasons[missing.standardizedFileURL.path])
        XCTAssertTrue(
            missingReason.contains(missing.lastPathComponent),
            "the reason available for the preparation failure must stay: \(missingReason)"
        )
        XCTAssertTrue(
            missingReason.contains("Files and Folders"),
            "an access failure must carry the applicable check path: \(missingReason)"
        )
        let deniedReason = try XCTUnwrap(reasons[denied.standardizedFileURL.path])
        XCTAssertTrue(
            deniedReason.contains("operation not permitted"),
            "the reason available for the service-reported failure must stay: \(deniedReason)"
        )
        XCTAssertTrue(
            deniedReason.contains("Files and Folders"),
            "an access failure must carry the applicable check path: \(deniedReason)"
        )
        let imageReason = try XCTUnwrap(reasons[image.standardizedFileURL.path])
        XCTAssertFalse(
            imageReason.contains("Files and Folders"),
            "an unsupported file must not be pointed at authorization checks: \(imageReason)"
        )
        let deniedSession = try XCTUnwrap(coordinator.session(deniedIdentity))
        XCTAssertEqual(deniedSession.phase, .terminated, "an access failure must not fake a running session")
        XCTAssertNil(deniedSession.url)
    }

    @MainActor
    func testSuccessfulBatchPublishesNoReport() async throws {
        let directories = try (0..<3).map { try makeDirectory("clean\($0)") }
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)
        let reports = ReportRecorder(model: model)

        await model.openBatch(directories)

        XCTAssertTrue(coordinator.sessions.allSatisfy { $0.phase == .running })
        XCTAssertEqual(browser.requests, (0..<3).map { launches.url($0) })
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
        XCTAssertEqual(reports.count, 0, "a fully successful batch must not publish a report")
    }

    // MARK: - Batch waits behind the quit admission

    @MainActor
    func testQuitAdmissionDuringTheQuantityConfirmationStartsNothing() async throws {
        let directories = try (0..<6).map { try makeDirectory("waiting\($0)") }
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)
        // The quit admission closes while the user is answering the confirmation.
        confirmation.onConfirm = { model.beginTermination() }

        await model.openBatch(directories)

        XCTAssertEqual(confirmation.requests, [6])
        XCTAssertEqual(factory.count, 0, "a batch approved after the quit admission closed must not start a service")
        XCTAssertTrue(coordinator.sessions.isEmpty)
        XCTAssertTrue(browser.requests.isEmpty, "no browser request may follow the closed admission")
        XCTAssertTrue(model.lastOperationFailures.isEmpty, "unprocessed targets are not failures")
        let mayExit = await model.completeTermination()
        XCTAssertTrue(mayExit, "nothing owned remains to confirm")
    }

    @MainActor
    func testQuitAdmissionDuringABatchLaunchStopsTheRemainingTargets() async throws {
        let directories = try (0..<3).map { try makeDirectory("parked\($0)") }
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let confirmation = ControlledConfirmation()
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, confirmation: confirmation, recentStore: recentStore, loginItem: loginItem)

        async let batch = model.openBatch(directories)
        await waitUntil("the first launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        XCTAssertFalse(model.beginTermination(), "the parked launch still needs its real exit")

        process.completeStart(success(process, url: urlA, pid: 601))
        await batch

        XCTAssertEqual(factory.count, 1, "the targets after the parked launch must not start a service")
        XCTAssertEqual(coordinator.sessions.count, 1)
        let id = try XCTUnwrap(coordinator.sessions.first?.id)
        XCTAssertEqual(coordinator.session(id)?.phase, .running)
        XCTAssertEqual(coordinator.session(id)?.url, urlA, "the created child keeps its verified session facts")
        XCTAssertTrue(browser.requests.isEmpty, "a verified URL must not open after the quit admission closed")
        XCTAssertTrue(model.lastOperationFailures.isEmpty)

        let termination = Task { await model.completeTermination() }
        await waitUntil("the stop") { process.stopCalls == 1 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .exited(exitCode: 0, reason: .exit)))
        let mayExit = await termination.value

        XCTAssertTrue(mayExit, "the owned child confirmed its real exit")
        XCTAssertEqual(coordinator.session(id)?.phase, .terminated)
        XCTAssertNil(coordinator.session(id)?.url)
        XCTAssertTrue(browser.requests.isEmpty)
    }

    // MARK: - Recent targets

    /// A target becomes recent only after a verified running session exists; a
    /// preparation failure must not enter the list, and a rejected browser
    /// request must not revoke the target that did become running.
    @MainActor
    func testOnlyVerifiedSessionsEnterRecentEvenWhenTheBrowserRejects() async throws {
        let directory = try makeDirectory("recent-docs")
        let missing = tempRoot.appendingPathComponent("missing")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        browser.result = false
        async let open = model.openBatch([directory, missing])
        await waitUntil("the launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 601))
        await open

        let session = try XCTUnwrap(coordinator.session(resolvedPath(directory)))
        XCTAssertEqual(session.phase, .running)
        XCTAssertEqual(session.url, urlA)
        XCTAssertEqual(browser.requests, [urlA], "the verified target requests the browser once")
        XCTAssertNotNil(session.browserFailure, "a rejected browser request is a browser failure, not a target failure")

        XCTAssertEqual(
            model.recentTargets.map(\.identity),
            [resolvedPath(directory)],
            "only the verified target may enter recents; the missing target must not"
        )
        let entry = try XCTUnwrap(model.recentTargets.first)
        XCTAssertEqual(entry.displayPath, directory.standardizedFileURL.path)
        XCTAssertEqual(entry.mode, .directory)
        XCTAssertEqual(
            RecentTargetsStore(defaults: recentDefaults).load().map(\.identity),
            [resolvedPath(directory)],
            "the verified target must be persisted even though the browser request was rejected"
        )
        XCTAssertTrue(
            model.lastOperationFailures.contains { $0.displayPath == missing.standardizedFileURL.path },
            "the failed target stays in the one batch report"
        )
    }

    /// Reopening a running target from its recent row or through its symbolic
    /// link alias shares the production batch entry: the same session (PID,
    /// generation, verified URL) is reused, one identity keeps one recent row,
    /// and only the recency order changes.
    @MainActor
    func testReopeningARunningTargetFromRecentOrItsAliasReusesTheSessionAndRefreshesRecency() async throws {
        let alpha = try makeDirectory("alpha")
        let beta = try makeDirectory("beta")
        let alias = tempRoot.appendingPathComponent("alias-alpha")
        try FileManager.default.createSymbolicLink(at: alias, withDestinationURL: alpha)

        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)
        let launches = SequentialLaunchScript()
        launches.install(on: factory)

        await model.openBatch([alpha])
        await model.openBatch([beta])
        XCTAssertEqual(model.recentTargets.map(\.identity), [resolvedPath(beta), resolvedPath(alpha)])

        let alphaSession = try XCTUnwrap(coordinator.session(resolvedPath(alpha)))
        let alphaURL = try XCTUnwrap(alphaSession.url)

        let entry = try XCTUnwrap(model.recentTargets.first { $0.identity == resolvedPath(alpha) })
        await model.reopenRecent(entry)

        XCTAssertEqual(factory.count, 2, "reopening a running target must not start a second service")
        let reused = try XCTUnwrap(coordinator.session(resolvedPath(alpha)))
        XCTAssertEqual(reused.generation, alphaSession.generation)
        XCTAssertEqual(reused.pid, alphaSession.pid)
        XCTAssertEqual(reused.url, alphaURL)
        XCTAssertEqual(model.recentTargets.map(\.identity), [resolvedPath(alpha), resolvedPath(beta)])

        await model.openBatch([alias])

        XCTAssertEqual(factory.count, 2, "opening the alias of a running target reuses the same session")
        XCTAssertEqual(coordinator.session(resolvedPath(alpha))?.generation, alphaSession.generation)
        XCTAssertEqual(coordinator.session(resolvedPath(alpha))?.url, alphaURL)
        XCTAssertEqual(
            model.recentTargets.filter { $0.identity == resolvedPath(alpha) }.count,
            1,
            "an alias must not create a duplicate recent row"
        )
        XCTAssertEqual(model.recentTargets.map(\.identity), [resolvedPath(alpha), resolvedPath(beta)])
        XCTAssertEqual(
            model.recentTargets.first?.displayPath,
            alias.standardizedFileURL.path,
            "the recency row must keep the latest user-selected path (the alias)"
        )
        XCTAssertEqual(
            browser.requests.filter { $0 == alphaURL }.count,
            3,
            "each explicit reopen requests the browser for the same verified URL"
        )

        // A link retargeted after recording must not change what the recent
        // row opens: the recorded normalized target is the location, and the
        // recorded user-selected path stays the display path.
        let gamma = try makeDirectory("gamma")
        try FileManager.default.removeItem(at: alias)
        try FileManager.default.createSymbolicLink(at: alias, withDestinationURL: gamma)

        await model.reopenRecent(try XCTUnwrap(model.recentTargets.first { $0.identity == resolvedPath(alpha) }))

        XCTAssertEqual(factory.count, 2, "a retargeted link must not start a service for the new destination")
        XCTAssertNil(coordinator.session(resolvedPath(gamma)), "the recent row must not open the retargeted destination")
        XCTAssertEqual(coordinator.session(resolvedPath(alpha))?.generation, alphaSession.generation)
        XCTAssertEqual(coordinator.session(resolvedPath(alpha))?.url, alphaURL)
        XCTAssertEqual(
            model.recentTargets.filter { $0.identity == resolvedPath(alpha) }.count,
            1,
            "reopening the recorded target must not add a second row"
        )
        XCTAssertFalse(model.recentTargets.contains { $0.identity == resolvedPath(gamma) })
        XCTAssertEqual(
            model.recentTargets.first?.displayPath,
            alias.standardizedFileURL.path,
            "the recorded selected path stays the display path after a retargeted reopen"
        )

        // Task 4.1 requires recency to follow explicit reopens, and the
        // running row's Open is such a reopen: the older target moves back to
        // the front of the recent list without duplicating it, without
        // changing its recorded display path and without starting a service.
        let betaDisplayPath = try XCTUnwrap(model.recentTargets.first { $0.identity == resolvedPath(beta) }?.displayPath)
        let betaSession = try XCTUnwrap(coordinator.session(resolvedPath(beta)))
        let betaURL = try XCTUnwrap(betaSession.url)
        model.openBrowser(betaSession.id)

        XCTAssertEqual(factory.count, 2, "reopening from the running row must not start a service")
        XCTAssertEqual(model.recentTargets.map(\.identity), [resolvedPath(beta), resolvedPath(alpha)])
        XCTAssertEqual(model.recentTargets.filter { $0.identity == resolvedPath(beta) }.count, 1)
        XCTAssertEqual(
            model.recentTargets.first?.displayPath,
            betaDisplayPath,
            "the moved row keeps its recorded display path"
        )
        XCTAssertEqual(
            browser.requests.filter { $0 == betaURL }.count,
            2,
            "the running row's reopen requests the browser for the same verified URL"
        )
    }

    /// Clearing recents only clears the independent key and the published
    /// list: the running session keeps its URL and PID, its stop control still
    /// stops the real service, and a later launch stays empty.
    @MainActor
    func testClearingRecentTargetsKeepsTheRunningSessionAndItsStopControl() async throws {
        let directory = try makeDirectory("clear-docs")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(coordinator: coordinator, browser: browser, recentStore: recentStore, loginItem: loginItem)

        async let open = model.openBatch([directory])
        await waitUntil("the launch") { self.factory.count == 1 }
        let process = try XCTUnwrap(factory.created.first)
        process.completeStart(success(process, url: urlA, pid: 602))
        await open

        let session = try XCTUnwrap(coordinator.session(resolvedPath(directory)))
        XCTAssertEqual(model.recentTargets.count, 1)

        model.clearRecentTargets()

        XCTAssertTrue(model.recentTargets.isEmpty)
        XCTAssertTrue(RecentTargetsStore(defaults: recentDefaults).load().isEmpty, "a later launch must stay empty")
        let running = try XCTUnwrap(coordinator.session(resolvedPath(directory)))
        XCTAssertEqual(running.phase, .running, "clearing recents must not touch the running session")
        XCTAssertEqual(running.url, urlA)
        XCTAssertEqual(running.pid, 602)
        XCTAssertEqual(browser.requests, [urlA], "clearing recents must not open or restart anything")

        async let stop = model.stop(running.id)
        await waitUntil("the stop") { process.stopCalls == 1 }
        process.completeStop(ManagedStopOutcome(generation: process.generation, result: .exited(exitCode: 0, reason: .exit)))
        await stop

        XCTAssertEqual(coordinator.session(running.id)?.phase, .terminated)
        XCTAssertTrue(model.recentTargets.isEmpty)
    }

    // MARK: - Opt-in login startup (task 5.3)

    /// A rejected enable keeps the real system status and reports the reason
    /// the system gave; the panel must never present the requested value as if
    /// it had taken effect.
    @MainActor
    func testRejectedLoginStartEnableKeepsTheRealStatusAndReportsTheReason() throws {
        loginItem.status = .notRegistered
        loginItem.registerError = loginReason("Operation not permitted")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(
            coordinator: coordinator,
            browser: browser,
            recentStore: recentStore,
            loginItem: loginItem
        )
        let recorder = ReportRecorder(model: model)

        model.setLoginItemEnabled(true)

        XCTAssertEqual(model.loginItemStatus, .notRegistered, "the real status must win over the requested one")
        XCTAssertEqual(recorder.count, 1, "a rejected operation is reported once through the existing root report")
        XCTAssertEqual(model.lastOperationFailures.count, 1)
        XCTAssertTrue(
            model.lastOperationFailures[0].reason.contains("Operation not permitted"),
            "the reason the system gave stays available"
        )
    }

    /// A register that leaves the system asking for approval is shown as that
    /// real state — never as enabled — and is not a failure report.
    @MainActor
    func testRegistrationRequiringApprovalIsShownAsItsRealStateNotAsEnabled() throws {
        loginItem.status = .notRegistered
        loginItem.registeredStatus = .requiresApproval
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(
            coordinator: coordinator,
            browser: browser,
            recentStore: recentStore,
            loginItem: loginItem
        )
        let recorder = ReportRecorder(model: model)

        model.setLoginItemEnabled(true)

        XCTAssertEqual(model.loginItemStatus, .requiresApproval, "approval is the real state after registering")
        XCTAssertNotEqual(model.loginItemStatus, .enabled, "a registered item waiting for approval is not enabled")
        XCTAssertEqual(recorder.count, 0, "waiting for approval is a real state, not a failed operation")
        XCTAssertTrue(model.lastOperationFailures.isEmpty)
    }

    /// A rejected disable keeps the enabled status and reports the reason; the
    /// displayed value is never flipped to pretend the change happened.
    @MainActor
    func testRejectedLoginStartDisableKeepsTheEnabledStatusAndReportsTheReason() throws {
        loginItem.status = .enabled
        loginItem.unregisterError = loginReason("Operation not permitted")
        let coordinator = PreviewSessionCoordinator(factory: factory)
        let model = PreviewAppModel(
            coordinator: coordinator,
            browser: browser,
            recentStore: recentStore,
            loginItem: loginItem
        )
        let recorder = ReportRecorder(model: model)

        model.setLoginItemEnabled(false)

        XCTAssertEqual(model.loginItemStatus, .enabled, "the real status must win over the requested one")
        XCTAssertEqual(recorder.count, 1)
        XCTAssertTrue(model.lastOperationFailures[0].reason.contains("Operation not permitted"))
    }

    /// A controlled system error carrying the exact reason the real system
    /// would provide.
    private func loginReason(_ description: String) -> NSError {
        NSError(
            domain: "SMAppServiceErrorDomain",
            code: 1,
            userInfo: [NSLocalizedDescriptionKey: description]
        )
    }

    // MARK: - Fixtures and helpers

    private func makeDirectory(_ relativePath: String) throws -> URL {
        let url = tempRoot.appendingPathComponent(relativePath)
        try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        return url
    }

    private func preparedResult(_ url: URL) async -> Result<PreviewTarget, TargetPreparationFailure> {
        await TargetPreparation.prepare(url)
    }

    private func resolvedPath(_ url: URL) -> String {
        URL(fileURLWithPath: url.path).resolvingSymlinksInPath().path
    }

    private func success(_ process: ControlledProcess, url: URL, pid: Int32) -> Result<ManagedLaunchSuccess, ManagedLaunchFailure> {
        .success(ManagedLaunchSuccess(
            generation: process.generation,
            url: url,
            reload: ManagedReloadSnapshot(state: "active", reason: nil),
            pid: pid
        ))
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

/// Controlled login-item seam: scripts the real system's transitions and
/// failures, so the panel's consumer facts are observed without touching the
/// real login items of the machine. Internal to the test target because both
/// host test files inject it.
final class ControlledLoginItem: LoginItemRegistering {
    /// The status the system reports before any operation.
    var status: SMAppService.Status
    /// The status the system reports after a successful register; the real
    /// system may require approval instead of enabling the item directly.
    var registeredStatus: SMAppService.Status = .enabled
    var registerError: Error?
    var unregisterError: Error?

    init(status: SMAppService.Status = .notRegistered) {
        self.status = status
    }

    func register() throws {
        if let registerError { throw registerError }
        status = registeredStatus
    }

    func unregister() throws {
        if let unregisterError { throw unregisterError }
        status = .notRegistered
    }

    func openLoginItemsSettings() {}
}

/// Records every non-empty report the model publishes, so "one alert per
/// batch" is observed instead of inferred.
@MainActor
private final class ReportRecorder {
    private(set) var reports: [[OperationFailure]] = []
    private var cancellable: AnyCancellable?

    var count: Int { reports.count }

    init(model: PreviewAppModel) {
        cancellable = model.$lastOperationFailures.sink { [weak self] failures in
            guard !failures.isEmpty else { return }
            self?.reports.append(failures)
        }
    }
}

/// Controlled confirmation seam: records every requested count and returns the
/// scripted answer, so the batch boundary is deterministic.
@MainActor
private final class ControlledConfirmation: BatchQuantityConfirming {
    private(set) var requests: [Int] = []
    var result = true
    /// Runs while the batch waits for the answer, for the quit-admission window.
    var onConfirm: (() -> Void)?

    func confirmOpening(count: Int) async -> Bool {
        requests.append(count)
        onConfirm?()
        return result
    }
}

/// Completes every created launch immediately with its own verified URL in
/// creation order, so a sequential batch runs to completion without parking.
private final class SequentialLaunchScript: @unchecked Sendable {
    private let lock = NSLock()
    private var next = 0

    /// Resolved target paths whose launch must fail instead of opening.
    var failing: Set<String> = []

    func url(_ index: Int) -> URL {
        URL(string: "http://127.0.0.1:\(49_200 + index)/")!
    }

    func install(on factory: ControlledFactory) {
        factory.onCreate = { [self] process in
            lock.lock()
            let index = next
            next += 1
            lock.unlock()
            if failing.contains(process.target.path) {
                process.completeStart(.failure(ManagedLaunchFailure(
                    generation: process.generation,
                    reason: .spawnFailed("the preview service refused to start"),
                    diagnostics: ""
                )))
                return
            }
            process.completeStart(.success(ManagedLaunchSuccess(
                generation: process.generation,
                url: url(index),
                reload: ManagedReloadSnapshot(state: "active", reason: nil),
                pid: Int32(500 + index)
            )))
        }
    }
}

/// Controlled browser seam: records every requested URL and returns the
/// scripted system result.
private final class ControlledBrowser: BrowserOpening, @unchecked Sendable {
    private let lock = NSLock()
    private var recorded: [URL] = []
    private var scriptedResult = true
    private var rejected: Set<URL> = []

    var requests: [URL] {
        lock.lock()
        defer { lock.unlock() }
        return recorded
    }

    var result: Bool {
        get {
            lock.lock()
            defer { lock.unlock() }
            return scriptedResult
        }
        set {
            lock.lock()
            scriptedResult = newValue
            lock.unlock()
        }
    }

    /// URLs the system rejects regardless of `result`.
    var rejectURLs: Set<URL> {
        get {
            lock.lock()
            defer { lock.unlock() }
            return rejected
        }
        set {
            lock.lock()
            rejected = newValue
            lock.unlock()
        }
    }

    @discardableResult
    func open(_ url: URL) -> Bool {
        lock.lock()
        recorded.append(url)
        let result = rejected.contains(url) ? false : scriptedResult
        lock.unlock()
        return result
    }
}

/// One-shot, thread-safe flag shared between the test, a preparation closure
/// and a main-actor job.
private final class OneShotFlag: @unchecked Sendable {
    private let lock = NSLock()
    private var value = false

    var isSet: Bool {
        lock.lock()
        defer { lock.unlock() }
        return value
    }

    func set() {
        lock.lock()
        value = true
        lock.unlock()
    }

    func take() -> Bool {
        lock.lock()
        defer { lock.unlock() }
        let taken = value
        value = false
        return taken
    }
}

/// Controlled preparation seam: parks each request until the test completes
/// it, so the quit admission window can be exercised deterministically.
private final class ControlledPreparation: @unchecked Sendable {
    private let lock = NSLock()
    private var waiters: [(URL, CheckedContinuation<Result<PreviewTarget, TargetPreparationFailure>, Never>)] = []

    var pendingCount: Int {
        lock.lock()
        defer { lock.unlock() }
        return waiters.count
    }

    func prepare(_ url: URL) async -> Result<PreviewTarget, TargetPreparationFailure> {
        await withCheckedContinuation { continuation in
            lock.lock()
            waiters.append((url, continuation))
            lock.unlock()
        }
    }

    func complete(_ url: URL, result: Result<PreviewTarget, TargetPreparationFailure>) {
        lock.lock()
        guard let index = waiters.firstIndex(where: { $0.0 == url }) else {
            lock.unlock()
            return
        }
        let waiter = waiters.remove(at: index)
        lock.unlock()
        waiter.1.resume(returning: result)
    }
}

/// Controlled launch seam: one scripted process per generation.
private final class ControlledProcess: ManagedProcessLaunching, @unchecked Sendable {
    let generation: String
    /// The prepared target this launch was created for (resolved identity path).
    let target: URL

    private let events: (ManagedEvent) -> Void
    private let lock = NSLock()
    private var startResult: Result<ManagedLaunchSuccess, ManagedLaunchFailure>?
    private var startWaiters: [CheckedContinuation<Result<ManagedLaunchSuccess, ManagedLaunchFailure>, Never>] = []
    private var stopWaiters: [CheckedContinuation<ManagedStopOutcome, Never>] = []
    private var stopCallCount = 0
    private var running = false

    init(generation: String, target: URL, events: @escaping (ManagedEvent) -> Void) {
        self.generation = generation
        self.target = target
        self.events = events
    }

    var stopCalls: Int {
        lock.lock()
        defer { lock.unlock() }
        return stopCallCount
    }

    var isRunning: Bool {
        lock.lock()
        defer { lock.unlock() }
        return running
    }

    func start() async -> Result<ManagedLaunchSuccess, ManagedLaunchFailure> {
        await withCheckedContinuation { continuation in
            lock.lock()
            if let result = startResult {
                lock.unlock()
                continuation.resume(returning: result)
            } else {
                startWaiters.append(continuation)
                lock.unlock()
            }
        }
    }

    func stop() async -> ManagedStopOutcome {
        await withCheckedContinuation { continuation in
            lock.lock()
            stopCallCount += 1
            stopWaiters.append(continuation)
            lock.unlock()
        }
    }

    func completeStart(_ result: Result<ManagedLaunchSuccess, ManagedLaunchFailure>) {
        lock.lock()
        startResult = result
        if case .success = result {
            running = true
        }
        let waiters = startWaiters
        startWaiters = []
        lock.unlock()
        for waiter in waiters {
            waiter.resume(returning: result)
        }
    }

    func completeStop(_ outcome: ManagedStopOutcome) {
        lock.lock()
        let waiters = stopWaiters
        stopWaiters = []
        lock.unlock()
        for waiter in waiters {
            waiter.resume(returning: outcome)
        }
    }

    func emit(_ kind: ManagedEvent.Kind) {
        events(ManagedEvent(generation: generation, kind: kind))
    }
}

/// Records every controlled process so tests can drive and count launches.
private final class ControlledFactory: ManagedProcessFactory, @unchecked Sendable {
    private let lock = NSLock()
    private var processes: [ControlledProcess] = []
    private var creationHandler: ((ControlledProcess) -> Void)?

    /// Called once per created process, outside the factory lock; used to fail
    /// a launch that must not exist.
    var onCreate: ((ControlledProcess) -> Void)? {
        get {
            lock.lock()
            defer { lock.unlock() }
            return creationHandler
        }
        set {
            lock.lock()
            creationHandler = newValue
            lock.unlock()
        }
    }

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
        let process = ControlledProcess(generation: generation, target: target, events: events)
        lock.lock()
        processes.append(process)
        let handler = creationHandler
        lock.unlock()
        handler?(process)
        return process
    }
}
