import Combine
import Foundation

/// One managed preview session as published by the coordinator: the single
/// fact source for target identity, generation, phase and observed state.
struct PreviewSession: Identifiable, Equatable {
    typealias ID = String

    enum Phase: Equatable {
        case starting
        case running
        case stopping
        case terminated
    }

    /// Normalized identity (resolved path); the same target opened through an
    /// alias is the same session.
    let id: ID
    /// The path the user selected, kept for display.
    let displayPath: String
    let mode: ManagedProcess.TargetMode
    var generation: String
    var phase: Phase
    var url: URL?
    var pid: Int32?
    var targetStatus: ManagedTargetSnapshot?
    var reloadStatus: ManagedReloadSnapshot?
    var lastFailure: String?
    /// Last rejected default-browser request, kept separate from the phase and
    /// the verified URL: the service stays running.
    var browserFailure: String?
}

/// Result of an open request: the session that owns (or already owned) the
/// target, or the reason the target could not be opened.
enum OpenOutcome: Equatable {
    case opened(PreviewSession.ID)
    case failed(OpenFailure)
}

struct OpenFailure: Equatable {
    let displayPath: String
    let reason: String
}

/// Result of preparing one batch before any record, child or browser request
/// exists: the distinct valid targets in selection order and the inputs that
/// cannot become a preview session.
struct BatchPreparation: Equatable {
    let targets: [PreviewTarget]
    let failures: [OpenFailure]
}

/// Prepares one selected target in the background. The production value is
/// `TargetPreparation.prepare`; behavior tests substitute a controlled result
/// so the quit admission window can be exercised deterministically.
typealias TargetPreparing = (URL) async -> Result<PreviewTarget, TargetPreparationFailure>

/// The single main-actor session fact source. It receives targets, prepares
/// their normalized identity in the background, and coordinates one owned
/// managed child per identity: starting requests share one launch, running
/// requests reuse the verified URL, and a reopen while stopping waits for the
/// real exit before a new generation may start. It never persists running
/// state and never restarts a session by itself.
@MainActor
final class PreviewSessionCoordinator: ObservableObject {
    @Published private(set) var sessions: [PreviewSession] = []

    private let factory: any ManagedProcessFactory
    private let prepare: TargetPreparing
    private var records: [PreviewSession.ID: Record] = [:]
    private var isAcceptingOpens = true

    init(
        factory: any ManagedProcessFactory,
        prepare: @escaping TargetPreparing = { await TargetPreparation.prepare($0) }
    ) {
        self.factory = factory
        self.prepare = prepare
    }

    /// One-way quit gate: after this, an open refuses instead of creating a
    /// child, including one whose background preparation completes later.
    func stopAcceptingNewOpens() {
        isAcceptingOpens = false
    }

    /// Called when an open has passed the admission checks and must wait for a
    /// previous generation to stop. Production leaves it unset; behavior tests
    /// use it as the deterministic synchronization point for the quit
    /// admission window (closing admission from inside the wait, not from a
    /// scheduler guess).
    var willWaitForStop: (() -> Void)?

    convenience init(executableURL: URL) {
        self.init(factory: FoundationManagedProcessFactory(executableURL: executableURL))
    }

    func session(_ id: PreviewSession.ID) -> PreviewSession? {
        sessions.first { $0.id == id }
    }

    /// Opens one selected target. Concurrent requests for the same normalized
    /// target share a single launch, a running session is reused with its
    /// verified URL, and an open during stopping waits for the real exit
    /// before a new generation starts.
    @discardableResult
    func open(_ selected: URL) async -> OpenOutcome {
        guard isAcceptingOpens else {
            return .failed(OpenFailure(
                displayPath: selected.standardizedFileURL.path,
                reason: Self.quittingReason
            ))
        }
        switch await prepare(selected) {
        case .failure(let failure):
            return .failed(OpenFailure(
                displayPath: selected.standardizedFileURL.path,
                reason: Self.describe(failure)
            ))
        case .success(let target):
            return await open(prepared: target)
        }
    }

