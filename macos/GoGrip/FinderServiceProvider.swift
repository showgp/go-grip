import AppKit

/// The single Finder Services entry of the app. A request may arrive right
/// after registration, before the first panel display, so this object reads
/// the file URLs of the request synchronously and hands them to the same
/// production batch entry the manual open panel uses. The batch (background
/// preparation, deduplication, quantity confirmation, one failure report and
/// the default-browser requests) runs asynchronously through the shared
/// coordinator; this entry never waits for it and never touches the request
/// again.
@MainActor
final class FinderServiceProvider: NSObject {
    private let model: PreviewAppModel

    init(model: PreviewAppModel) {
        self.model = model
    }

    /// The Services selector declared in Info.plist as
    /// `NSMessage = openWithGoGrip`. It accepts every file URL of the request;
    /// the `error` pointer is only used before returning, and only for a
    /// request that cannot be read at all.
    @objc func openWithGoGrip(
        _ pboard: NSPasteboard,
        userData: String?,
        error: AutoreleasingUnsafeMutablePointer<NSString?>
    ) {
        let objects = pboard.readObjects(
            forClasses: [NSURL.self],
            options: [.urlReadingFileURLsOnly: true]
        )
        let urls = (objects as? [URL]) ?? []
        guard !urls.isEmpty else {
            error.pointee = NSLocalizedString(
                "The service request did not contain a readable file",
                comment: "Immediate Services error for an unreadable request"
            ) as NSString
            return
        }
        // Capture the values, not the pasteboard: the request ends when the
        // selector returns, and target, launch and browser errors belong to
        // the batch's own report, not to this error pointer.
        Task { @MainActor [model, urls] in
            await model.openBatch(urls)
        }
    }
}
