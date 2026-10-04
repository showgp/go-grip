import Foundation
import XCTest

/// Framing, fragmentation and v1 validation of the managed NDJSON stream.
/// Expected frames come from the approved wire contract, not from re-encoding.
final class ManagedProtocolTests: XCTestCase {
    private let generation = "gen-003"

    private func decoder() -> ManagedNDJSONDecoder {
        ManagedNDJSONDecoder(generation: generation)
    }

    private func lines(_ raw: String...) -> Data {
        Data((raw.joined(separator: "\n") + "\n").utf8)
    }

    private func assertInvalidFrame(
        file: StaticString = #filePath,
        line: UInt = #line,
        _ body: () throws -> Void
    ) {
        XCTAssertThrowsError(try body(), file: file, line: line) { error in
            guard case ManagedProtocolError.invalidFrame = error else {
                XCTFail("expected invalidFrame, got \(error)", file: file, line: line)
                return
            }
        }
    }

    func testAssemblesReadyFrameAcrossChunks() throws {
        let line = #"{"version":1,"event":"ready","generation":"gen-003","url":"http://127.0.0.1:54321/","reload":{"state":"pending"}}"# + "\n"
        let bytes = Data(line.utf8)
        let first = bytes.prefix(10)
        let second = bytes.dropFirst(10).prefix(60)
        let third = bytes.dropFirst(70)

        let decoder = decoder()
        XCTAssertEqual(try decoder.feed(Data(first)), [])
        XCTAssertEqual(try decoder.feed(Data(second)), [])
        XCTAssertEqual(
            try decoder.feed(Data(third)),
            [.ready(url: "http://127.0.0.1:54321/", reload: ManagedReloadSnapshot(state: "pending", reason: nil))]
        )
        XCTAssertNoThrow(try decoder.finish())
    }

    func testKeepsUTF8CharacterSplitAcrossChunks() throws {
        let line = #"{"version":1,"event":"reload-status","generation":"gen-003","reload":{"state":"degraded","reason":"预算不足：目录过多"}}"# + "\n"
        let bytes = Data(line.utf8)
        // Split inside the multi-byte encoding of 「预」.
        let splitIndex = line.utf8.count - Data("预算不足：目录过多\"}}".utf8).count + 1
        let decoder = decoder()
        XCTAssertEqual(try decoder.feed(Data(bytes.prefix(splitIndex))), [])
        XCTAssertEqual(
            try decoder.feed(Data(bytes.dropFirst(splitIndex))),
            [.reloadStatus(ManagedReloadSnapshot(state: "degraded", reason: "预算不足：目录过多"))]
        )
    }

    func testEmitsMultipleFramesFromSingleChunk() throws {
        let chunk = lines(
            #"{"version":1,"event":"reload-status","generation":"gen-003","reload":{"state":"active"}}"#,
            #"{"version":1,"event":"target-status","generation":"gen-003","target":{"state":"unavailable","reason":"moved"}}"#,
            #"{"version":1,"event":"fatal","generation":"gen-003","code":"serve-failed","message":"boom"}"#
        )
        XCTAssertEqual(
            try decoder().feed(chunk),
            [
                .reloadStatus(ManagedReloadSnapshot(state: "active", reason: nil)),
                .targetStatus(ManagedTargetSnapshot(state: "unavailable", reason: "moved")),
                .fatal(code: "serve-failed", message: "boom"),
            ]
        )
    }

    func testRejectsMalformedJSON() {
        assertInvalidFrame {
            _ = try self.decoder().feed(Data("not json\n".utf8))
        }
    }

    func testRejectsUnsupportedVersion() {
        XCTAssertThrowsError(
            try decoder().feed(lines(#"{"version":2,"event":"ready","generation":"gen-003","url":"http://127.0.0.1:1/","reload":{"state":"pending"}}"#))
        ) { error in
            XCTAssertEqual(error as? ManagedProtocolError, .unsupportedVersion(2))
        }
    }

    func testRejectsGenerationMismatch() {
        XCTAssertThrowsError(
            try decoder().feed(lines(#"{"version":1,"event":"ready","generation":"other","url":"http://127.0.0.1:1/","reload":{"state":"pending"}}"#))
        ) { error in
            XCTAssertEqual(error as? ManagedProtocolError, .generationMismatch(expected: self.generation, received: "other"))
        }
    }

    func testRejectsReadyWithoutURL() {
        assertInvalidFrame {
            _ = try self.decoder().feed(self.lines(#"{"version":1,"event":"ready","generation":"gen-003","reload":{"state":"pending"}}"#))
        }
    }

    func testRejectsReadyWithoutReloadSnapshot() {
        assertInvalidFrame {
            _ = try self.decoder().feed(self.lines(#"{"version":1,"event":"ready","generation":"gen-003","url":"http://127.0.0.1:1/"}"#))
        }
    }

    func testRejectsUnknownEvent() {
        assertInvalidFrame {
            _ = try self.decoder().feed(self.lines(#"{"version":1,"event":"heartbeat","generation":"gen-003"}"#))
        }
    }

    func testRejectsReloadStatusWithoutNestedReload() {
        assertInvalidFrame {
            _ = try self.decoder().feed(self.lines(#"{"version":1,"event":"reload-status","generation":"gen-003","state":"active"}"#))
        }
    }

    func testRejectsTargetStatusWithUnknownState() {
        assertInvalidFrame {
            _ = try self.decoder().feed(self.lines(#"{"version":1,"event":"target-status","generation":"gen-003","target":{"state":"maybe"}}"#))
        }
    }

    func testRejectsFatalWithoutCode() {
        assertInvalidFrame {
            _ = try self.decoder().feed(self.lines(#"{"version":1,"event":"fatal","generation":"gen-003","message":"boom"}"#))
        }
    }

    func testRejectsUnterminatedResidualOverLimit() {
        let oversized = Data(repeating: UInt8(ascii: "x"), count: ManagedNDJSONDecoder.maxResidualBytes + 1)
        XCTAssertThrowsError(try decoder().feed(oversized)) { error in
            XCTAssertEqual(error as? ManagedProtocolError, .residualLimitExceeded(ManagedNDJSONDecoder.maxResidualBytes + 1))
        }
    }

    func testFinishRejectsUnterminatedResidual() {
        assertInvalidFrame {
            let decoder = self.decoder()
            _ = try decoder.feed(Data(#"{"version":1,"event":"ready""#.utf8))
            try decoder.finish()
        }
    }

    func testRejectsOversizedCompleteFrameAcrossChunks() throws {
        let oversized = String(repeating: "x", count: ManagedNDJSONDecoder.maxResidualBytes + 1024) + "\n"
        let bytes = Data(oversized.utf8)
        let split = ManagedNDJSONDecoder.maxResidualBytes - 1024
        let decoder = decoder()

        XCTAssertEqual(try decoder.feed(Data(bytes.prefix(split))), [])
        assertInvalidFrame {
            _ = try decoder.feed(Data(bytes.dropFirst(split)))
        }
    }

    func testRejectsOversizedSingleChunkCompleteLine() {
        let oversized = Data(
            (String(repeating: "x", count: ManagedNDJSONDecoder.maxResidualBytes + 1) + "\n").utf8
        )
        assertInvalidFrame {
            _ = try self.decoder().feed(oversized)
        }
    }

    func testRejectsFatalWithoutMessage() {
        assertInvalidFrame {
            _ = try self.decoder().feed(
                self.lines(#"{"version":1,"event":"fatal","generation":"gen-003","code":"serve-failed"}"#)
            )
        }
    }
}
