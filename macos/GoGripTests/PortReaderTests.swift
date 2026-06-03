import XCTest
@testable import GoGrip

final class PortReaderTests: XCTestCase {
    func testReadPortFromStdout_NormalJson() async {
        let pipe = Pipe()
        let data = #"{"port": 6420}"#.data(using: .utf8)!
        pipe.fileHandleForWriting.write(data)
        pipe.fileHandleForWriting.closeFile()

        let port = await ProcessManager.readPortFromPipe(pipe, timeout: 2)
        XCTAssertEqual(port, 6420)
    }

    func testReadPortFromStdout_MixedContent() async {
        let pipe = Pipe()
        let data = "Starting server...\n{\"port\": 6420}\nServer ready.\n".data(using: .utf8)!
        pipe.fileHandleForWriting.write(data)
        pipe.fileHandleForWriting.closeFile()

        let port = await ProcessManager.readPortFromPipe(pipe, timeout: 2)
        XCTAssertEqual(port, 6420)
    }

    func testReadPortFromStdout_EOFWithoutJson() async {
        let pipe = Pipe()
        pipe.fileHandleForWriting.closeFile()

        let port = await ProcessManager.readPortFromPipe(pipe, timeout: 2)
        XCTAssertEqual(port, 6419)
    }

    func testReadPortFromStdout_Timeout() async {
        let pipe = Pipe()

        let port = await ProcessManager.readPortFromPipe(pipe, timeout: 1)
        XCTAssertEqual(port, 6419)
    }
}
