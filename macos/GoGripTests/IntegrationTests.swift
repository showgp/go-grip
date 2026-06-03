import XCTest
@testable import GoGrip

final class IntegrationTests: XCTestCase {
    var binaryPath: String {
        if let bundleURL = Bundle.main.url(forResource: "go-grip", withExtension: nil) {
            return bundleURL.path
        }
        let url = URL(fileURLWithPath: #file)
            .deletingLastPathComponent().deletingLastPathComponent()
            .deletingLastPathComponent()
        return url.appendingPathComponent("go-grip").path
    }

    var binaryExists: Bool {
        FileManager.default.fileExists(atPath: binaryPath)
    }

    func testGoGripBinaryExists() {
        XCTAssertTrue(binaryExists, "go-grip binary not found at \(binaryPath)")
    }

    func testGoGripProcessStarts() async throws {
        try XCTSkipUnless(binaryExists, "go-grip binary not found")
        let process = Process()
        process.executableURL = URL(fileURLWithPath: binaryPath)
        process.arguments = ["--json", "--browser=false", "--no-reload"]
        process.standardOutput = Pipe()
        process.standardError = FileHandle.nullDevice

        addTeardownBlock {
            if process.isRunning {
                process.terminate()
                process.waitUntilExit()
            }
        }

        try process.run()
        XCTAssertTrue(process.isRunning)

        process.terminate()
        process.waitUntilExit()
    }

    func testGoGripStdoutJson() async throws {
        try XCTSkipUnless(binaryExists, "go-grip binary not found")
        let process = Process()
        process.executableURL = URL(fileURLWithPath: binaryPath)
        process.arguments = ["--json", "--browser=false", "--no-reload"]
        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = FileHandle.nullDevice

        addTeardownBlock {
            if process.isRunning {
                process.terminate()
                process.waitUntilExit()
            }
        }

        try process.run()

        let port = await ProcessManager.readPortFromPipe(pipe, timeout: 10)
        XCTAssertGreaterThan(port, 0)

        process.terminate()
        process.waitUntilExit()
    }

    func testGoGripHttpRequest() async throws {
        try XCTSkipUnless(binaryExists, "go-grip binary not found")
        let process = Process()
        process.executableURL = URL(fileURLWithPath: binaryPath)
        process.arguments = ["--json", "--browser=false", "--no-reload"]
        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = FileHandle.nullDevice

        addTeardownBlock {
            if process.isRunning {
                process.terminate()
                process.waitUntilExit()
            }
        }

        try process.run()

        let port = await ProcessManager.readPortFromPipe(pipe, timeout: 10)
        guard port > 0, let url = URL(string: "http://127.0.0.1:\(port)/") else {
            process.terminate()
            process.waitUntilExit()
            XCTFail("Failed to get valid port")
            return
        }

        let (_, response) = try await URLSession.shared.data(from: url)
        let httpResponse = try XCTUnwrap(response as? HTTPURLResponse)
        XCTAssertEqual(httpResponse.statusCode, 200)

        process.terminate()
        process.waitUntilExit()
    }
}
