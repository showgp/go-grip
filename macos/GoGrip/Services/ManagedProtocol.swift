import Foundation

/// One decoded v1 managed protocol frame.
enum ManagedFrame: Equatable {
    case ready(url: String, reload: ManagedReloadSnapshot)
    case reloadStatus(ManagedReloadSnapshot)
    case targetStatus(ManagedTargetSnapshot)
    case fatal(code: String, message: String)
}

/// Reload coverage carried by ready and reload-status frames.
struct ManagedReloadSnapshot: Equatable {
    /// pending while the initial scan runs, active when the watch set is
    /// complete, degraded when setup, the budget or a runtime error left it
    /// incomplete, disabled for the approved --no-reload snapshot.
    static let states: Set<String> = ["pending", "active", "degraded", "disabled"]

    let state: String
    let reason: String?
}

/// Target accessibility carried by target-status frames.
struct ManagedTargetSnapshot: Equatable {
    /// available / unavailable from real target access.
    static let states: Set<String> = ["available", "unavailable"]

    let state: String
    let reason: String?
}

/// A v1 protocol violation. The integration fails on any of these instead of
/// degrading the parse.
enum ManagedProtocolError: Error, Equatable {
    case invalidFrame(String)
    case unsupportedVersion(Int)
    case generationMismatch(expected: String, received: String)
    case residualLimitExceeded(Int)
}

/// Incremental decoder for the managed v1 NDJSON stream. Bytes are buffered
/// until a complete newline-terminated frame arrives, so frames split across
/// reads (including inside a multi-byte UTF-8 sequence) decode intact. The
/// residual is bounded so truncated or hostile feedback cannot accumulate
/// without limit.
final class ManagedNDJSONDecoder {
    /// Bound on the buffered bytes without a complete frame and on a single
    /// complete frame, so neither unterminated nor oversized feedback can
    /// accumulate without limit.
    static let maxResidualBytes = 64 * 1024

    private let generation: String
    private var residual = Data()

    init(generation: String) {
        self.generation = generation
    }

    /// Consumes one read chunk and returns the frames it completed. Bytes are
    /// consumed incrementally and every limit is checked before data is
    /// appended, so neither a single oversized read nor an unterminated
    /// stream can exceed the 64 KiB bound.
    func feed(_ chunk: Data) throws -> [ManagedFrame] {
        var frames: [ManagedFrame] = []
        var cursor = chunk.startIndex
        while cursor < chunk.endIndex {
            guard let newline = chunk[cursor...].firstIndex(of: UInt8(ascii: "\n")) else {
                let remaining = chunk.distance(from: cursor, to: chunk.endIndex)
                guard residual.count + remaining <= Self.maxResidualBytes else {
                    throw ManagedProtocolError.residualLimitExceeded(residual.count + remaining)
                }
                residual.append(contentsOf: chunk[cursor...])
                break
            }
            let lineLength = residual.count + chunk.distance(from: cursor, to: newline)
            guard lineLength <= Self.maxResidualBytes else {
                throw ManagedProtocolError.invalidFrame(
                    "frame of \(lineLength) bytes exceeds the 64 KiB cap"
                )
            }
            residual.append(contentsOf: chunk[cursor..<newline])
            let line = residual
            residual.removeAll(keepingCapacity: true)
            frames.append(try decode(line))
            cursor = chunk.index(after: newline)
        }
        return frames
    }

    /// Called when the feedback stream ends: a partial frame is a violation.
    func finish() throws {
        guard residual.isEmpty else {
            throw ManagedProtocolError.invalidFrame("unterminated frame of \(residual.count) bytes")
        }
    }

    private func decode(_ line: Data) throws -> ManagedFrame {
        let wire: WireEvent
        do {
            wire = try JSONDecoder().decode(WireEvent.self, from: line)
        } catch {
            throw ManagedProtocolError.invalidFrame("not a v1 JSON frame: \(error)")
        }
        guard wire.version == 1 else {
            throw ManagedProtocolError.unsupportedVersion(wire.version)
        }
        guard wire.generation == generation else {
            throw ManagedProtocolError.generationMismatch(expected: generation, received: wire.generation)
        }
        switch wire.event {
        case "ready":
            guard let url = wire.url, !url.isEmpty else {
                throw ManagedProtocolError.invalidFrame("ready frame is missing url")
            }
            guard let reload = wire.reload else {
                throw ManagedProtocolError.invalidFrame("ready frame is missing the nested reload snapshot")
            }
            return .ready(url: url, reload: try reload.snapshot())
        case "reload-status":
            guard let reload = wire.reload else {
                throw ManagedProtocolError.invalidFrame("reload-status frame is missing the nested reload snapshot")
            }
            return .reloadStatus(try reload.snapshot())
        case "target-status":
            guard let target = wire.target else {
                throw ManagedProtocolError.invalidFrame("target-status frame is missing the nested target snapshot")
            }
            guard ManagedTargetSnapshot.states.contains(target.state) else {
                throw ManagedProtocolError.invalidFrame("unknown target state \(target.state)")
            }
            return .targetStatus(ManagedTargetSnapshot(state: target.state, reason: target.reason))
        case "fatal":
            guard let code = wire.code, !code.isEmpty else {
                throw ManagedProtocolError.invalidFrame("fatal frame is missing code")
            }
            guard let message = wire.message, !message.isEmpty else {
                throw ManagedProtocolError.invalidFrame("fatal frame is missing message")
            }
            return .fatal(code: code, message: message)
        default:
            throw ManagedProtocolError.invalidFrame("unknown event \(wire.event)")
        }
    }
}

private struct WireEvent: Decodable {
    let version: Int
    let event: String
    let generation: String
    let url: String?
    let reload: WireReload?
    let target: WireTarget?
    let code: String?
    let message: String?
}

private struct WireReload: Decodable {
    let state: String
    let reason: String?

    func snapshot() throws -> ManagedReloadSnapshot {
        guard ManagedReloadSnapshot.states.contains(state) else {
            throw ManagedProtocolError.invalidFrame("unknown reload state \(state)")
        }
        return ManagedReloadSnapshot(state: state, reason: reason)
    }
}

private struct WireTarget: Decodable {
    let state: String
    let reason: String?
}
