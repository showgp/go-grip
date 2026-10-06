import Combine
import Foundation
import ServiceManagement

/// Requests the default browser for one verified preview URL. The production
/// implementation is `WorkspaceBrowserOpening`; behavior tests substitute a
/// controlled system result.
protocol BrowserOpening {
    /// Returns whether the system accepted the request, exactly like
    /// `NSWorkspace.open`. `true` does not prove that a page rendered.
    @discardableResult
    func open(_ url: URL) -> Bool
}

/// Asks the user to confirm a batch that exceeds the direct-open limit. The
/// production implementation is `AppAlertPresenter`; behavior tests substitute
/// a controlled result. The confirmation is awaited instead of blocking the
/// main thread, so an incoming Finder Services request stays serviceable while
/// the alert is visible.
@MainActor
protocol BatchQuantityConfirming {
    /// Returns true only when the user approved opening `count` targets.
    func confirmOpening(count: Int) async -> Bool
}

/// One failed user operation kept for the panel and the immediate native
/// alert: the target it concerned and the reason that was available. It is not
/// a process phase and never changes a session's running facts.
struct OperationFailure: Equatable, Hashable {
    let displayPath: String
    let reason: String
}

/// App-level orchestration on top of the single session fact source: opens a
/// selected batch through the coordinator, requests the default browser only
/// for a verified running session, reports the failures of one operation in a
/// single native alert, and owns the one-way termination admission used by the
/// menu bar host.
@MainActor
final class PreviewAppModel: ObservableObject {
    /// Distinct valid targets a batch opens without asking the user first.
    static let directOpenLimit = 5

    let coordinator: PreviewSessionCoordinator
    private let browser: any BrowserOpening
    private let confirmation: any BatchQuantityConfirming
    private let recentStore: RecentTargetsStore
    private let loginItem: any LoginItemRegistering

    /// Failures of the most recent failed operation or batch, kept for the
    /// panel and the one native alert. Empty means the last operation reported
    /// nothing; it is not a running fact and never changes a session.
    @Published private(set) var lastOperationFailures: [OperationFailure] = []
    /// The persisted recent targets, most recently used first: the exact
    /// normalized identities with the paths the user selected. Loading
    /// restores display data only; it never probes a target or starts
    /// anything.
    @Published private(set) var recentTargets: [RecentTarget] = []
    /// The system's current login-start registration for this app, read from
    /// the given service. It is the only value the panel shows for the login
    /// option; no desired state is persisted next to it.
    @Published private(set) var loginItemStatus: SMAppService.Status
    /// True from the first quit request on: new opens and browser requests are
    /// refused while the host waits for its owned services.
    private(set) var isTerminating = false

    init(
        coordinator: PreviewSessionCoordinator,
        browser: any BrowserOpening,
        confirmation: (any BatchQuantityConfirming)? = nil,
        recentStore: RecentTargetsStore,
        loginItem: any LoginItemRegistering
    ) {
        self.coordinator = coordinator
        self.browser = browser
        // The production host injects the same shared alert slot it uses for
        // the root failure reports; the isolated default only serves tests
        // that never exceed the direct-open limit.
        self.confirmation = confirmation ?? AppAlertPresenter()
        self.recentStore = recentStore
        self.loginItem = loginItem
        loginItemStatus = loginItem.status
        recentTargets = recentStore.load()
    }

    /// Opens every selected target as one batch: prepare and deduplicate in
    /// the background first, ask for confirmation when more than five distinct
    /// valid targets remain, then process the approved targets one at a time
    /// through the production coordinator. A failing target does not stop the
    /// rest, and the batch reports all failures once at the end. A closed quit
    /// admission stops the batch before any further child or browser request.
    ///
    /// `displayPathOverrides` is keyed by the prepared identity (the recorded
    /// path for a recent target): it keeps the user's selected path as the
    /// displayed path for the session and the failure context while the
    /// recorded normalized target stays the location being opened.
    func openBatch(
        _ selected: [URL],
        displayPathOverrides: [String: String] = [:]
    ) async {
        guard !isTerminating else { return }
        let preparation = await coordinator.prepareBatch(selected)
        guard !isTerminating else { return }
        if preparation.targets.count > Self.directOpenLimit {
            let approved = await confirmation.confirmOpening(count: preparation.targets.count)
            if !approved { return }
        }
        guard !isTerminating else { return }
        var failures = preparation.failures.map {
            OperationFailure(
                displayPath: displayPathOverrides[$0.displayPath] ?? $0.displayPath,
                reason: $0.reason
            )
        }
        for target in preparation.targets {
            guard !isTerminating else { break }
            let requested = PreviewTarget(
                identity: target.identity,
                displayPath: displayPathOverrides[target.identity] ?? target.displayPath,
                mode: target.mode
            )
            let outcome = await coordinator.open(prepared: requested)
            guard !isTerminating else { break }
            switch outcome {
            case .opened(let id):
                // The target is verified running now: its recency is updated
                // before the browser request, so a rejected request cannot
                // revoke a successfully established target.
                recordRecent(identity: id, displayPath: requested.displayPath, mode: requested.mode)
                if let failure = requestBrowser(id) {
                    failures.append(failure)
                }
            case .failed(let failure):
                failures.append(OperationFailure(displayPath: failure.displayPath, reason: failure.reason))
            }
        }
        if !failures.isEmpty {
            lastOperationFailures = failures
        }
    }