    /// Prepares every selected input in the background and keeps one target per
    /// normalized identity, in selection order; inputs that cannot be opened
    /// are collected as failures. No record, child or browser request exists
    /// until the caller opens one of the returned targets.
    func prepareBatch(_ selected: [URL]) async -> BatchPreparation {
        var targets: [PreviewTarget] = []
        var identities: Set<String> = []
        var failures: [OpenFailure] = []
        for selectedURL in selected {
            switch await prepare(selectedURL) {
            case .success(let target):
                if identities.insert(target.identity).inserted {
                    targets.append(target)
                }
            case .failure(let failure):
                failures.append(OpenFailure(
                    displayPath: selectedURL.standardizedFileURL.path,
                    reason: Self.describe(failure)
                ))
            }
        }
        return BatchPreparation(targets: targets, failures: failures)
    }

    /// Opens one already-prepared target through the single record, launch and
    /// stop machinery; `open(_:)` is this after preparation, so a batch reuses
    /// its prepared identity instead of preparing the same target twice.
    @discardableResult
    func open(prepared target: PreviewTarget) async -> OpenOutcome {
        // Preparation can complete after the quit admission closed; the late
        // result must not create a record or start a child.
        guard isAcceptingOpens else {
            return .failed(OpenFailure(
                displayPath: target.displayPath,
                reason: Self.quittingReason
            ))
        }
        let record = records[target.identity] ?? makeRecord(target)
        while true {
            if record.phase == .stopping {
                willWaitForStop?()
                await ensureTermination(record).value
                if record.phase == .stopping {
                    // The stop could not confirm the real exit; starting a new
                    // generation now could overlap the old child, and retrying
                    // in this loop would never terminate.
                    return .failed(OpenFailure(
                        displayPath: record.displayPath,
                        reason: record.lastFailure ?? NSLocalizedString(
                            "The previous session has not stopped",
                            comment: "Reopen refused while a previous generation could not confirm its exit"
                        )
                    ))
                }
                continue
            }
            if let launch = record.activeLaunch {
                return await launch.task.value
            }
            switch record.phase {
            case .running:
                return .opened(record.id)
            case .starting, .stopping:
                // A starting record is always paired with an active launch and
                // stopping is handled above; re-check instead of spin.
                await Task.yield()
                continue
            case .terminated:
                // The quit admission can close while a reopen waits for a
                // previous generation to stop; a late reopen must not start a
                // child that no later stop would confirm.
                guard isAcceptingOpens else {
                    return .failed(OpenFailure(
                        displayPath: record.displayPath,
                        reason: Self.quittingReason
                    ))
                }
                return await startNewLaunch(record).task.value
            }
        }
    }

    /// Stops one owned session and returns only after the service has really
    /// exited. A stop that cannot confirm the exit keeps the session and its
    /// reason and can be retried; repeat stops share the same real finish.
    func stop(_ id: PreviewSession.ID) async {
        guard let record = records[id], record.phase != .terminated else { return }
        await ensureTermination(record).value
    }

    /// Stops every recorded session, including ones still starting. A target
    /// still being prepared is not yet a session; the caller that forbids new
    /// starts during termination (application quit) covers that window.
    func stopAll() async {
        let ids = records.values.filter { $0.phase != .terminated }.map(\.id)
        await withTaskGroup(of: Void.self) { group in
            for id in ids {
                group.addTask { @MainActor in
                    await self.stop(id)
                }
            }
        }
    }

    /// Records the last rejected default-browser request for one session. It
    /// is a user-visible operation failure only: the phase, the verified URL
    /// and ownership are untouched, and a successful retry clears it.
    func recordBrowserFailure(_ reason: String, for id: PreviewSession.ID) {
        guard let record = records[id] else { return }
        record.browserFailure = reason
        publish()
    }

    /// Clears a recorded browser failure after the system accepted a request.
    func clearBrowserFailure(for id: PreviewSession.ID) {
        guard let record = records[id], record.browserFailure != nil else { return }
        record.browserFailure = nil
        publish()
    }

    // MARK: - Session records

    private func makeRecord(_ target: PreviewTarget) -> Record {
        let record = Record(id: target.identity, displayPath: target.displayPath, mode: target.mode)
        records[record.id] = record
        return record
    }

    private func publish() {
        sessions = records.values
            .map { record in
                PreviewSession(
                    id: record.id,
                    displayPath: record.displayPath,
                    mode: record.mode,
                    generation: record.generation,
                    phase: record.phase,
                    url: record.url,
                    pid: record.pid,
                    targetStatus: record.targetStatus,
                    reloadStatus: record.reloadStatus,
                    lastFailure: record.lastFailure,
                    browserFailure: record.browserFailure
                )
            }
            .sorted { $0.id < $1.id }
    }

