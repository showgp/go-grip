import AppKit

/// The single native-alert slot of the app: the batch quantity confirmation
/// and the root failure report share it, so at most one alert is visible at a
/// time and a later request waits instead of stacking. Nothing here runs an
/// app-modal session: `NSAlert.runModal` suspends the main thread in a modal
/// run loop that does not serve unrelated events or tasks (AppKit documents
/// this), which starved the pasteboard exchange of an incoming Finder Services
/// request and lost it; a modeless alert window keeps the default run loop
/// servicing. This is the whole state the two app-level alerts need, not a
/// general modal queue.
@MainActor
final class AppAlertPresenter: BatchQuantityConfirming {
    /// Presents one alert and reports its dismissal response. Production shows
    /// a modeless window; tests inject a controlled seam.
    typealias Show = @MainActor (NSAlert, @escaping (NSApplication.ModalResponse) -> Void) -> Void

    private enum Job {
        case failure([OperationFailure])
        case confirmation(count: Int, continuation: CheckedContinuation<Bool, Never>)
        case guidance
    }

    private let injectedShow: Show?
    private var current: Job?
    private var currentAlert: NSAlert?
    private var currentRelay: ButtonRelay?
    private var currentObserver: NSObjectProtocol?
    private var pending: [Job] = []

    /// The production initializer presents modeless alerts; only the policy
    /// tests pass a `show` seam.
    init(show: Show? = nil) {
        injectedShow = show
    }

    /// Queues one failed operation or batch for its one native summary. The
    /// panel keeps listing the same report; a success publishes nothing.
    func presentFailureReport(_ failures: [OperationFailure]) {
        enqueue(.failure(failures))
    }

    /// Presents the short Finder/Services guidance: automatically once at
    /// first launch and again whenever the user opens it from the panel. It
    /// shares the single slot with the failure report and the quantity
    /// confirmation, so no two app-level alerts are visible at once and the
    /// app keeps servicing requests while it is shown.
    func presentServiceGuidance() {
        enqueue(.guidance)
    }

    /// Asks the user to approve opening `count` distinct targets and suspends
    /// until the alert is answered; the app, its panel and incoming Finder
    /// Services requests stay serviced while the confirmation is visible.
    func confirmOpening(count: Int) async -> Bool {
        await withCheckedContinuation { continuation in
            enqueue(.confirmation(count: count, continuation: continuation))
        }
    }

    private func enqueue(_ job: Job) {
        pending.append(job)
        presentNextIfIdle()
    }

    private func presentNextIfIdle() {
        guard current == nil, !pending.isEmpty else { return }
        let job = pending.removeFirst()
        current = job
        let alert = Self.alert(for: job)
        let respond: (NSApplication.ModalResponse) -> Void = { [weak self] response in
            self?.finish(response)
        }
        if let injectedShow {
            injectedShow(alert, respond)
        } else {
            presentModelessly(alert, respond: respond)
        }
    }

    /// One job at a time: the answer ends the visible alert before the next
    /// queued request is presented, and it can only be applied once.
    private func finish(_ response: NSApplication.ModalResponse) {
        guard let job = current else { return }
        current = nil
        if let observer = currentObserver {
            NotificationCenter.default.removeObserver(observer)
            currentObserver = nil
        }
        currentRelay = nil
        currentAlert?.window.orderOut(nil)
        currentAlert = nil
        if case .confirmation(_, let continuation) = job {
            continuation.resume(returning: response == .alertFirstButtonReturn)
        }
        presentNextIfIdle()
    }

