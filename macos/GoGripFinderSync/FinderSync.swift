import Cocoa
import FinderSync

class FinderSync: FIFinderSync {

    override init() {
        super.init()
        let homeURL = FileManager.default.homeDirectoryForCurrentUser
        FIFinderSyncController.default().directoryURLs = [homeURL]
    }

    // MARK: - Context Menu

    override func menu(for menuKind: FIMenuKind) -> NSMenu? {
        guard menuKind == .contextualMenuForItems else { return nil }

        let menu = NSMenu(title: "")
        let menuItem = NSMenuItem(
            title: "Open with GoGrip",
            action: #selector(openWithGoGrip(_:)),
            keyEquivalent: ""
        )
        menuItem.target = self
        menu.addItem(menuItem)
        return menu
    }

    @objc func openWithGoGrip(_ sender: NSMenuItem) {
        guard let items = FIFinderSyncController.default().selectedItemURLs(),
              let selectedURL = items.first else { return }

        let path = selectedURL.path
        guard let encodedPath = path.addingPercentEncoding(
            withAllowedCharacters: .urlQueryAllowed
        ) else { return }
        guard let openURL = URL(string: "gogrip://open?path=\(encodedPath)") else { return }

        NSWorkspace.shared.open(openURL)
    }
}
