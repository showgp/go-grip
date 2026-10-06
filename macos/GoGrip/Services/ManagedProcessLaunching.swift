import Foundation

/// Minimal launch seam for one owned managed child. The production
/// implementation is `ManagedProcess`; behavior tests substitute a controlled
/// process so lifecycle races are deterministic. It adds no second process
/// abstraction: it exposes exactly what the coordinator consumes.
protocol ManagedProcessLaunching: AnyObject {
    /// Generation assigned to this launch; events and results carry the same value.
    var generation: String { get }
    /// Whether the exact owned child is still running.
    var isRunning: Bool { get }
    /// Launches the child and validates readiness within the adapter deadline.
    func start() async -> Result<ManagedLaunchSuccess, ManagedLaunchFailure>
    /// Closes the ownership writer, waits for the real exit and applies the
    /// bounded SIGKILL fallback for this exact child.
    func stop() async -> ManagedStopOutcome
}

extension ManagedProcess: ManagedProcessLaunching {}

/// Creates the owned process for one launch; the coordinator assigns the
/// generation and consumes the generation-tagged events.
protocol ManagedProcessFactory {
    func makeProcess(
        generation: String,
        target: URL,
        mode: ManagedProcess.TargetMode,
        events: @escaping (ManagedEvent) -> Void
    ) -> ManagedProcessLaunching
}

/// Production factory: the Foundation adapter that launches the bundled Go tool.
struct FoundationManagedProcessFactory: ManagedProcessFactory {
    let executableURL: URL

    func makeProcess(
        generation: String,
        target: URL,
        mode: ManagedProcess.TargetMode,
        events: @escaping (ManagedEvent) -> Void
    ) -> ManagedProcessLaunching {
        ManagedProcess(
            executableURL: executableURL,
            generation: generation,
            target: target,
            mode: mode,
            events: events
        )
    }
}
