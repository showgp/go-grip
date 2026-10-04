import Foundation

/// Timing for one launch: the internal startup deadline covering machine
/// feedback and the single readiness HEAD, and the termination grace before
/// the exact owned child is SIGKILLed.
struct ManagedProcessTiming: Equatable {
    var startupDeadline: TimeInterval = 15
    var terminationGrace: TimeInterval = 4

    static let `default` = ManagedProcessTiming()
}

/// Generation-tagged host event for one owned launch. Every event belongs to
/// the adapter's generation, and `spawned` is delivered on the launch queue
/// right after the child exists, before readiness validation starts.
struct ManagedEvent: Equatable {
    let generation: String
    let kind: Kind

    enum Kind: Equatable {
        case spawned(pid: Int32)
        case reloadStatus(ManagedReloadSnapshot)
        case targetStatus(ManagedTargetSnapshot)
        case fatal(code: String, message: String)
        case protocolViolation(detail: String)
        case exited(exitCode: Int32, reason: Process.TerminationReason)
    }
}

/// Verified startup result: the child is still running and answered the
/// readiness HEAD with the matching generation.
struct ManagedLaunchSuccess: Equatable {
    let generation: String
    let url: URL
    let reload: ManagedReloadSnapshot
    let pid: Int32
}

/// Failed launch, reported only after the owned child (if any) has exited.
/// `diagnostics` is the retained stderr tail.
struct ManagedLaunchFailure: Error, Equatable {
    let generation: String
    let reason: Reason
    let diagnostics: String

    enum Reason: Equatable {
        case spawnFailed(String)
        case invalidFeedback(String)
        case fatal(code: String, message: String)
        case exitedEarly(status: Int32)
        case timedOut
        case readinessFailed(detail: String)
    }
}

/// Stop result for the exact owned child.
struct ManagedStopOutcome: Equatable {
    let generation: String
    let result: ManagedStopResult
}

enum ManagedStopResult: Equatable {
    case exited(exitCode: Int32, reason: Process.TerminationReason)
    case killedAfterGrace(exitCode: Int32, reason: Process.TerminationReason)
    case failed(String)
}

/// Bounded stderr tail: keeps the newest `capacity` bytes appended so far.
/// Arrival scheduling in a real launch is inherently best-effort (a read can
/// still be outstanding when a failure completes), so the cap and suffix
/// semantics are pinned by direct tests of this type.
struct ManagedDiagnosticsTail {
    static let capacity = 64 * 1024

    private var data = Data()

    mutating func append(_ chunk: Data) {
        data.append(chunk)
        if data.count > Self.capacity {
            data.removeFirst(data.count - Self.capacity)
        }
    }

    var text: String {
        String(decoding: data, as: UTF8.self)
    }
}

/// Production Foundation adapter for one managed Go preview process: one
/// launch attempt, one owned child, generation-tagged events. It never builds
/// another session dictionary, guesses a port or reuses the legacy manager's
/// success rules.
final class ManagedProcess: @unchecked Sendable {
    enum TargetMode: Equatable {
        case directory
        case file
    }

    let generation: String

    private let executableURL: URL
    private let targetURL: URL
    private let mode: TargetMode
    private let timing: ManagedProcessTiming
    private let headSessionConfiguration: URLSessionConfiguration?
    private let events: (ManagedEvent) -> Void

    /// Every launch in this process shares one serial critical section: pipe
    /// creation, the FD_CLOEXEC check and Process.run must not interleave with
    /// another spawn, or a not-yet-protected descriptor could leak into it.
    private static let launchQueue = DispatchQueue(label: "com.showgp.go-grip.managed-launch")

