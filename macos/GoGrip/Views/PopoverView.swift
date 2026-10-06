import AppKit
import SwiftUI
import UniformTypeIdentifiers

/// The menu bar panel: running sessions read from the single coordinator fact
/// source and stay above the persisted recent targets, with explicit
/// reopen-browser, copy-URL, stop-one, stop-all, recent-reopen, clear-history,
/// manual batch open and quit actions. Every state is shown with its word and
/// reason; no status dot hides an action.
struct PopoverView: View {
    @ObservedObject var coordinator: PreviewSessionCoordinator
    @ObservedObject var model: PreviewAppModel
    /// Presented by the panel's Help entry: reopens the Finder/Services
    /// guidance without touching sessions or the persisted recent targets.
    let onHelp: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
            Divider()
            ScrollView {
                VStack(alignment: .leading, spacing: 12) {
                    sessionsSection
                    Divider()
                    recentSection
                    Divider()
                    loginSection
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(12)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            if !model.lastOperationFailures.isEmpty {
                Divider()
                failureView(model.lastOperationFailures)
            }
            Divider()
            footer
        }
        .frame(width: 420)
    }

    private var header: some View {
        Text("GoGrip")
            .font(.headline)
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
    }

    @ViewBuilder
    private var sessionsSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Sessions")
                .font(.subheadline)
                .bold()
            if coordinator.sessions.isEmpty {
                Text("No preview sessions")
                    .foregroundColor(.secondary)
            } else {
                ForEach(coordinator.sessions) { session in
                    SessionRow(session: session, model: model)
                    if session.id != coordinator.sessions.last?.id {
                        Divider()
                    }
                }
            }
        }
    }

    @ViewBuilder
    private var recentSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("Recent")
                    .font(.subheadline)
                    .bold()
                Spacer()
                Button("Clear") {
                    model.clearRecentTargets()
                }
                .controlSize(.small)
                .disabled(model.recentTargets.isEmpty)
            }
            if model.recentTargets.isEmpty {
                Text("No recent targets")
                    .foregroundColor(.secondary)
            } else {
                ForEach(model.recentTargets) { target in
                    RecentRow(target: target, model: model)
                }
            }
        }
    }

    /// The opt-in login-start option: the real system status from
    /// `SMAppService.mainApp` and an explicit action matching it. The item is
    /// never turned on or off by anything but these buttons, and a
    /// registration waiting for system approval is shown with the settings
    /// check that finishes it instead of a fake enabled switch.
    @ViewBuilder
    private var loginSection: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                Text("Launch at login")
                    .font(.subheadline)
                    .bold()
                Spacer()
                Text(loginStatusLabel)
                    .font(.caption)
                    .foregroundColor(.secondary)
                loginActions
            }
            if model.loginItemStatus == .requiresApproval {
                Text("GoGrip can launch at login only after you allow it in System Settings → General → Login Items.")
                    .font(.caption)
                    .foregroundColor(.orange)
            }
        }
    }

    /// The real system status as one of the panel's words. `notFound` means
    /// the framework found no registration record for this app — the state a
    /// never-registered app reports on current macOS — which the panel shows
    /// like `notRegistered`, with the explicit action; `register()` and the
    /// re-read status remain the only judges of the outcome.
    private var loginStatusLabel: String {
        switch model.loginItemStatus {
        case .notRegistered, .notFound:
            return NSLocalizedString("Not set", comment: "Login item status")
        case .enabled:
            return NSLocalizedString("Enabled", comment: "Login item status")
        case .requiresApproval:
            return NSLocalizedString("Waiting for approval", comment: "Login item status")
        @unknown default:
            return NSLocalizedString("Unavailable", comment: "Login item status")
        }
    }

    @ViewBuilder
    private var loginActions: some View {
        switch model.loginItemStatus {
        case .notRegistered, .notFound:
            Button("Turn On") {
                model.setLoginItemEnabled(true)
            }
            .controlSize(.small)
        case .enabled:
            Button("Turn Off") {
                model.setLoginItemEnabled(false)
            }
            .controlSize(.small)
        case .requiresApproval:
            HStack(spacing: 8) {
                Button("Open Login Items…") {
                    model.openLoginItemsSettings()
                }
                Button("Turn Off") {
                    model.setLoginItemEnabled(false)
                }
            }
            .controlSize(.small)
        @unknown default:
            EmptyView()
        }
    }

    private func failureView(_ failures: [OperationFailure]) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 2) {
                ForEach(failures, id: \.self) { failure in
                    Text(failure.displayPath)
                        .font(.caption)
                        .lineLimit(1)
                        .truncationMode(.middle)
                        .help(failure.displayPath)
                    Text(failure.reason)
                        .font(.caption)
                        .foregroundColor(.red)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .frame(maxHeight: 80)
        .padding(.horizontal, 12)
        .padding(.vertical, 6)
    }

    private var footer: some View {
        HStack(spacing: 8) {
            Button("Open…", action: showOpenPanel)
            Button("Help", action: onHelp)
            Button("Stop All") {
                Task { await model.stopAll() }
            }
            .disabled(coordinator.sessions.allSatisfy { $0.phase == .terminated })
            Spacer()
            Button("Quit") {
                NSApp.terminate(nil)
            }
        }
        .controlSize(.small)
        .padding(12)
    }

    private func showOpenPanel() {
        let panel = NSOpenPanel()
        panel.allowsMultipleSelection = true
        panel.canChooseDirectories = true
        panel.canChooseFiles = true
        var types: [UTType] = [.folder]
        if let markdown = UTType(filenameExtension: "md") {
            types.append(markdown)
        }
        panel.allowedContentTypes = types
        panel.begin { response in
            guard response == .OK else { return }
            let selected = panel.urls
            Task { await model.openBatch(selected) }
        }
    }
}