    /// Requests the default browser again for a running session's verified
    /// URL. A rejected request keeps the session, its URL and its management
    /// entry; a request the system accepted clears the recorded failure. The
    /// explicit reopen of a running target also refreshes its recency, unless
    /// the quit admission has already closed.
    func openBrowser(_ id: PreviewSession.ID) {
        guard !isTerminating else { return }
        if let session = coordinator.session(id), session.phase == .running, session.url != nil {
            recordRecent(identity: session.id, displayPath: session.displayPath, mode: session.mode)
        }
        if let failure = requestBrowser(id) {
            lastOperationFailures = [failure]
        }
    }

    /// Reopens one recent target through the same production batch entry the
    /// Finder service and the open panel use: preparation, normalized
    /// identity, session reuse, quit admission and the one failure report stay
    /// shared. The location is the recorded normalized target, while the
    /// recorded user-selected path stays the displayed path and the failure
    /// context; a changed or removed symbolic link is never re-resolved to a
    /// different target.
    func reopenRecent(_ target: RecentTarget) async {
        await openBatch(
            [URL(fileURLWithPath: target.identity)],
            displayPathOverrides: [target.identity: target.displayPath]
        )
    }

    /// Clears the recent list only: running sessions, their verified URLs and
    /// their management entries stay untouched, and later launches stay empty
    /// until the user explicitly opens a target again.
    func clearRecentTargets() {
        recentStore.clear()
        recentTargets = []
    }

    /// Re-reads the real login-start registration. The panel calls this when
    /// it becomes visible, so a change made in System Settings is reflected
    /// without an app restart. The read never registers or unregisters.
    func refreshLoginItemStatus() {
        loginItemStatus = loginItem.status
    }

    /// Applies an explicit user choice through the real system operation and
    /// then shows the status the system actually reports — never the
    /// requested value as an outcome. A thrown operation keeps the real
    /// status and publishes the available reason through the one root report;
    /// a registration that still needs approval is a real state, not an
    /// error. A framework-reported `notFound` before any registration is the
    /// same user-visible state as `notRegistered`, so the explicit action
    /// attempts the real `register()` and the re-read status decides the
    /// result. Nothing here registers or unregisters without this call.
    func setLoginItemEnabled(_ enabled: Bool) {
        do {
            if enabled {
                try loginItem.register()
            } else {
                try loginItem.unregister()
            }
        } catch {
            lastOperationFailures = [OperationFailure(
                displayPath: NSLocalizedString("Launch at login", comment: "Login item failure context"),
                reason: String(
                    format: NSLocalizedString(
                        "Could not change the login startup setting: %@",
                        comment: "Login item operation failed; %@ is the system reason"
                    ),
                    error.localizedDescription
                )
            )]
        }
        refreshLoginItemStatus()
    }

    /// Opens the system Login Items settings from an explicit user action,
    /// used when the system asks for approval. It never approves, retries or
    /// changes anything by itself.
    func openLoginItemsSettings() {
        loginItem.openLoginItemsSettings()
    }

    /// Requests the browser for one verified running session and returns the
    /// failure instead of publishing it, so a batch can consolidate every
    /// failure of the batch into its single report.
    private func requestBrowser(_ id: PreviewSession.ID) -> OperationFailure? {
        guard !isTerminating,
              let session = coordinator.session(id),
              session.phase == .running,
              let url = session.url else { return nil }
        if browser.open(url) {
            coordinator.clearBrowserFailure(for: id)
            return nil
        }
        let reason = NSLocalizedString(
            "The system could not open the default browser for the preview",
            comment: "Browser request rejected by the system"
        )
        coordinator.recordBrowserFailure(reason, for: id)
        return OperationFailure(displayPath: session.displayPath, reason: reason)
    }

    /// Updates the persisted recency for one successfully established or
    /// explicitly reopened target. Only the verified target identity with the
    /// user-selected path and kind is stored; no running fact is persisted.
    private func recordRecent(identity: String, displayPath: String, mode: ManagedProcess.TargetMode) {
        recentTargets = recentStore.record(identity: identity, displayPath: displayPath, mode: mode)
    }

    func stop(_ id: PreviewSession.ID) async {
        await coordinator.stop(id)
    }

    func stopAll() async {
        await coordinator.stopAll()
    }

    /// First step of an application quit: refuse new opens and browser
    /// requests, then report whether anything still has to stop. Idempotent;
    /// a retried quit reuses the same closed admission.
    func beginTermination() -> Bool {
        isTerminating = true
        coordinator.stopAcceptingNewOpens()
        return coordinator.sessions.allSatisfy { $0.phase == .terminated }
    }

    /// Second step of an application quit: wait for every recorded owned
    /// service, including ones still starting, and allow the exit only after
    /// each one confirmed its real exit. An unconfirmed stop keeps the session
    /// and its management entry and refuses the exit.
    func completeTermination() async -> Bool {
        await coordinator.stopAll()
        let unfinished = coordinator.sessions.filter { $0.phase != .terminated }
        guard unfinished.isEmpty else {
            if let pending = unfinished.first {
                lastOperationFailures = [OperationFailure(
                    displayPath: pending.displayPath,
                    reason: pending.lastFailure ?? NSLocalizedString(
                        "The preview service did not confirm its exit",
                        comment: "Quit wait for an owned service"
                    )
                )]
            }
            return false
        }
        return true
    }
}