    private let stateLock = NSLock()
    private var process: Process?
    private var stdinWriter: FileHandle?
    private var ownedPID: Int32?
    private var startInstant: DispatchTime?
    private var stdoutEnded = false
    private var diagnostics = ManagedDiagnosticsTail()
    private var readyURL: URL?
    private var pendingFailure: ManagedLaunchFailure.Reason?
    private var startupOutcome: Result<ManagedLaunchSuccess, ManagedLaunchFailure>?
    private var startupContinuation: CheckedContinuation<Result<ManagedLaunchSuccess, ManagedLaunchFailure>, Never>?
    private var startupSucceeded = false
    private var runtimeFailureEmitted = false
    private var exitStatus: Int32?
    private var exitReason: Process.TerminationReason?
    private var exitEventEmitted = false
    private var terminationArmed = false
    private var sigkillSent = false
    private var stopWaiters: [CheckedContinuation<ManagedStopOutcome, Never>] = []
    private var recordedStopOutcome: ManagedStopOutcome?

    init(
        executableURL: URL,
        generation: String,
        target: URL,
        mode: TargetMode,
        timing: ManagedProcessTiming = .default,
        headSessionConfiguration: URLSessionConfiguration? = nil,
        events: @escaping (ManagedEvent) -> Void = { _ in }
    ) {
        self.executableURL = executableURL
        self.generation = generation
        self.targetURL = target
        self.mode = mode
        self.timing = timing
        self.headSessionConfiguration = headSessionConfiguration
        self.events = events
    }

    /// Synchronous scoped access to the state lock, safe to call from async
    /// functions where NSLock's direct lock() is unavailable in Swift 6 mode.
    private func withStateLock<T>(_ body: () -> T) -> T {
        stateLock.lock()
        defer { stateLock.unlock() }
        return body()
    }

    /// PID of the exact owned child, available from the `spawned` event on.
    var pid: Int32? {
        withStateLock { ownedPID }
    }

    var isRunning: Bool {
        withStateLock { process?.isRunning ?? false }
    }

    /// Launches the managed child and validates readiness within the startup
    /// deadline. Returns only after a failed launch's child has exited.
    func start() async -> Result<ManagedLaunchSuccess, ManagedLaunchFailure> {
        let spawnFailure = await withCheckedContinuation { continuation in
            Self.launchQueue.async { continuation.resume(returning: self.spawn()) }
        }
        if let reason = spawnFailure {
            return .failure(ManagedLaunchFailure(generation: generation, reason: reason, diagnostics: ""))
        }
        return await withCheckedContinuation { continuation in
            stateLock.lock()
            if let outcome = startupOutcome {
                stateLock.unlock()
                continuation.resume(returning: outcome)
            } else {
                startupContinuation = continuation
                stateLock.unlock()
            }
        }
    }

    /// Closes the ownership writer, waits for the real exit and, after the
    /// termination grace, SIGKILLs this exact owned child. The result carries
    /// the real exit status.
    func stop() async -> ManagedStopOutcome {
        if let recorded = withStateLock({ recordedStopOutcome }) {
            return recorded
        }

        return await withCheckedContinuation { continuation in
            stateLock.lock()
            if let recorded = recordedStopOutcome {
                stateLock.unlock()
                continuation.resume(returning: recorded)
                return
            }
            guard process != nil else {
                let outcome = ManagedStopOutcome(generation: generation, result: .failed("no owned child process"))
                recordedStopOutcome = outcome
                stateLock.unlock()
                continuation.resume(returning: outcome)
                return
            }
            if let status = exitStatus, let reason = exitReason {
                let outcome = ManagedStopOutcome(generation: generation, result: .exited(exitCode: status, reason: reason))
                recordedStopOutcome = outcome
                stateLock.unlock()
                continuation.resume(returning: outcome)
                return
            }
            stopWaiters.append(continuation)
            stateLock.unlock()
            armTermination()
        }
    }

    // MARK: - Launch critical section