    // MARK: - Launching

    /// Starts one new generation for the record and returns its shared launch.
    /// The launch task is assigned before it can run, so concurrent opens join
    /// the same launch.
    private func startNewLaunch(_ record: Record) -> Launch {
        let generation = UUID().uuidString
        record.phase = .starting
        record.generation = generation
        record.url = nil
        record.pid = nil
        record.targetStatus = nil
        record.reloadStatus = nil
        record.lastFailure = nil
        record.browserFailure = nil
        record.activeProcess = nil
        record.activeLaunch = nil
        record.termination = nil
        record.stopIntent = false
        record.spawned = false
        record.spawnResolved = false
        record.spawnWaiters = []
        publish()

        let task = Task { @MainActor [weak self] in
            guard let self else {
                return OpenOutcome.failed(OpenFailure(
                    displayPath: record.displayPath,
                    reason: NSLocalizedString(
                        "The preview coordinator was released before the launch completed",
                        comment: "Internal open failure"
                    )
                ))
            }
            return await self.runLaunch(record, generation: generation)
        }
        let launch = Launch(generation: generation, task: task)
        record.activeLaunch = launch
        return launch
    }

    private func runLaunch(_ record: Record, generation: String) async -> OpenOutcome {
        guard record.generation == generation else {
            return .failed(OpenFailure(
                displayPath: record.displayPath,
                reason: NSLocalizedString(
                    "The launch was superseded before it started",
                    comment: "Internal open failure"
                )
            ))
        }
        if record.stopIntent {
            // Stopped before this launch began: no child is created, so no
            // adapter stop is needed and nothing can be left behind.
            clearLaunch(record, generation: generation)
            return .failed(OpenFailure(
                displayPath: record.displayPath,
                reason: NSLocalizedString(
                    "The session was stopped before it started",
                    comment: "Open cancelled by a stop before it began"
                )
            ))
        }
        guard isAcceptingOpens else {
            // The quit admission can close while this launch is still queued
            // for the main actor: the resuming start must not create a child,
            // and the cancelled record must not stay starting.
            finishTermination(record, generation: generation)
            return .failed(OpenFailure(displayPath: record.displayPath, reason: Self.quittingReason))
        }

        let process = factory.makeProcess(
            generation: generation,
            target: URL(fileURLWithPath: record.id),
            mode: record.mode,
            events: { [weak self, weak record] event in
                // Events must be applied in the order they are reported: a
                // synchronous main-actor emitter is applied inline, and any
                // other queue hops to the main actor. The production adapter
                // emits from its own queues, so this branch only serves
                // in-process emitters already isolated to the main actor.
                if Thread.isMainThread {
                    MainActor.assumeIsolated {
                        self?.handle(event, record: record)
                    }
                } else {
                    Task { @MainActor in
                        self?.handle(event, record: record)
                    }
                }
            }
        )
        record.activeProcess = process

        switch await process.start() {
        case .success(let success):
            return commitSuccess(record, generation: generation, process: process, success: success)
        case .failure(let failure):
            return commitFailure(record, generation: generation, failure: failure)
        }
    }

    /// Commits a verified startup exactly once. The child must still be
    /// running and the record must still be the same generation in starting,
    /// so a stop or an exit can never be overwritten by a late success.
    private func commitSuccess(
        _ record: Record,
        generation: String,
        process: any ManagedProcessLaunching,
        success: ManagedLaunchSuccess
    ) -> OpenOutcome {
        // A verified startup proves the owned child exists: a stop requested
        // around readiness must reach the adapter instead of waiting forever.
        resolveSpawn(record, generation: generation, spawned: true)

        if record.generation == generation,
           record.phase == .starting,
           !record.stopIntent,
           process.isRunning {
            record.phase = .running
            record.url = success.url
            record.pid = success.pid
            if record.reloadStatus == nil {
                record.reloadStatus = success.reload
            }
            clearLaunch(record, generation: generation)
            publish()
            return .opened(record.id)
        }

        if record.generation == generation, record.phase == .starting, !record.stopIntent {
            record.phase = .terminated
            record.url = nil
            record.activeProcess = nil
            if record.lastFailure == nil {
                record.lastFailure = NSLocalizedString(
                    "The preview service exited before it could be reported",
                    comment: "Managed service exited before its state could be published"
                )
            }
            clearLaunch(record, generation: generation)
            publish()
        }
        let reason = record.stopIntent
            ? NSLocalizedString(
                "The session was stopped while starting",
                comment: "Open cancelled by a stop during startup"
            )
            : record.lastFailure ?? NSLocalizedString(
                "The preview service exited before it could be reported",
                comment: "Managed service exited before its state could be published"
            )
        return .failed(OpenFailure(displayPath: record.displayPath, reason: reason))
    }

