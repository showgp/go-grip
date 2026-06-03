import SwiftUI
import UniformTypeIdentifiers

struct SizePreferenceKey: PreferenceKey {
    static var defaultValue: CGSize = .zero
    static func reduce(value: inout CGSize, nextValue: () -> CGSize) {}
}

struct PopoverView: View {
    @ObservedObject var processManager: ProcessManager
    @State private var history: [HistoryEntry] = []
    @State private var isDragOver = false
    @State private var isValidDrop = true
    @State private var showSettings = false
    @State private var binaryNotFound = false
    @State private var showStartError = false
    @State private var startErrorMessage = ""

    var onDragStateChange: ((Bool) -> Void)?
    var onSizeChange: ((CGSize) -> Void)?
    let storage = Storage()

    var body: some View {
        ZStack {
            VStack(spacing: 0) {
                titleBar
                Divider()
                contentArea
                Divider()
                openButton
            }

            if isDragOver {
                DropOverlayView(isValidDrop: isValidDrop)
                    .transition(.opacity)
                    .animation(.easeInOut(duration: 0.2), value: isDragOver)
            }
        }
        .onDrop(of: [.fileURL], isTargeted: $isDragOver) { providers in
            return handleDrop(providers)
        }
        .onChange(of: isDragOver) { newValue in
            onDragStateChange?(newValue)
            if newValue {
                isValidDrop = true
            }
        }
        .sheet(isPresented: $showSettings) {
            SettingsView()
        }
        .alert("未找到 go-grip 二进制文件", isPresented: $binaryNotFound) {
            Button("确定", role: .cancel) {}
        } message: {
            Text("请确保 go-grip 已放置在应用程序的 Resources 目录中。")
        }
        .onAppear {
            history = storage.load()
        }
        .alert("启动失败", isPresented: $showStartError) {
            Button("确定", role: .cancel) {}
        } message: {
            Text(startErrorMessage)
        }
        .onReceive(processManager.objectWillChange) { _ in
            if let error = processManager.lastError {
                startErrorMessage = error
                showStartError = true
                processManager.lastError = nil
            }
        }
        .background(GeometryReader { geo in
            Color.clear.preference(key: SizePreferenceKey.self, value: geo.size)
        })
        .onPreferenceChange(SizePreferenceKey.self) { size in
            onSizeChange?(size)
        }
        .frame(minWidth: 400, maxWidth: 400, minHeight: 200, maxHeight: 800)
    }

    private var titleBar: some View {
        HStack {
            Text("GoGrip").font(.headline)
            Spacer()
            Button(action: { showSettings.toggle() }) {
                Image(systemName: "gearshape")
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
    }

    @ViewBuilder
    private var contentArea: some View {
        if history.isEmpty {
            emptyStateView
        } else {
            HistoryListView(
                history: history,
                processManager: processManager,
                onDelete: { id in
                    if let entry = history.first(where: { $0.id == id }),
                       processManager.isRunning(path: entry.path) {
                        processManager.stop(path: entry.path)
                    }
                    storage.removeFromHistory(id: id)
                    history = storage.load()
                },
                onTap: { entry in
                    Task { await openPathAsync(entry.path) }
                }
            )
        }
    }

    private var emptyStateView: some View {
        VStack(spacing: 16) {
            Spacer()
            Image(systemName: "folder")
                .font(.system(size: 48))
                .foregroundColor(.secondary)
            Text("拖拽目录或 .md 文件到此处")
                .font(.body)
                .foregroundColor(.secondary)
            Text("或点击下方按钮打开")
                .font(.caption)
                .foregroundColor(.secondary)
            Spacer()
        }
    }

    private var openButton: some View {
        Button(action: showOpenPanel) {
            Label("打开", systemImage: "folder")
                .frame(maxWidth: .infinity)
        }
        .buttonStyle(.bordered)
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
    }

    private func handleDrop(_ providers: [NSItemProvider]) -> Bool {
        guard let provider = providers.first else { return false }
        provider.loadObject(ofClass: NSURL.self) { url, _ in
            guard let url = url as? URL else { return }
            let path = url.path
            var isDir: ObjCBool = false
            let exists = FileManager.default.fileExists(atPath: path, isDirectory: &isDir)
            let isMarkdown = path.hasSuffix(".md")
            guard exists, isDir.boolValue || isMarkdown else {
                DispatchQueue.main.async {
                    isValidDrop = false
                }
                return
            }
            DispatchQueue.main.async {
                Task { await self.openPathAsync(path) }
            }
        }
        return true
    }

    private func showOpenPanel() {
        let panel = NSOpenPanel()
        panel.allowsMultipleSelection = false
        panel.canChooseDirectories = true
        panel.canChooseFiles = true
        panel.allowedContentTypes = [.folder, UTType(filenameExtension: "md")].compactMap { $0 }
        panel.begin { response in
            if response == .OK, let url = panel.url {
                Task { await self.openPathAsync(url.path) }
            }
        }
    }

    private func openPathAsync(_ path: String) async {
        var isDir: ObjCBool = false
        let exists = FileManager.default.fileExists(atPath: path, isDirectory: &isDir)
        let isMarkdown = path.hasSuffix(".md")
        guard exists, isDir.boolValue || isMarkdown else { return }

        if processManager.isRunning(path: path) {
            if let port = processManager.port(for: path) {
                guard let url = URL(string: "http://localhost:\(port)") else { return }
                await MainActor.run { NSWorkspace.shared.open(url) }
            }
        } else {
            await processManager.start(path: path)
            if let port = processManager.port(for: path) {
                guard let url = URL(string: "http://localhost:\(port)") else { return }
                await MainActor.run { NSWorkspace.shared.open(url) }
            }
        }
        await MainActor.run {
            storage.addToHistory(path: path)
            history = storage.load()
        }
    }
}