    /// Runs inside the shared launch critical section; returns the failure
    /// reason or nil when the child was spawned.
    private func spawn() -> ManagedLaunchFailure.Reason? {
        let stdinPipe = Pipe()
        let stdoutPipe = Pipe()
        let stderrPipe = Pipe()

        // Set and check FD_CLOEXEC on the non-stdio ends the host keeps: the
        // ownership writer and the two readers. An inherited writer would mask
        // owner loss (EOF) for this child, and a leaked reader would keep the
        // pipe open past the child's exit.
        guard setCloseOnExec(stdinPipe.fileHandleForWriting.fileDescriptor),
              setCloseOnExec(stdoutPipe.fileHandleForReading.fileDescriptor),
              setCloseOnExec(stderrPipe.fileHandleForReading.fileDescriptor)
        else {
            closePipes(stdinPipe, stdoutPipe, stderrPipe)
            return .spawnFailed("could not set FD_CLOEXEC on the ownership pipe")
        }

        let process = Process()
        process.executableURL = executableURL
        var arguments = ["--managed", generation]
        if mode == .directory {
            arguments.append("-r")
        }
        arguments.append(contentsOf: ["--", targetURL.path])
        process.arguments = arguments
        process.standardInput = stdinPipe
        process.standardOutput = stdoutPipe
        process.standardError = stderrPipe
        process.terminationHandler = { [weak self] process in
            self?.handleExit(status: process.terminationStatus, reason: process.terminationReason)
        }

        do {
            try process.run()
        } catch {
            closePipes(stdinPipe, stdoutPipe, stderrPipe)
            return .spawnFailed(error.localizedDescription)
        }

        stateLock.lock()
        self.process = process
        self.stdinWriter = stdinPipe.fileHandleForWriting
        self.ownedPID = process.processIdentifier
        self.startInstant = DispatchTime.now()
        stateLock.unlock()

        emit(.spawned(pid: process.processIdentifier))
        readFrames(from: stdoutPipe.fileHandleForReading, decoder: ManagedNDJSONDecoder(generation: generation))
        readDiagnostics(from: stderrPipe.fileHandleForReading)
        DispatchQueue.global().asyncAfter(deadline: .now() + timing.startupDeadline) { [weak self] in
            self?.resolveTimeout()
        }
        return nil
    }

    private func setCloseOnExec(_ descriptor: Int32) -> Bool {
        guard descriptor >= 0 else { return false }
        guard fcntl(descriptor, F_SETFD, FD_CLOEXEC) != -1 else { return false }
        return fcntl(descriptor, F_GETFD) & FD_CLOEXEC != 0
    }

    /// Releases both ends of every pipe in a set whose launch produced no
    /// child. FileHandle.close() is idempotent with deallocation.
    private func closePipes(_ pipes: Pipe...) {
        for pipe in pipes {
            try? pipe.fileHandleForWriting.close()
            try? pipe.fileHandleForReading.close()
        }
    }

    // MARK: - Feedback consumption

    private func readFrames(from handle: FileHandle, decoder: ManagedNDJSONDecoder) {
        DispatchQueue(label: "com.showgp.go-grip.managed-stdout").async { [weak self] in
            guard let self else { return }
            var decodingHalted = false
            while true {
                let chunk = handle.availableData
                if chunk.isEmpty { break }
                // After a violation the stream is no longer trusted for
                // state, but it keeps being drained so the child never blocks
                // on a full stdout pipe while it shuts down.
                if decodingHalted { continue }
                do {
                    for frame in try decoder.feed(chunk) {
                        self.handle(frame)
                    }
                } catch {
                    self.handleProtocolViolation(error)
                    decodingHalted = true
                }
            }
            // EOF: release the read end now. The adapter keeps the Process for
            // its real exit status, so deallocation must not be what closes
            // this descriptor.
            try? handle.close()
            self.concludeStdout(decoder: decoder, decodingHalted: decodingHalted)
        }
    }

    private func readDiagnostics(from handle: FileHandle) {
        DispatchQueue(label: "com.showgp.go-grip.managed-stderr").async { [weak self] in
            guard let self else { return }
            while true {
                let chunk = handle.availableData
                if chunk.isEmpty { break }
                self.stateLock.lock()
                self.diagnostics.append(chunk)
                self.stateLock.unlock()
            }
            // EOF: release the read end now, for the same reason as stdout.
            try? handle.close()
        }
    }

