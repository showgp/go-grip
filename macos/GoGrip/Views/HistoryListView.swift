import SwiftUI

struct HistoryListView: View {
    let history: [HistoryEntry]
    let processManager: ProcessManager
    let onDelete: (UUID) -> Void
    let onTap: (HistoryEntry) -> Void

    var body: some View {
        VStack(spacing: 4) {
            HStack {
                Text("📋 最近打开")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                Spacer()
                if !history.isEmpty {
                    Button("清空") {
                        Storage().clearHistory()
                    }
                    .buttonStyle(.plain)
                    .font(.caption)
                    .foregroundColor(.secondary)
                }
            }
            .padding(.horizontal, 12)
            .padding(.top, 4)

            if history.isEmpty {
                VStack(spacing: 8) {
                    Spacer()
                    Text("暂无历史记录")
                        .foregroundColor(.secondary)
                        .font(.caption)
                    Spacer()
                }
                .frame(maxWidth: .infinity, minHeight: 60)
            } else {
                ScrollView {
                    VStack(spacing: 2) {
                        ForEach(history) { entry in
                            HistoryRowView(
                                entry: entry,
                                isRunning: processManager.isRunning(path: entry.path),
                                onTap: { onTap(entry) },
                                onStop: { processManager.stop(path: entry.path) },
                                onDelete: { onDelete(entry.id) }
                            )
                        }
                    }
                    .padding(.vertical, 2)
                }
            }
        }
    }
}
