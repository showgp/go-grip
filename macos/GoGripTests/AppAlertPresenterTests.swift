import AppKit
import XCTest

/// Policy regressions for the app's single native-alert slot: the root failure
/// report and the batch quantity confirmation share one non-modal slot, so a
/// second request waits until the visible alert is answered, and every queued
/// request is presented exactly once. The real AppKit window surface (a
/// modeless alert, dismissal, and Services staying serviceable while it is
/// visible) is covered by the native smoke, not here.
final class AppAlertPresenterTests: XCTestCase {
    @MainActor
    func testFailureReportsArePresentedOneAtATimeInArrivalOrder() {
        let show = ControlledAlertShow()
        let presenter = AppAlertPresenter(show: show.show)

        presenter.presentFailureReport([OperationFailure(displayPath: "/tmp/first", reason: "gone")])
        XCTAssertEqual(show.shown.count, 1, "the first report is presented immediately")
        XCTAssertTrue(show.shown[0].informativeText.contains("/tmp/first"))
        XCTAssertTrue(show.shown[0].informativeText.contains("gone"))

        presenter.presentFailureReport([OperationFailure(displayPath: "/tmp/second", reason: "denied")])
        XCTAssertEqual(show.shown.count, 1, "a second report must not stack while one alert is visible")

        show.answer(.alertFirstButtonReturn)
        XCTAssertEqual(show.shown.count, 2, "the queued report is presented once the first is answered")
        XCTAssertTrue(show.shown[1].informativeText.contains("/tmp/second"))
        XCTAssertTrue(show.shown[1].informativeText.contains("denied"))

        show.answer(.alertFirstButtonReturn)
        XCTAssertEqual(show.shown.count, 2, "each report is presented exactly once")
    }

    @MainActor
    func testQuantityConfirmationResumesWithTheAnsweredResponse() async {
        let show = ControlledAlertShow()
        let presenter = AppAlertPresenter(show: show.show)

        let approved = Task { await presenter.confirmOpening(count: 6) }
        await waitUntil { show.shown.count == 1 }
        XCTAssertTrue(show.shown[0].messageText.contains("6"))
        show.answer(.alertFirstButtonReturn)
        let approvedResult = await approved.value
        XCTAssertTrue(approvedResult)

        let cancelled = Task { await presenter.confirmOpening(count: 7) }
        await waitUntil { show.shown.count == 2 }
        XCTAssertTrue(show.shown[1].messageText.contains("7"))
        show.answer(.alertSecondButtonReturn)
        let cancelledResult = await cancelled.value
        XCTAssertFalse(cancelledResult, "a cancelled confirmation must resume as not approved")
    }

    @MainActor
    func testConfirmationWaitsBehindAVisibleFailureReportAndDoesNotStack() async {
        let show = ControlledAlertShow()
        let presenter = AppAlertPresenter(show: show.show)

        presenter.presentFailureReport([OperationFailure(displayPath: "/tmp/report", reason: "gone")])

        let started = StartedFlag()
        let confirmation = Task { @MainActor in
            started.isSet = true
            return await presenter.confirmOpening(count: 6)
        }
        await waitUntil { started.isSet }
        XCTAssertEqual(show.shown.count, 1, "the confirmation must not appear while the report is visible")

        show.answer(.alertFirstButtonReturn)
        XCTAssertEqual(show.shown.count, 2, "the queued confirmation is presented once the report is answered")
        XCTAssertTrue(show.shown[1].messageText.contains("6"))

        show.answer(.alertFirstButtonReturn)
        let result = await confirmation.value
        XCTAssertTrue(result)
        XCTAssertEqual(show.shown.count, 2, "nothing is presented after the slot is drained")
    }

    @MainActor
    private func waitUntil(timeout: TimeInterval = 2, _ condition: () -> Bool) async {
        let deadline = Date().addingTimeInterval(timeout)
        while !condition() {
            if Date() >= deadline {
                XCTFail("condition was not reached before the deadline")
                return
            }
            await Task.yield()
        }
    }
}

/// Set by a task right before it enters the presenter, so the test knows the
/// request was enqueued (the enqueue happens synchronously before the
/// continuation suspends).
@MainActor
private final class StartedFlag {
    var isSet = false
}

/// Controlled presentation seam: records every alert handed to the slot and
/// answers them on demand.
@MainActor
private final class ControlledAlertShow {
    private(set) var shown: [NSAlert] = []
    private var responses: [(NSApplication.ModalResponse) -> Void] = []

    func show(_ alert: NSAlert, respond: @escaping (NSApplication.ModalResponse) -> Void) {
        shown.append(alert)
        responses.append(respond)
    }

    func answer(_ response: NSApplication.ModalResponse) {
        responses.removeFirst()(response)
    }
}