    private func handle(_ frame: ManagedFrame) {
        switch frame {
        case .ready(let urlString, let reload):
            stateLock.lock()
            let seen = readyURL != nil
            stateLock.unlock()
            guard !seen else { return }
            guard let url = URL(string: urlString),
                  Self.isLoopbackHTTP(url),
                  let readinessURL = Self.readinessURL(for: url)
            else {
                failStartup(.invalidFeedback("ready URL is not an HTTP loopback address: \(urlString)"))
                return
            }
            stateLock.lock()
            readyURL = url
            stateLock.unlock()
            Task { [weak self] in
                await self?.validateReadiness(at: readinessURL, readyURL: url, reload: reload)
            }
        case .reloadStatus(let snapshot):
            emit(.reloadStatus(snapshot))
        case .targetStatus(let snapshot):
            emit(.targetStatus(snapshot))
        case .fatal(let code, let message):
            if !failStartup(.fatal(code: code, message: message)) {
                stateLock.lock()
                let delivered = startupSucceeded
                stateLock.unlock()
                if delivered {
                    emit(.fatal(code: code, message: message))
                }
            }
        }
    }

    /// A decoder/validation violation before startup fails the launch; after a
    /// delivered success it is a runtime failure: the session is reported with
    /// the violation and its exact owned child is terminated.
    private func handleProtocolViolation(_ error: Error) {
        let detail = String(describing: error)
        if failStartup(.invalidFeedback(detail)) {
            return
        }
        let delivered = withStateLock {
            guard startupSucceeded && !runtimeFailureEmitted else { return false }
            runtimeFailureEmitted = true
            return true
        }
        if delivered {
            emit(.protocolViolation(detail: detail))
            armTermination()
        }
    }

    private func concludeStdout(decoder: ManagedNDJSONDecoder, decodingHalted: Bool) {
        if !decodingHalted {
            do {
                try decoder.finish()
            } catch {
                handleProtocolViolation(error)
            }
        }

        stateLock.lock()
        stdoutEnded = true
        var completionNeeded = false
        if pendingFailure == nil && startupOutcome == nil, let status = exitStatus {
            pendingFailure = readyURL != nil
                ? .readinessFailed(detail: "the owned process exited before readiness was confirmed")
                : .exitedEarly(status: status)
            completionNeeded = true
        }
        stateLock.unlock()
        if completionNeeded {
            completeStartupWithPendingFailure()
        }
    }

    // MARK: - Readiness validation

    private func validateReadiness(at readinessURL: URL, readyURL: URL, reload: ManagedReloadSnapshot) async {
        let remaining = remainingStartupTime()
        guard remaining > 0 else {
            failStartup(.timedOut)
            return
        }
        switch await headReadiness(at: readinessURL, timeout: remaining) {
        case .some(let detail):
            failStartup(.readinessFailed(detail: detail))
        case .none:
            if !commitStartupSuccess(url: readyURL, reload: reload) {
                failStartup(.readinessFailed(detail: "the owned process exited before readiness was confirmed"))
            }
        }
    }

    /// One lightweight HEAD against the same origin, confirming the generation
    /// header, HTTP 204 and that no redirect to another origin was followed.
    /// Returns nil on success and the failure detail otherwise.
    private func headReadiness(at url: URL, timeout: TimeInterval) async -> String? {
        let configuration = headSessionConfiguration ?? Self.defaultHeadSessionConfiguration()
        let delegate = RedirectRefusingDelegate()
        let session = URLSession(configuration: configuration, delegate: delegate, delegateQueue: nil)
        defer { session.finishTasksAndInvalidate() }

        var request = URLRequest(url: url)
        request.httpMethod = "HEAD"
        request.timeoutInterval = max(0.05, timeout)
        do {
            let (_, response) = try await session.data(for: request)
            guard let http = response as? HTTPURLResponse else {
                return "readiness response is not HTTP"
            }
            guard http.statusCode == 204 else {
                return "readiness HEAD returned HTTP \(http.statusCode)"
            }
            let received = http.value(forHTTPHeaderField: "X-GoGrip-Generation")
            guard received == generation else {
                return "readiness generation header \(received ?? "missing") does not match \(generation)"
            }
            guard let finalURL = http.url, finalURL.host == url.host, finalURL.port == url.port else {
                return "readiness response came from a different origin"
            }
            return nil
        } catch {
            return "readiness HEAD failed: \(error.localizedDescription)"
        }
    }

