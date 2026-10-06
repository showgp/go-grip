import Darwin
import Foundation
import XCTest

/// Identity seam: real directories, case-insensitive Markdown files, symlinks,
/// unsupported files and missing targets are prepared against the real file
/// system. The resolved path decides the identity, the selected path stays for
/// display, and a rejected target never becomes its parent directory.
final class TargetPreparationTests: XCTestCase {
    private var tempRoot: URL!

    override func setUpWithError() throws {
        tempRoot = FileManager.default.temporaryDirectory
            .appendingPathComponent("gogrip-target-prep-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: tempRoot, withIntermediateDirectories: true)
    }

    override func tearDownWithError() throws {
        if let tempRoot {
            try? FileManager.default.removeItem(at: tempRoot)
        }
    }

    func testDirectoryAndUppercaseMarkdownExtensionUseTheirActualModes() async throws {
        let directory = tempRoot.appendingPathComponent("docs")
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        let uppercase = tempRoot.appendingPathComponent("docs/说明.MD")
        try Data("# 说明".utf8).write(to: uppercase)

        guard let directoryTarget = await prepared(directory),
              let fileTarget = await prepared(uppercase)
        else { return }

        XCTAssertEqual(directoryTarget.mode, .directory)
        XCTAssertEqual(fileTarget.mode, .file)
    }

    func testSymlinkResolvesToTheSameIdentityWhileDisplayKeepsTheSelectedPath() async throws {
        let real = tempRoot.appendingPathComponent("real")
        try FileManager.default.createDirectory(at: real, withIntermediateDirectories: true)
        let alias = tempRoot.appendingPathComponent("alias")
        try FileManager.default.createSymbolicLink(at: alias, withDestinationURL: real)

        guard let realTarget = await prepared(real), let aliasTarget = await prepared(alias) else { return }

        XCTAssertEqual(realTarget.identity, aliasTarget.identity)
        XCTAssertEqual(realTarget.displayPath, real.standardizedFileURL.path)
        XCTAssertEqual(aliasTarget.displayPath, alias.standardizedFileURL.path)
        XCTAssertNotEqual(realTarget.displayPath, aliasTarget.displayPath)
    }

    func testUnsupportedFileIsRejectedInsteadOfBecomingItsParentDirectory() async throws {
        let image = tempRoot.appendingPathComponent("photo.png")
        try Data("png".utf8).write(to: image)

        switch await TargetPreparation.prepare(image) {
        case .success:
            XCTFail("A non-Markdown file must not become a preview target")
        case .failure(.unsupportedFile(_)):
            break
        case .failure(.unavailable(_)):
            XCTFail("An existing unsupported file is not an unavailable target")
        }
    }

    func testMarkdownNamedNonRegularTargetIsRejected() async throws {
        let fifo = tempRoot.appendingPathComponent("pipe.md")
        XCTAssertEqual(mkfifo(fifo.path, 0o600), 0, "could not create the FIFO fixture")

        switch await TargetPreparation.prepare(fifo) {
        case .success:
            XCTFail("A non-regular Markdown-named target must not open")
        case .failure(.unsupportedFile(_)):
            break
        case .failure(.unavailable(_)):
            XCTFail("An existing non-regular target is not an unavailable target")
        }
    }

    func testMissingTargetReportsUnavailable() async throws {
        let missing = tempRoot.appendingPathComponent("missing")

        switch await TargetPreparation.prepare(missing) {
        case .success:
            XCTFail("A missing target must not become a preview target")
        case .failure(.unavailable(_)):
            break
        case .failure(.unsupportedFile(_)):
            XCTFail("A missing path is not an unsupported file")
        }
    }

    private func prepared(_ url: URL, line: UInt = #line) async -> PreviewTarget? {
        switch await TargetPreparation.prepare(url) {
        case .success(let target):
            return target
        case .failure(let failure):
            XCTFail("unexpected preparation failure: \(failure)", line: line)
            return nil
        }
    }
}
