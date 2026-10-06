import Foundation

/// Target identity shared by every opening entry: the resolved path decides
/// which session is reused, the user-selected path stays for display, and the
/// mode decides directory-recursive versus single-file preview.
struct PreviewTarget: Equatable {
    let identity: String
    let displayPath: String
    let mode: ManagedProcess.TargetMode
}

/// Why a selected target cannot become a preview session. Preparation never
/// falls back to the parent directory or guesses a type.
enum TargetPreparationFailure: Error, Equatable {
    case unsupportedFile(String)
    case unavailable(String)
}

/// Background target preparation: standardize the absolute file URL, resolve
/// symlinks, and classify the actual type. File-manager metadata can block, so
/// this is a nonisolated async function that callers await off the main actor;
/// Go remains the final judge of whether the target can actually be served.
enum TargetPreparation {
    static func prepare(
        _ selected: URL,
        fileManager: FileManager = .default
    ) async -> Result<PreviewTarget, TargetPreparationFailure> {
        let absolute = URL(fileURLWithPath: selected.path).standardizedFileURL
        let resolved = absolute.resolvingSymlinksInPath()
        do {
            let attributes = try fileManager.attributesOfItem(atPath: resolved.path)
            let type = attributes[.type] as? FileAttributeType
            let mode: ManagedProcess.TargetMode
            if type == .typeDirectory {
                mode = .directory
            } else if type == .typeRegular, resolved.pathExtension.lowercased() == "md" {
                mode = .file
            } else {
                return .failure(.unsupportedFile(NSLocalizedString(
                    "Only directories and Markdown files can be previewed",
                    comment: "Target classification: unsupported file"
                )))
            }
            return .success(PreviewTarget(identity: resolved.path, displayPath: absolute.path, mode: mode))
        } catch {
            return .failure(.unavailable(error.localizedDescription))
        }
    }
}
