import AppKit
import SwiftUI
import Combine

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    var popover: NSPopover!
    var processManager: ProcessManager!
    private var cancellables = Set<AnyCancellable>()

    func applicationDidFinishLaunching(_ notification: Notification) {
        processManager = ProcessManager()

        setupStatusItem()
        setupPopover()
        setupBadgeObserver()
    }

    private func setupStatusItem() {
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        if let button = statusItem.button {
            button.image = NSImage(systemSymbolName: "doc.text", accessibilityDescription: "GoGrip")
            button.action = #selector(togglePopover(_:))
        }
    }

    private func setupPopover() {
        popover = NSPopover()
        popover.contentSize = NSSize(width: 400, height: 280)
        popover.behavior = .applicationDefined
        popover.animates = true

        let popoverView = PopoverView(
            processManager: processManager,
            onDragStateChange: { [weak self] dragging in
                self?.popover.behavior = dragging ? .applicationDefined : .applicationDefined
            },
            onSizeChange: { [weak self] size in
                guard let self, size.height >= 200 else { return }
                let capped = NSSize(
                    width: max(360, min(size.width, 480)),
                    height: max(200, min(size.height, 800))
                )
                self.popover.contentSize = capped
            }
        )
        popover.contentViewController = NSHostingController(rootView: popoverView)
    }

    private func setupBadgeObserver() {
        processManager.objectWillChange
            .receive(on: DispatchQueue.main)
            .sink { [weak self] _ in
                self?.updateBadge(count: self?.processManager.count ?? 0)
            }
            .store(in: &cancellables)
    }

    func applicationWillTerminate(_ notification: Notification) {
        processManager.stopAll()
    }

    @objc func togglePopover(_ sender: AnyObject?) {
        if let button = statusItem.button {
            if popover.isShown {
                popover.performClose(sender)
            } else {
                popover.show(relativeTo: button.bounds, of: button, preferredEdge: .minY)
            }
        }
    }

    func updateBadge(count: Int) {
        guard let button = statusItem.button,
              let image = NSImage(systemSymbolName: "doc.text", accessibilityDescription: "GoGrip") else {
            return
        }

        if count > 0 {
            let badgeImage = NSImage(size: NSSize(width: 18, height: 18), flipped: false) { rect in
                let circleRect = NSRect(x: 8, y: 8, width: 18, height: 18)
                NSColor.red.setFill()
                NSBezierPath(ovalIn: circleRect).fill()

                let text = "\(count)" as NSString
                let attrs: [NSAttributedString.Key: Any] = [
                    .font: NSFont.systemFont(ofSize: 11, weight: .bold),
                    .foregroundColor: NSColor.white,
                ]
                let textSize = text.size(withAttributes: attrs)
                let textRect = NSRect(
                    x: circleRect.midX - textSize.width / 2,
                    y: circleRect.midY - textSize.height / 2,
                    width: textSize.width,
                    height: textSize.height
                )
                text.draw(in: textRect, withAttributes: attrs)
                return true
            }
            image.lockFocus()
            badgeImage.draw(at: .zero, from: .zero, operation: .sourceOver, fraction: 1.0)
            image.unlockFocus()
        }

        if count == 0 {
            image.isTemplate = true
        }
        button.image = image
    }
}