    /// Modeless presentation: the alert window is made key and visible while
    /// the app's normal run loop keeps running, so the pasteboard exchange of
    /// a Services request arriving now is still serviced within its deadline.
    /// `layout()` finalizes the alert's lazy layout before showing, which the
    /// modal display paths would otherwise do themselves.
    private func presentModelessly(_ alert: NSAlert, respond: @escaping (NSApplication.ModalResponse) -> Void) {
        NSApp.activate(ignoringOtherApps: true)
        let relay = ButtonRelay { index in
            respond(index == 0 ? .alertFirstButtonReturn : .alertSecondButtonReturn)
        }
        for (index, button) in alert.buttons.enumerated() {
            button.target = relay
            button.action = #selector(ButtonRelay.fire(_:))
            button.tag = index
        }
        currentAlert = alert
        currentRelay = relay
        currentObserver = NotificationCenter.default.addObserver(
            forName: NSWindow.willCloseNotification,
            object: alert.window,
            queue: nil
        ) { [weak self] _ in
            Task { @MainActor [weak self] in self?.finish(.abort) }
        }
        alert.layout()
        alert.window.center()
        alert.window.makeKeyAndOrderFront(nil)
    }

    /// The one summary per failure report, the quantity confirmation and the
    /// Finder/Services guidance: exact user-facing wrapper text is looked up
    /// in the app's `Localizable.strings` (English development region, zh-Hans
    /// localization), while target paths and available reasons stay verbatim.
    private static func alert(for job: Job) -> NSAlert {
        let alert = NSAlert()
        alert.alertStyle = .warning
        switch job {
        case .failure(let failures):
            alert.messageText = failures.count == 1 ? failures[0].displayPath : "GoGrip"
            alert.informativeText = failures
                .map { "\($0.displayPath)\n\($0.reason)" }
                .joined(separator: "\n\n")
            alert.addButton(withTitle: NSLocalizedString("OK", comment: "Dismiss a failure summary"))
        case .confirmation(let count, _):
            alert.messageText = String(
                format: NSLocalizedString("Open %ld previews?", comment: "Quantity confirmation title; %ld is the number of targets"),
                count
            )
            alert.informativeText = String(
                format: NSLocalizedString(
                    "GoGrip will open %ld distinct targets and ask the default browser for each one.",
                    comment: "Quantity confirmation body; %ld is the number of targets"
                ),
                count
            )
            alert.addButton(withTitle: NSLocalizedString("Open", comment: "Confirm opening the batch"))
            alert.addButton(withTitle: NSLocalizedString("Cancel", comment: "Cancel the batch"))
        case .guidance:
            alert.messageText = Self.guidanceTitle
            alert.informativeText = Self.guidanceBody
            alert.addButton(withTitle: NSLocalizedString("OK", comment: "Dismiss the service guidance"))
        }
        return alert
    }

    /// The short first-use/revisited Finder instruction: how the Services
    /// command is invoked, where to check when it does not appear, the access
    /// it relies on, and the manual entry of the panel. It deliberately does
    /// not enable the service or claim the menu is now visible.
    private static var guidanceTitle: String {
        NSLocalizedString("Using GoGrip from Finder", comment: "Service guidance title")
    }

    private static var guidanceBody: String {
        [
            NSLocalizedString(
                "In Finder, select one or more folders or Markdown files, then choose Services → Open with GoGrip from the context menu; the preview opens in your default browser.",
                comment: "Service guidance: how to invoke the service"
            ),
            NSLocalizedString(
                "If the command is not there, check System Settings → Keyboard → Keyboard Shortcuts… → Services and make sure GoGrip's service is enabled.",
                comment: "Service guidance: where to check a missing or disabled service"
            ),
            NSLocalizedString(
                "GoGrip opens what macOS allows it to access. Ordinary targets need no extra permission; if macOS asks while opening a target, allow the GoGrip app to access it under Privacy & Security → Files and Folders.",
                comment: "Service guidance: applicable file-access check"
            ),
            NSLocalizedString(
                "You can also open a target from this panel with Open…, and reopen this guide any time with Help.",
                comment: "Service guidance: manual entry and how to revisit the guide"
            ),
        ].joined(separator: "\n\n")
    }
}

/// Retargets the alert buttons to the slot's own response path: the buttons
/// are the only dismiss controls besides the window's close button.
@MainActor
private final class ButtonRelay: NSObject {
    private let handler: (Int) -> Void

    init(handler: @escaping (Int) -> Void) {
        self.handler = handler
    }

    @objc func fire(_ sender: NSButton) {
        handler(sender.tag)
    }
}