    private static func defaultHeadSessionConfiguration() -> URLSessionConfiguration {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.connectionProxyDictionary = [:]
        configuration.waitsForConnectivity = false
        configuration.requestCachePolicy = .reloadIgnoringLocalCacheData
        configuration.timeoutIntervalForRequest = ManagedProcessTiming.default.startupDeadline
        configuration.timeoutIntervalForResource = ManagedProcessTiming.default.startupDeadline
        return configuration
    }

    private static func isLoopbackHTTP(_ url: URL) -> Bool {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
              components.scheme?.lowercased() == "http",
              components.host == "127.0.0.1",
              let port = components.port, (1...65535).contains(port)
        else {
            return false
        }
        return true
    }

    private static func readinessURL(for url: URL) -> URL? {
        var components = URLComponents()
        components.scheme = "http"
        components.host = "127.0.0.1"
        components.port = url.port
        components.path = "/__gogrip/ready"
        return components.url
    }

    private func remainingStartupTime() -> TimeInterval {
        guard let instant = withStateLock({ startInstant }) else { return 0 }
        let elapsed = Double(DispatchTime.now().uptimeNanoseconds - instant.uptimeNanoseconds) / 1_000_000_000
        return timing.startupDeadline - elapsed
    }

    // MARK: - Startup resolution and owned-child termination

    @discardableResult
    private func failStartup(_ reason: ManagedLaunchFailure.Reason) -> Bool {
        stateLock.lock()
        guard pendingFailure == nil && startupOutcome == nil else {
            stateLock.unlock()
            return false
        }
        pendingFailure = reason
        let hasChild = process != nil
        let exited = exitStatus != nil
        stateLock.unlock()

        if hasChild {
            armTermination()
            if exited {
                completeStartupWithPendingFailure()
            }
        } else {
            completeStartupWithPendingFailure()
        }
        return true
    }

    private func resolveTimeout() {
        stateLock.lock()
        let outstanding = pendingFailure == nil && startupOutcome == nil
        stateLock.unlock()
        guard outstanding else { return }
        failStartup(.timedOut)
    }

    private func completeStartupWithPendingFailure() {
        stateLock.lock()
        guard let reason = pendingFailure, startupOutcome == nil else {
            stateLock.unlock()
            return
        }
        let diagnostics = self.diagnostics.text
        stateLock.unlock()
        completeStartupFailure(ManagedLaunchFailure(
            generation: generation,
            reason: reason,
            diagnostics: diagnostics
        ))
    }

    /// Commits a verified startup exactly once, re-checking every success
    /// precondition under the same lock that failure and exit transitions use:
    /// a concurrent timeout, violation or exit can therefore never be
    /// overwritten by a success. The caller reported HEAD 204 with the
    /// matching generation; the child must still be running.
    private func commitStartupSuccess(url: URL, reload: ManagedReloadSnapshot) -> Bool {
        var continuation: CheckedContinuation<Result<ManagedLaunchSuccess, ManagedLaunchFailure>, Never>?
        var success: ManagedLaunchSuccess?
        withStateLock {
            guard startupOutcome == nil, pendingFailure == nil, exitStatus == nil,
                  let process, process.isRunning, let pid = ownedPID else { return }
            let result = ManagedLaunchSuccess(generation: generation, url: url, reload: reload, pid: pid)
            startupOutcome = .success(result)
            startupSucceeded = true
            success = result
            continuation = startupContinuation
            startupContinuation = nil
        }
        if let success {
            continuation?.resume(returning: .success(success))
        }
        return success != nil
    }

