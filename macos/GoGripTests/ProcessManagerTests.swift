import XCTest
@testable import GoGrip

final class ProcessManagerTests: XCTestCase {
    var manager: ProcessManager!
    var mockBinaryURLs: [URL] = []

    override func setUp() {
        super.setUp()
        manager = ProcessManager()
        mockBinaryURLs = []
    }

    override func tearDown() {
        for url in mockBinaryURLs {
            try? FileManager.default.removeItem(at: url)
        }
        mockBinaryURLs = []
        super.tearDown()
    }

    func createMockBinary(port: Int) -> URL {
        let tempDir = FileManager.default.temporaryDirectory
        let binPath = tempDir.appendingPathComponent("go-grip-mock-\(UUID().uuidString)")
        let script = """
        #!/bin/bash
        echo '{"port":\(port),"host":"localhost","url":"http://localhost:\(port)/test.md"}'
        sleep 30
        """
        try? script.write(to: binPath, atomically: true, encoding: .utf8)
        try? FileManager.default.setAttributes([.posixPermissions: 0o755], ofItemAtPath: binPath.path)
        mockBinaryURLs.append(binPath)
        return binPath
    }

    func testInitialState() {
        XCTAssertTrue(manager.instances.isEmpty)
        XCTAssertEqual(manager.count, 0)
    }

    func testCount() {
        let process = Process()
        manager.instances["test"] = RunningInstance(path: "/tmp/go-grip", process: process, port: 6420)
        XCTAssertEqual(manager.count, 1)
    }

    func testIsRunning() {
        XCTAssertFalse(manager.isRunning(path: "/nonexistent"))

        let process = Process()
        manager.instances["test"] = RunningInstance(path: "/tmp/go-grip", process: process, port: 6420)
        XCTAssertFalse(manager.isRunning(path: "/tmp/go-grip"))
    }

    func testPort() {
        XCTAssertNil(manager.port(for: "/nonexistent"))

        let process = Process()
        manager.instances["test"] = RunningInstance(path: "/tmp/go-grip", process: process, port: 6420)
        XCTAssertEqual(manager.port(for: "/tmp/go-grip"), 6420)
    }

    func testStopAll() {
        let process1 = Process()
        let process2 = Process()
        manager.instances["a"] = RunningInstance(path: "/tmp/a", process: process1, port: 6420)
        manager.instances["b"] = RunningInstance(path: "/tmp/b", process: process2, port: 6421)

        manager.stopAll()

        let expectation = XCTestExpectation(description: "stopAll cleanup")
        DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
            XCTAssertTrue(self.manager.instances.isEmpty)
            XCTAssertEqual(self.manager.count, 0)
            expectation.fulfill()
        }
        wait(for: [expectation], timeout: 5)
    }

    func testProcessManagerIsObservableObject() {
        _ = manager.objectWillChange
    }

    // MARK: - C3: ProcessManager.start() tests

    func testStartWithMockBinary() async throws {
        let mockURL = createMockBinary(port: 6420)
        manager._testBinaryURL = mockURL
        await manager.start(path: "/tmp/test.md")

        XCTAssertEqual(manager.count, 1)
        XCTAssertEqual(manager.port(for: "/tmp/test.md"), 6420)
        XCTAssertTrue(manager.isRunning(path: "/tmp/test.md"))

        manager.stop(path: "/tmp/test.md")
        let expectation = XCTestExpectation(description: "process termination cleanup")
        DispatchQueue.main.asyncAfter(deadline: .now() + 1) {
            XCTAssertNil(self.manager.instances["/tmp/test.md"])
            expectation.fulfill()
        }
        wait(for: [expectation], timeout: 3)
    }

    func testStartMissingBinary() async throws {
        await manager.start(path: "/tmp/test.md")
        XCTAssertTrue(manager.instances.isEmpty)
    }

    func testStartMultipleInstances() async throws {
        let mockURL1 = createMockBinary(port: 6420)
        let mockURL2 = createMockBinary(port: 6421)

        manager._testBinaryURL = mockURL1
        await manager.start(path: "/tmp/a.md")

        manager._testBinaryURL = mockURL2
        await manager.start(path: "/tmp/b.md")

        XCTAssertEqual(manager.count, 2)
        XCTAssertEqual(manager.port(for: "/tmp/a.md"), 6420)
        XCTAssertEqual(manager.port(for: "/tmp/b.md"), 6421)

        manager.stopAll()
        let expectation = XCTestExpectation(description: "stopAll cleanup")
        DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
            XCTAssertEqual(self.manager.count, 0)
            expectation.fulfill()
        }
        wait(for: [expectation], timeout: 5)
    }
}
