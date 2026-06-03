import SwiftUI

struct HistoryRowView: View {
    let entry: HistoryEntry
    let isRunning: Bool
    let onTap: () -> Void
    let onStop: () -> Void
    let onDelete: () -> Void

    @State private var isHovering = false

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 6)
                .fill(isHovering ? Color.accentColor.opacity(0.1) : Color.clear)

            HStack(spacing: 8) {
                Image(systemName: entry.isDirectory ? "folder" : "doc")
                    .foregroundColor(.secondary)
                    .frame(width: 16)

                Text(entry.displayName)
                    .lineLimit(1)
                    .truncationMode(.middle)
                    .font(.body)

                Spacer()

                if isHovering {
                    Button(action: onDelete) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(.secondary)
                    }
                    .buttonStyle(.plain)
                    .transition(.opacity)
                }

                Circle()
                    .fill(isRunning ? Color.green : Color.gray.opacity(0.3))
                    .frame(width: 10, height: 10)
                    .help(isRunning ? "Click to stop" : "Not running")
                    .onTapGesture {
                        if isRunning { onStop() }
                    }
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 6)
        }
        .contentShape(RoundedRectangle(cornerRadius: 6))
        .onTapGesture { onTap() }
        .modifier(ReliableHoverModifier(
            onHover: { isHovering = $0 }
        ))
    }
}

// MARK: - Reliable hover tracking via NSTrackingArea
// SwiftUI's .onHover uses hit-testing that can be offset in NSPopover + NSHostingController
// contexts. NSTrackingArea operates in AppKit's native coordinate system, bypassing this bug.
private struct ReliableHoverModifier: ViewModifier {
    let onHover: (Bool) -> Void

    func body(content: Content) -> some View {
        content
            .background(
                GeometryReader { proxy in
                    Representable(
                        frame: proxy.frame(in: .global),
                        onHover: onHover
                    )
                }
            )
    }

    private struct Representable: NSViewRepresentable {
        let frame: NSRect
        let onHover: (Bool) -> Void

        func makeCoordinator() -> Coordinator {
            Coordinator(onHover: onHover)
        }

        func makeNSView(context: Context) -> NSView {
            let view = TrackingAreaView()
            view.coordinator = context.coordinator
            updateTrackingArea(for: view)
            return view
        }

        func updateNSView(_ nsView: NSView, context: Context) {
            updateTrackingArea(for: nsView)
        }

        private func updateTrackingArea(for view: NSView) {
            view.trackingAreas.forEach { view.removeTrackingArea($0) }
            let options: NSTrackingArea.Options = [
                .mouseEnteredAndExited,
                .inVisibleRect,
                .activeAlways,
            ]
            let area = NSTrackingArea(
                rect: view.bounds,
                options: options,
                owner: view,
                userInfo: nil
            )
            view.addTrackingArea(area)
        }

        class Coordinator {
            let onHover: (Bool) -> Void
            init(onHover: @escaping (Bool) -> Void) {
                self.onHover = onHover
            }
        }

        class TrackingAreaView: NSView {
            var coordinator: Coordinator?

            override var acceptsFirstResponder: Bool { false }

            override func mouseEntered(with event: NSEvent) {
                coordinator?.onHover(true)
            }

            override func mouseExited(with event: NSEvent) {
                coordinator?.onHover(false)
            }

            override func updateTrackingAreas() {
                super.updateTrackingAreas()
                trackingAreas.forEach { removeTrackingArea($0) }
                let options: NSTrackingArea.Options = [
                    .mouseEnteredAndExited,
                    .inVisibleRect,
                    .activeAlways,
                ]
                let area = NSTrackingArea(
                    rect: bounds,
                    options: options,
                    owner: self,
                    userInfo: nil
                )
                addTrackingArea(area)
            }
        }
    }
}