    /// Completes a failed launch exactly once, after the owned child (if any)
    /// has exited. Success is committed through `commitStartupSuccess`.
    private func completeStartupFailure(_ failure: ManagedLaunchFailure) {
        var continuation: CheckedContinuation<Result<ManagedLaunchSuccess, ManagedLaunchFailure>, Never>?
        withStateLock {
            guard startupOutcome == nil else { return }
            startupOutcome = .failure(failure)
            continuation = startupContinuation
            startupContinuation = nil
        }
        continuation?.resume(returning: .failure(failure))
    }

    private func handleExit(status: Int32, reason: Process.TerminationReason) {
        var stopOutcome: ManagedStopOutcome?
        var stopContinuations: [CheckedContinuation<ManagedStopOutcome, Never>] = []
        var exitEvent: ManagedEvent.Kind?
        var completionNeeded = false

        stateLock.lock()
        guard exitStatus == nil else {
            stateLock.unlock()
            return
        }
        exitStatus = status
        exitReason = reason
        let writer = stdinWriter
        stdinWriter = nil

        if startupSucceeded && !exitEventEmitted {
            exitEventEmitted = true
            exitEvent = .exited(exitCode: status, reason: reason)
        }
        if !stopWaiters.isEmpty {
            let result: ManagedStopResult = sigkillSent
                ? .killedAfterGrace(exitCode: status, reason: reason)
                : .exited(exitCode: status, reason: reason)
            let outcome = ManagedStopOutcome(generation: generation, result: result)
            recordedStopOutcome = outcome
            stopOutcome = outcome
            stopContinuations = stopWaiters
            stopWaiters = []
        }
        if pendingFailure != nil {
            completionNeeded = true
        } else if startupOutcome == nil && stdoutEnded {
            pendingFailure = readyURL != nil
                ? .readinessFailed(detail: "the owned process exited before readiness was confirmed")
                : .exitedEarly(status: status)
            completionNeeded = true
        }
        stateLock.unlock()

        // The child is gone, so the ownership writer is released on every
        // exit path, including ones that never armed termination.
        try? writer?.close()

        if let exitEvent {
            emit(exitEvent)
        }
        for continuation in stopContinuations {
            continuation.resume(returning: stopOutcome!)
        }
        if completionNeeded {
            completeStartupWithPendingFailure()
        }
    }

    /// Closes the ownership writer once, then arms the termination grace: an
    /// exact owned child still alive after the grace is SIGKILLed. Normal
    /// stops rely on the child observing EOF; the fallback is bounded and
    /// never targets anything but this child's PID.
    private func armTermination() {
        stateLock.lock()
        let writer = stdinWriter
        stdinWriter = nil
        let alreadyExited = exitStatus != nil
        let shouldArm = !terminationArmed
        terminationArmed = true
        let pid = ownedPID
        stateLock.unlock()

        try? writer?.close()
        guard shouldArm, let pid, !alreadyExited else { return }
        DispatchQueue.global().asyncAfter(deadline: .now() + timing.terminationGrace) { [weak self] in
            guard let self else { return }
            self.stateLock.lock()
            let alive = self.exitStatus == nil
            if alive {
                self.sigkillSent = true
            }
            self.stateLock.unlock()
            guard alive else { return }
            kill(pid, SIGKILL)
        }
    }

    private func emit(_ kind: ManagedEvent.Kind) {
        events(ManagedEvent(generation: generation, kind: kind))
    }
}

/// Refuses every redirect that would move the readiness check to another
/// origin; the 3xx response itself then fails validation.
private final class RedirectRefusingDelegate: NSObject, URLSessionTaskDelegate {
    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        willPerformHTTPRedirection response: HTTPURLResponse,
        newRequest request: URLRequest,
        completionHandler: @escaping (URLRequest?) -> Void
    ) {
        completionHandler(nil)
    }
}