/// One running session with its target path, actual phase, target
/// accessibility, hot-reload state, verified URL and available failures, plus
/// the explicit per-session actions. Every observable state stays separate:
/// an unavailable target or a degraded watcher does not hide the running
/// service or its management entry.
private struct SessionRow: View {
    let session: PreviewSession
    let model: PreviewAppModel

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(session.displayPath)
                .font(.callout)
                .lineLimit(1)
                .truncationMode(.middle)
                .help(session.displayPath)
            HStack(spacing: 8) {
                Text(phaseLabel)
                    .font(.caption)
                    .foregroundColor(.secondary)
                if let url = session.url {
                    Text(url.absoluteString)
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }
            }
            if let target = session.targetStatus, target.state == "unavailable" {
                Text("Target unavailable: \(target.reason ?? reasonFallback)")
                    .font(.caption)
                    .foregroundColor(.red)
            }
            if let reload = session.reloadStatus {
                reloadLine(reload.state, reason: reload.reason)
            }
            if let browserFailure = session.browserFailure {
                Text("Browser: \(browserFailure)")
                    .font(.caption)
                    .foregroundColor(.red)
            }
            if let failure = session.lastFailure {
                Text(failure)
                    .font(.caption)
                    .foregroundColor(.red)
            }
            HStack(spacing: 8) {
                Button("Open in Browser") {
                    model.openBrowser(session.id)
                }
                .disabled(session.phase != .running || session.url == nil)
                Button("Copy URL", action: copyURL)
                    .disabled(session.url == nil)
                Button("Stop") {
                    Task { await model.stop(session.id) }
                }
                .disabled(session.phase == .terminated)
            }
            .controlSize(.small)
        }
    }

    /// The real hot-reload state. `active` is the quiet normal case; pending,
    /// degraded and off are shown with their available reason so incomplete
    /// coverage is never presented as complete automatic refresh.
    @ViewBuilder
    private func reloadLine(_ state: String, reason: String?) -> some View {
        switch state {
        case "pending":
            Text("Hot reload: preparing")
                .font(.caption)
                .foregroundColor(.secondary)
        case "degraded":
            Text("Hot reload degraded: \(reason ?? reasonFallback)")
                .font(.caption)
                .foregroundColor(.orange)
        case "disabled":
            Text("Hot reload: off")
                .font(.caption)
                .foregroundColor(.secondary)
        default:
            EmptyView()
        }
    }

    /// Used only when a state was reported without its available reason; it
    /// must not read like a machine diagnostic.
    private var reasonFallback: String {
        NSLocalizedString("the available reason was not reported", comment: "Fallback when a state has no reason")
    }

    private var phaseLabel: String {
        switch session.phase {
        case .starting: return NSLocalizedString("Starting", comment: "Session phase")
        case .running: return NSLocalizedString("Running", comment: "Session phase")
        case .stopping: return NSLocalizedString("Stopping", comment: "Session phase")
        case .terminated: return NSLocalizedString("Stopped", comment: "Session phase")
        }
    }

    private func copyURL() {
        guard let url = session.url else { return }
        let pasteboard = NSPasteboard.general
        pasteboard.clearContents()
        pasteboard.setString(url.absoluteString, forType: .string)
    }
}

/// One persisted recent target with its user-selected path and an explicit
/// reopen action that re-enters the same production batch entry as Finder and
/// the open panel.
private struct RecentRow: View {
    let target: RecentTarget
    let model: PreviewAppModel

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: target.mode == .directory ? "folder" : "doc.text")
                .foregroundColor(.secondary)
            Text(target.displayPath)
                .font(.callout)
                .lineLimit(1)
                .truncationMode(.middle)
                .help(target.displayPath)
            Spacer()
            Button("Open") {
                Task { await model.reopenRecent(target) }
            }
            .controlSize(.small)
        }
    }
}
