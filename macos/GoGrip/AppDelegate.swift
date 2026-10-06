import AppKit
import Combine
import SwiftUI

/// Strongly-held AppKit root of the menu bar host: it owns the single
/// production coordinator (through the app model), the status item and the
/// popover, and defers application termination until every owned preview
/// service confirmed its real exit.
final class AppDelegate: NSObject, NSApplicationDelegate {
    private(set) var model: PreviewAppModel!
    private var alertPresenter: AppAlertPresenter!
    private var servicesProvider: FinderServiceProvider!
    private var statusItem: NSStatusItem!
    private var popover: NSPopover!
    private var failureAlertCancellable: AnyCancellable?

    func applicationDidFinishLaunching(_ notification: Notification) {
        // The single deterministic bundled tool location; there is no
        // Resources, PATH or source-tree fallback.
        let toolURL = Bundle.main.bundleURL
            .appendingPathComponent("Contents/MacOS/go-grip")
        let coordinator = PreviewSessionCoordinator(executableURL: toolURL)
        let alertPresenter = AppAlertPresenter()
        self.alertPresenter = alertPresenter
        model = PreviewAppModel(
            coordinator: coordinator,
            browser: WorkspaceBrowserOpening(),
            confirmation: alertPresenter,
            recentStore: RecentTargetsStore(),
            loginItem: MainAppLoginItem()
        )

        // Failure reports are owned by the root, not by the panel: a cold
        // started Services request can fail before the panel was ever shown,
        // and the user still has to learn about it. One native summary per
        // published report through the shared single-alert slot, independent
        // of popover visibility; the panel keeps listing the same report. The
        // slot presents modelessly, so a Services request arriving while the
        // alert is visible is still serviced, and a second report waits for
        // the visible alert instead of stacking. Deferred by one main-queue
        // turn so the alert is never presented inside the publishing call
        // stack.
        failureAlertCancellable = model.$lastOperationFailures
            .filter { !$0.isEmpty }
            .sink { failures in
                Task { @MainActor in alertPresenter.presentFailureReport(failures) }
            }

        // The only Services provider is registered after the coordinator and
        // the batch entry above exist: the system may deliver the first
        // request immediately, before the panel was ever shown, and an early
        // request must not meet an unready root.
        servicesProvider = FinderServiceProvider(model: model)
        NSApp.servicesProvider = servicesProvider

        setupStatusItem()
        setupPopover()

        // First-use Finder/Services guidance: presented once from its own
        // dedicated flag. It never gates provider registration or the panel,
        // and the panel's Help entry re-presents it without writing this flag
        // or touching any session or recent-target state.
        let guidanceStore = FirstUseGuidanceStore()
        if guidanceStore.shouldPresent() {
            guidanceStore.markPresented()
            alertPresenter.presentServiceGuidance()
        }
    }

    /// Quit path: refuse new opens and browser requests first, then stop every
    /// recorded owned service, including ones still starting, and reply only
    /// after each one confirmed its real exit. Closing the panel or this
    /// callback never stops anything by itself; a forced termination is
    /// covered by the per-child ownership pipe instead.
    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        guard let model else { return .terminateNow }
        if model.beginTermination() {
            return .terminateNow
        }
        Task { @MainActor in
            let mayExit = await model.completeTermination()
            NSApp.reply(toApplicationShouldTerminate: mayExit)
        }
        return .terminateLater
    }

    @MainActor
    private func setupStatusItem() {
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        if let button = statusItem.button {
            button.image = NSImage(systemSymbolName: "doc.text", accessibilityDescription: "GoGrip")
            button.target = self
            button.action = #selector(togglePopover(_:))
        }
    }

    @MainActor
    private func setupPopover() {
        popover = NSPopover()
        // Tall enough for the session rows with their full state reasons, the
        // recent list and the failure report; the content scrolls inside.
        let contentSize = NSSize(width: 420, height: 460)
        popover.contentSize = contentSize
        popover.behavior = .transient
        let hosting = NSHostingController(
            rootView: PopoverView(
                coordinator: model.coordinator,
                model: model,
                onHelp: { [weak self] in self?.alertPresenter.presentServiceGuidance() }
            )
        )
        // Keep the fixed content size above instead of letting AppKit resize
        // the popover from SwiftUI content, which could reposition the panel
        // away from the status item (for example with an empty session list).
        hosting.sizingOptions = []
        popover.contentViewController = hosting
    }

    @MainActor
    @objc private func togglePopover(_ sender: AnyObject?) {
        guard let button = statusItem.button else { return }
        if popover.isShown {
            popover.performClose(sender)
        } else {
            // The login-start status belongs to the system: re-read it every
            // time the panel is shown, so a change made in System Settings is
            // reflected without an app restart.
            model.refreshLoginItemStatus()
            popover.show(relativeTo: button.bounds, of: button, preferredEdge: .minY)
            popover.contentViewController?.view.window?.makeKey()
        }
    }
}
