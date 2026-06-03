import Foundation

struct RunningInstance {
    let path: String
    let process: Process
    let port: Int
    var pid: Int32 { process.processIdentifier }
}

class ProcessManager: ObservableObject, @unchecked Sendable {
    @Published var instances: [String: RunningInstance] = [:]
    @Published var lastError: String? = nil
    private var startingPaths: Set<String> = []

    var count: Int { instances.count }

#if DEBUG
    var _testBinaryURL: URL? = nil
#endif

    func start(path: String) async {
        guard await MainActor.run(body: {
            guard !startingPaths.contains(path) else { return false }
            startingPaths.insert(path)
            return true
        }) else { return }
        defer {
            Task { await MainActor.run { startingPaths.remove(path) } }
        }

        let url: URL? = {
            #if DEBUG
            if let testURL = _testBinaryURL { return testURL }
            #endif
            return Bundle.main.url(forResource: "go-grip", withExtension: nil)
        }()
        guard let url else {
            await MainActor.run {
                lastError = "未找到 go-grip 二进制文件"
                instances[path] = nil
            }
            return
        }
        try? FileManager.default.setAttributes([.posixPermissions: 0o755], ofItemAtPath: url.path)

        let process = Process()
        process.executableURL = url

        var isDir: ObjCBool = false
        FileManager.default.fileExists(atPath: path, isDirectory: &isDir)
        var args = ["--json", "--browser=false"]
        if isDir.boolValue { args.append("-r") }
        args.append(path)
        process.arguments = args

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = FileHandle.nullDevice

        process.terminationHandler = { [weak self] _ in
            DispatchQueue.main.async {
                self?.instances.removeValue(forKey: path)
            }
        }

        do {
            try process.run()
        } catch {
            await MainActor.run {
                lastError = "启动 go-grip 失败: \(error.localizedDescription)"
                instances[path] = nil
            }
            return
        }

        let port = await readPortFromStdout(pipe: pipe)
        let instance = RunningInstance(path: path, process: process, port: port)
        await MainActor.run { instances[path] = instance }
    }

    private func readPortFromStdout(pipe: Pipe) async -> Int {
        await Self.readPortFromPipe(pipe, timeout: 5)
    }

    static func readPortFromPipe(_ pipe: Pipe, timeout: TimeInterval = 5) async -> Int {
        await withCheckedContinuation { continuation in
            var resumed = false
            let lock = NSLock()

            DispatchQueue.global().asyncAfter(deadline: .now() + timeout) {
                lock.lock()
                if !resumed {
                    resumed = true
                    lock.unlock()
                    continuation.resume(returning: 6419)
                } else {
                    lock.unlock()
                }
            }

            DispatchQueue.global().async {
                let handle = pipe.fileHandleForReading
                var buffer = Data()

                while true {
                    let chunk = handle.availableData
                    if chunk.isEmpty { break }
                    buffer.append(chunk)

                    while let newlineRange = buffer.range(of: Data("\n".utf8)) {
                        let lineData = buffer.subdata(in: buffer.startIndex..<newlineRange.lowerBound)
                        buffer.removeSubrange(buffer.startIndex..<newlineRange.upperBound)

                        if let json = try? JSONSerialization.jsonObject(with: lineData) as? [String: Any],
                           let port = json["port"] as? Int {
                            lock.lock()
                            if !resumed {
                                resumed = true
                                lock.unlock()
                                continuation.resume(returning: port)
                                return
                            } else {
                                lock.unlock()
                                return
                            }
                        }
                    }
                }

                lock.lock()
                if !resumed {
                    resumed = true
                    lock.unlock()
                    continuation.resume(returning: 6419)
                } else {
                    lock.unlock()
                }
            }
        }
    }

    func stop(path: String) {
        guard let instance = instances[path] else { return }
        instance.process.terminate()
    }

    func stopAll() {
        let allInstances = instances
        for (_, instance) in allInstances {
            instance.process.terminate()
        }

        DispatchQueue.main.asyncAfter(deadline: .now() + 2) { [weak self] in
            guard let self else { return }
            for (_, instance) in self.instances {
                if instance.process.isRunning {
                    instance.process.interrupt()
                }
            }
            self.instances.removeAll()
        }
    }

    func isRunning(path: String) -> Bool {
        instances[path]?.process.isRunning ?? false
    }

    func port(for path: String) -> Int? {
        instances[path]?.port
    }
}
