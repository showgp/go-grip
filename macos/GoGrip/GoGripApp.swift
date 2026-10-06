import AppKit

/// AppKit entry point of the menu bar host: the application lifecycle is owned
/// directly, without a SwiftUI `App`, scene or settings window. SwiftUI only
/// renders the popover content.
@main
enum GoGripMain {
    /// `NSApplication.delegate` is weak; the app root stays alive for the
    /// whole process.
    private static let delegate = AppDelegate()

    static func main() {
        let application = NSApplication.shared
        application.delegate = delegate
        application.setActivationPolicy(.accessory)
        application.run()
    }
}