    private func commitFailure(
        _ record: Record,
        generation: String,
        failure: ManagedLaunchFailure
    ) -> OpenOutcome {
        resolveSpawn(record, generation: generation, spawned: record.spawned)
        let reason = Self.describe(failure)
        if record.generation == generation {
            if !record.stopIntent {
                record.phase = .terminated
                record.url = nil
                record.activeProcess = nil
                record.lastFailure = reason
            }
            clearLaunch(record, generation: generation)
            publish()
        }
        return .failed(OpenFailure(displayPath: record.displayPath, reason: reason))
    }

    // MARK: - Stopping

    /// Returns the termination task for the record, creating it on first use.
    /// The task closes the ownership writer through the adapter, waits for the
    /// real exit and applies the adapter's bounded SIGKILL fallback.
    private func ensureTermination(_ record: Record) -> Task<Void, Never> {
        if let termination = record.termination {
            return termination
        }
        let generation = record.generation
        record.stopIntent = true
        if record.phase != .terminated {
            record.phase = .stopping
            publish()
        }
        let task = Task { @MainActor [weak self] in
            guard let self else { return }
            await self.runTermination(record, generation: generation)
        }
        record.termination = task
        return task
    }

    private func runTermination(_ record: Record, generation: String) async {
        guard let process = record.activeProcess else {
            // No child was ever created for this generation.
            finishTermination(record, generation: generation)
            return
        }
        if !record.spawnResolved {
            await withCheckedContinuation { continuation in
                if record.spawnResolved {
                    continuation.resume()
                } else {
                    record.spawnWaiters.append(continuation)
                }
            }
        }
        guard record.spawned else {
            // The launch ended without establishing a child; the launch
            // result owns the session facts.
            finishTermination(record, generation: generation)
            return
        }

        switch await process.stop().result {
        case .exited, .killedAfterGrace:
            finishTermination(record, generation: generation)
        case .failed(let reason):
            guard record.generation == generation else { return }
            // The real exit is unconfirmed: keep the session, its failure and
            // the management entry so the user can retry the stop.
            record.termination = nil
            record.lastFailure = String(
                format: NSLocalizedString(
                    "Stop did not confirm exit: %@",
                    comment: "Stop failure; %@ is the raw adapter reason"
                ),
                reason
            )
            record.phase = .stopping
            publish()
        }
    }

    private func finishTermination(_ record: Record, generation: String) {
        guard record.generation == generation else { return }
        record.phase = .terminated
        record.url = nil
        record.activeProcess = nil
        clearLaunch(record, generation: generation)
        publish()
    }

    private func clearLaunch(_ record: Record, generation: String) {
        if record.activeLaunch?.generation == generation {
            record.activeLaunch = nil
        }
    }

    private func resolveSpawn(_ record: Record, generation: String, spawned: Bool) {
        guard record.generation == generation else { return }
        record.spawned = spawned
        record.spawnResolved = true
        let waiters = record.spawnWaiters
        record.spawnWaiters = []
        for waiter in waiters {
            waiter.resume()
        }
    }

    // MARK: - Generation-tagged events

    private func handle(_ event: ManagedEvent, record: Record?) {
        guard let record, record.generation == event.generation else { return }
        switch event.kind {
        case .spawned(let pid):
            record.pid = pid
            resolveSpawn(record, generation: event.generation, spawned: true)
            publish()
        case .reloadStatus(let snapshot):
            record.reloadStatus = snapshot
            publish()
        case .targetStatus(let snapshot):
            record.targetStatus = snapshot
            publish()
        case .fatal(let code, let message):
            // The service reports a fatal failure and then exits; the exit
            // event ends the running phase.
            record.lastFailure = "\(code): \(message)"
            publish()
        case .protocolViolation(let detail):
            record.lastFailure = String(
                format: NSLocalizedString(
                    "Protocol violation: %@",
                    comment: "Managed protocol violation; %@ is the raw diagnostic"
                ),
                detail
            )
            publish()
        case .exited(let status, _):
            guard !record.stopIntent, record.phase == .starting || record.phase == .running else { return }
            record.phase = .terminated
            record.url = nil
            record.activeProcess = nil
            if record.lastFailure == nil {
                record.lastFailure = String(
                    format: NSLocalizedString(
                        "The preview service exited unexpectedly (status %d)",
                        comment: "Unexpected service exit; %d is the raw exit status"
                    ),
                    status
                )
            }
            // An exit that beats the startup commit must not leave the
            // completed launch attached: a deliberate reopen has to be able
            // to start a new generation.
            clearLaunch(record, generation: event.generation)
            publish()
        }
    }

    // MARK: - Reasons

    private static let quittingReason = NSLocalizedString(
        "The application is quitting",
        comment: "Open refused because the app is quitting"
    )

    /// Conditional, class-neutral check path for failures that stopped the
    /// target from being read: it names the checks the user can perform (file
    /// or folder permissions, volume mount/sharing state, macOS Files and
    /// Folders) without asserting a cause and without demanding Full Disk
    /// Access. It is appended only to failures whose cause is about access.
    private static let accessCheckGuidance = NSLocalizedString(
        "Check the target's permissions, the volume's mount or sharing state, and — when macOS asked for authorization — System Settings → Privacy & Security → Files and Folders (GoGrip's entry).",
        comment: "Applicable checks for an access failure"
    )

    private static func describe(_ failure: TargetPreparationFailure) -> String {
        switch failure {
        case .unsupportedFile(let reason):
            return reason
        case .unavailable(let reason):
            return "\(reason) \(Self.accessCheckGuidance)"
        }
    }

    private static func describe(_ failure: ManagedLaunchFailure) -> String {
        switch failure.reason {
        case .spawnFailed(let detail):
            return String(
                format: NSLocalizedString(
                    "The preview service could not start: %@",
                    comment: "Managed launch failure; %@ is the raw process detail"
                ),
                detail
            )
        case .invalidFeedback(let detail):
            return String(
                format: NSLocalizedString(
                    "The preview service sent invalid startup feedback: %@",
                    comment: "Managed launch failure; %@ is the raw protocol detail"
                ),
                detail
            )
        case .fatal(let code, let message):
            let base = String(
                format: NSLocalizedString(
                    "The preview service reported %@: %@",
                    comment: "Managed fatal event; %@ is the raw code and message"
                ),
                code, message
            )
            return code == "target-unavailable"
                ? "\(base) \(Self.accessCheckGuidance)"
                : base
        case .exitedEarly(let status):
            return String(
                format: NSLocalizedString(
                    "The preview service exited during startup (status %d)",
                    comment: "Managed launch failure; %d is the raw exit status"
                ),
                status
            )
        case .timedOut:
            return NSLocalizedString(
                "The preview service did not become ready within the startup deadline",
                comment: "Managed launch failure"
            )
        case .readinessFailed(let detail):
            return String(
                format: NSLocalizedString(
                    "The preview service did not confirm readiness: %@",
                    comment: "Managed launch failure; %@ is the raw HTTP detail"
                ),
                detail
            )
        }
    }

    // MARK: - Records

    private final class Record {
        let id: PreviewSession.ID
        let displayPath: String
        let mode: ManagedProcess.TargetMode
        // A record without a launch yet has no live session; `startNewLaunch`
        // moves it to `.starting` together with its active launch.
        var phase: PreviewSession.Phase = .terminated
        var generation = ""
        var url: URL?
        var pid: Int32?
        var targetStatus: ManagedTargetSnapshot?
        var reloadStatus: ManagedReloadSnapshot?
        var lastFailure: String?
        var browserFailure: String?
        var activeProcess: (any ManagedProcessLaunching)?
        var activeLaunch: Launch?
        var termination: Task<Void, Never>?
        var stopIntent = false
        var spawned = false
        var spawnResolved = false
        var spawnWaiters: [CheckedContinuation<Void, Never>] = []

        init(id: PreviewSession.ID, displayPath: String, mode: ManagedProcess.TargetMode) {
            self.id = id
            self.displayPath = displayPath
            self.mode = mode
        }
    }

    private struct Launch {
        let generation: String
        let task: Task<OpenOutcome, Never>
    }
}
